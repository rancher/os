package handlers

import (
	cryptorand "crypto/rand"
	"expvar"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/distribution"
	"github.com/docker/distribution/configuration"
	ctxu "github.com/docker/distribution/context"
	"github.com/docker/distribution/notifications"
	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/auth"
	registrymiddleware "github.com/docker/distribution/registry/middleware/registry"
	repositorymiddleware "github.com/docker/distribution/registry/middleware/repository"
	"github.com/docker/distribution/registry/proxy"
	"github.com/docker/distribution/registry/storage"
	memorycache "github.com/docker/distribution/registry/storage/cache/memory"
	rediscache "github.com/docker/distribution/registry/storage/cache/redis"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/factory"
	storagemiddleware "github.com/docker/distribution/registry/storage/driver/middleware"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

// randomSecretSize is the number of random bytes to generate if no secret
// was specified.
const randomSecretSize = 32

// App is a global registry application object. Shared resources can be placed
// on this object that will be accessible from all requests. Any writable
// fields should be protected.
type App struct {
	context.Context

	Config configuration.Configuration

	router           *mux.Router                 // main application router, configured with dispatchers
	driver           storagedriver.StorageDriver // driver maintains the app global storage driver instance.
	registry         distribution.Namespace      // registry is the primary registry backend for the app instance.
	accessController auth.AccessController       // main access controller for application

	// events contains notification related configuration.
	events struct {
		sink   notifications.Sink
		source notifications.SourceRecord
	}

	redis *redis.Pool

	// true if this registry is configured as a pull through cache
	isCache bool
}

// NewApp takes a configuration and returns a configured app, ready to serve
// requests. The app only implements ServeHTTP and can be wrapped in other
// handlers accordingly.
func NewApp(ctx context.Context, configuration configuration.Configuration) *App {
	app := &App{
		Config:  configuration,
		Context: ctx,
		router:  v2.RouterWithPrefix(configuration.HTTP.Prefix),
		isCache: configuration.Proxy.RemoteURL != "",
	}

	app.Context = ctxu.WithLogger(app.Context, ctxu.GetLogger(app, "instance.id"))

	// Register the handler dispatchers.
	app.register(v2.RouteNameBase, func(ctx *Context, r *http.Request) http.Handler {
		return http.HandlerFunc(apiBase)
	})
	app.register(v2.RouteNameManifest, imageManifestDispatcher)
	app.register(v2.RouteNameCatalog, catalogDispatcher)
	app.register(v2.RouteNameTags, tagsDispatcher)
	app.register(v2.RouteNameBlob, blobDispatcher)
	app.register(v2.RouteNameBlobUpload, blobUploadDispatcher)
	app.register(v2.RouteNameBlobUploadChunk, blobUploadDispatcher)

	var err error
	app.driver, err = factory.Create(configuration.Storage.Type(), configuration.Storage.Parameters())
	if err != nil {
		// TODO(stevvooe): Move the creation of a service into a protected
		// method, where this is created lazily. Its status can be queried via
		// a health check.
		panic(err)
	}

	purgeConfig := uploadPurgeDefaultConfig()
	if mc, ok := configuration.Storage["maintenance"]; ok {
		for k, v := range mc {
			switch k {
			case "uploadpurging":
				purgeConfig = v.(map[interface{}]interface{})
			}
		}

	}

	startUploadPurger(app, app.driver, ctxu.GetLogger(app), purgeConfig)

	app.driver, err = applyStorageMiddleware(app.driver, configuration.Middleware["storage"])
	if err != nil {
		panic(err)
	}

	app.configureSecret(&configuration)
	app.configureEvents(&configuration)
	app.configureRedis(&configuration)
	app.configureLogHook(&configuration)

	// configure deletion
	var deleteEnabled bool
	if d, ok := configuration.Storage["delete"]; ok {
		e, ok := d["enabled"]
		if ok {
			if deleteEnabled, ok = e.(bool); !ok {
				deleteEnabled = false
			}
		}
	}

	// configure redirects
	var redirectDisabled bool
	if redirectConfig, ok := configuration.Storage["redirect"]; ok {
		v := redirectConfig["disable"]
		switch v := v.(type) {
		case bool:
			redirectDisabled = v
		default:
			panic(fmt.Sprintf("invalid type for redirect config: %#v", redirectConfig))
		}

		if redirectDisabled {
			ctxu.GetLogger(app).Infof("backend redirection disabled")
		}
	}

	// configure storage caches
	if cc, ok := configuration.Storage["cache"]; ok {
		v, ok := cc["blobdescriptor"]
		if !ok {
			// Backwards compatible: "layerinfo" == "blobdescriptor"
			v = cc["layerinfo"]
		}

		switch v {
		case "redis":
			if app.redis == nil {
				panic("redis configuration required to use for layerinfo cache")
			}
			app.registry = storage.NewRegistryWithDriver(app, app.driver, rediscache.NewRedisBlobDescriptorCacheProvider(app.redis), deleteEnabled, !redirectDisabled, app.isCache)
			ctxu.GetLogger(app).Infof("using redis blob descriptor cache")
		case "inmemory":
			app.registry = storage.NewRegistryWithDriver(app, app.driver, memorycache.NewInMemoryBlobDescriptorCacheProvider(), deleteEnabled, !redirectDisabled, app.isCache)
			ctxu.GetLogger(app).Infof("using inmemory blob descriptor cache")
		default:
			if v != "" {
				ctxu.GetLogger(app).Warnf("unknown cache type %q, caching disabled", configuration.Storage["cache"])
			}
		}
	}

	if app.registry == nil {
		// configure the registry if no cache section is available.
		app.registry = storage.NewRegistryWithDriver(app.Context, app.driver, nil, deleteEnabled, !redirectDisabled, app.isCache)
	}

	app.registry, err = applyRegistryMiddleware(app.Context, app.registry, configuration.Middleware["registry"])
	if err != nil {
		panic(err)
	}

	authType := configuration.Auth.Type()

	if authType != "" {
		accessController, err := auth.GetAccessController(configuration.Auth.Type(), configuration.Auth.Parameters())
		if err != nil {
			panic(fmt.Sprintf("unable to configure authorization (%s): %v", authType, err))
		}
		app.accessController = accessController
		ctxu.GetLogger(app).Debugf("configured %q access controller", authType)
	}

	// configure as a pull through cache
	if configuration.Proxy.RemoteURL != "" {
		app.registry, err = proxy.NewRegistryPullThroughCache(ctx, app.registry, app.driver, configuration.Proxy)
		if err != nil {
			panic(err.Error())
		}
		app.isCache = true
		ctxu.GetLogger(app).Info("Registry configured as a proxy cache to ", configuration.Proxy.RemoteURL)
	}

	return app
}

// register a handler with the application, by route name. The handler will be
// passed through the application filters and context will be constructed at
// request time.
func (app *App) register(routeName string, dispatch dispatchFunc) {

	// TODO(stevvooe): This odd dispatcher/route registration is by-product of
	// some limitations in the gorilla/mux router. We are using it to keep
	// routing consistent between the client and server, but we may want to
	// replace it with manual routing and structure-based dispatch for better
	// control over the request execution.

	app.router.GetRoute(routeName).Handler(app.dispatcher(dispatch))
}

// configureEvents prepares the event sink for action.
func (app *App) configureEvents(configuration *configuration.Configuration) {
	// Configure all of the endpoint sinks.
	var sinks []notifications.Sink
	for _, endpoint := range configuration.Notifications.Endpoints {
		if endpoint.Disabled {
			ctxu.GetLogger(app).Infof("endpoint %s disabled, skipping", endpoint.Name)
			continue
		}

		ctxu.GetLogger(app).Infof("configuring endpoint %v (%v), timeout=%s, headers=%v", endpoint.Name, endpoint.URL, endpoint.Timeout, endpoint.Headers)
		endpoint := notifications.NewEndpoint(endpoint.Name, endpoint.URL, notifications.EndpointConfig{
			Timeout:   endpoint.Timeout,
			Threshold: endpoint.Threshold,
			Backoff:   endpoint.Backoff,
			Headers:   endpoint.Headers,
		})

		sinks = append(sinks, endpoint)
	}

	// NOTE(stevvooe): Moving to a new queueing implementation is as easy as
	// replacing broadcaster with a rabbitmq implementation. It's recommended
	// that the registry instances also act as the workers to keep deployment
	// simple.
	app.events.sink = notifications.NewBroadcaster(sinks...)

	// Populate registry event source
	hostname, err := os.Hostname()
	if err != nil {
		hostname = configuration.HTTP.Addr
	} else {
		// try to pick the port off the config
		_, port, err := net.SplitHostPort(configuration.HTTP.Addr)
		if err == nil {
			hostname = net.JoinHostPort(hostname, port)
		}
	}

	app.events.source = notifications.SourceRecord{
		Addr:       hostname,
		InstanceID: ctxu.GetStringValue(app, "instance.id"),
	}
}

func (app *App) configureRedis(configuration *configuration.Configuration) {
	if configuration.Redis.Addr == "" {
		ctxu.GetLogger(app).Infof("redis not configured")
		return
	}

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			// TODO(stevvooe): Yet another use case for contextual timing.
			ctx := context.WithValue(app, "redis.connect.startedat", time.Now())

			done := func(err error) {
				logger := ctxu.GetLoggerWithField(ctx, "redis.connect.duration",
					ctxu.Since(ctx, "redis.connect.startedat"))
				if err != nil {
					logger.Errorf("redis: error connecting: %v", err)
				} else {
					logger.Infof("redis: connect %v", configuration.Redis.Addr)
				}
			}

			conn, err := redis.DialTimeout("tcp",
				configuration.Redis.Addr,
				configuration.Redis.DialTimeout,
				configuration.Redis.ReadTimeout,
				configuration.Redis.WriteTimeout)
			if err != nil {
				ctxu.GetLogger(app).Errorf("error connecting to redis instance %s: %v",
					configuration.Redis.Addr, err)
				done(err)
				return nil, err
			}

			// authorize the connection
			if configuration.Redis.Password != "" {
				if _, err = conn.Do("AUTH", configuration.Redis.Password); err != nil {
					defer conn.Close()
					done(err)
					return nil, err
				}
			}

			// select the database to use
			if configuration.Redis.DB != 0 {
				if _, err = conn.Do("SELECT", configuration.Redis.DB); err != nil {
					defer conn.Close()
					done(err)
					return nil, err
				}
			}

			done(nil)
			return conn, nil
		},
		MaxIdle:     configuration.Redis.Pool.MaxIdle,
		MaxActive:   configuration.Redis.Pool.MaxActive,
		IdleTimeout: configuration.Redis.Pool.IdleTimeout,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			// TODO(stevvooe): We can probably do something more interesting
			// here with the health package.
			_, err := c.Do("PING")
			return err
		},
		Wait: false, // if a connection is not avialable, proceed without cache.
	}

	app.redis = pool

	// setup expvar
	registry := expvar.Get("registry")
	if registry == nil {
		registry = expvar.NewMap("registry")
	}

	registry.(*expvar.Map).Set("redis", expvar.Func(func() interface{} {
		return map[string]interface{}{
			"Config": configuration.Redis,
			"Active": app.redis.ActiveCount(),
		}
	}))
}

// configureLogHook prepares logging hook parameters.
func (app *App) configureLogHook(configuration *configuration.Configuration) {
	entry, ok := ctxu.GetLogger(app).(*log.Entry)
	if !ok {
		// somehow, we are not using logrus
		return
	}

	logger := entry.Logger

	for _, configHook := range configuration.Log.Hooks {
		if !configHook.Disabled {
			switch configHook.Type {
			case "mail":
				hook := &logHook{}
				hook.LevelsParam = configHook.Levels
				hook.Mail = &mailer{
					Addr:     configHook.MailOptions.SMTP.Addr,
					Username: configHook.MailOptions.SMTP.Username,
					Password: configHook.MailOptions.SMTP.Password,
					Insecure: configHook.MailOptions.SMTP.Insecure,
					From:     configHook.MailOptions.From,
					To:       configHook.MailOptions.To,
				}
				logger.Hooks.Add(hook)
			default:
			}
		}
	}
}

// configureSecret creates a random secret if a secret wasn't included in the
// configuration.
func (app *App) configureSecret(configuration *configuration.Configuration) {
	if configuration.HTTP.Secret == "" {
		var secretBytes [randomSecretSize]byte
		if _, err := cryptorand.Read(secretBytes[:]); err != nil {
			panic(fmt.Sprintf("could not generate random bytes for HTTP secret: %v", err))
		}
		configuration.HTTP.Secret = string(secretBytes[:])
		ctxu.GetLogger(app).Warn("No HTTP secret provided - generated random secret. This may cause problems with uploads if multiple registries are behind a load-balancer. To provide a shared secret, fill in http.secret in the configuration file or set the REGISTRY_HTTP_SECRET environment variable.")
	}
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // ensure that request body is always closed.

	// Instantiate an http context here so we can track the error codes
	// returned by the request router.
	ctx := defaultContextManager.context(app, w, r)

	defer func() {
		status, ok := ctx.Value("http.response.status").(int)
		if ok && status >= 200 && status <= 399 {
			ctxu.GetResponseLogger(ctx).Infof("response completed")
		}
	}()
	defer defaultContextManager.release(ctx)

	// NOTE(stevvooe): Total hack to get instrumented responsewriter from context.
	var err error
	w, err = ctxu.GetResponseWriter(ctx)
	if err != nil {
		ctxu.GetLogger(ctx).Warnf("response writer not found in context")
	}

	// Set a header with the Docker Distribution API Version for all responses.
	w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
	app.router.ServeHTTP(w, r)
}

// dispatchFunc takes a context and request and returns a constructed handler
// for the route. The dispatcher will use this to dynamically create request
// specific handlers for each endpoint without creating a new router for each
// request.
type dispatchFunc func(ctx *Context, r *http.Request) http.Handler

// TODO(stevvooe): dispatchers should probably have some validation error
// chain with proper error reporting.

// dispatcher returns a handler that constructs a request specific context and
// handler, using the dispatch factory function.
func (app *App) dispatcher(dispatch dispatchFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context := app.context(w, r)

		if err := app.authorized(w, r, context); err != nil {
			ctxu.GetLogger(context).Warnf("error authorizing context: %v", err)
			return
		}

		// Add username to request logging
		context.Context = ctxu.WithLogger(context.Context, ctxu.GetLogger(context.Context, "auth.user.name"))

		if app.nameRequired(r) {
			repository, err := app.registry.Repository(context, getName(context))

			if err != nil {
				ctxu.GetLogger(context).Errorf("error resolving repository: %v", err)

				switch err := err.(type) {
				case distribution.ErrRepositoryUnknown:
					context.Errors = append(context.Errors, v2.ErrorCodeNameUnknown.WithDetail(err))
				case distribution.ErrRepositoryNameInvalid:
					context.Errors = append(context.Errors, v2.ErrorCodeNameInvalid.WithDetail(err))
				}

				if err := errcode.ServeJSON(w, context.Errors); err != nil {
					ctxu.GetLogger(context).Errorf("error serving error json: %v (from %v)", err, context.Errors)
				}
				return
			}

			// assign and decorate the authorized repository with an event bridge.
			context.Repository = notifications.Listen(
				repository,
				app.eventBridge(context, r))

			context.Repository, err = applyRepoMiddleware(context.Context, context.Repository, app.Config.Middleware["repository"])
			if err != nil {
				ctxu.GetLogger(context).Errorf("error initializing repository middleware: %v", err)
				context.Errors = append(context.Errors, errcode.ErrorCodeUnknown.WithDetail(err))

				if err := errcode.ServeJSON(w, context.Errors); err != nil {
					ctxu.GetLogger(context).Errorf("error serving error json: %v (from %v)", err, context.Errors)
				}
				return
			}
		}

		dispatch(context, r).ServeHTTP(w, r)
		// Automated error response handling here. Handlers may return their
		// own errors if they need different behavior (such as range errors
		// for layer upload).
		if context.Errors.Len() > 0 {
			if err := errcode.ServeJSON(w, context.Errors); err != nil {
				ctxu.GetLogger(context).Errorf("error serving error json: %v (from %v)", err, context.Errors)
			}

			app.logError(context, context.Errors)
		}
	})
}

func (app *App) logError(context context.Context, errors errcode.Errors) {
	for _, e1 := range errors {
		var c ctxu.Context

		switch e1.(type) {
		case errcode.Error:
			e, _ := e1.(errcode.Error)
			c = ctxu.WithValue(context, "err.code", e.Code)
			c = ctxu.WithValue(c, "err.message", e.Code.Message())
			c = ctxu.WithValue(c, "err.detail", e.Detail)
		case errcode.ErrorCode:
			e, _ := e1.(errcode.ErrorCode)
			c = ctxu.WithValue(context, "err.code", e)
			c = ctxu.WithValue(c, "err.message", e.Message())
		default:
			// just normal go 'error'
			c = ctxu.WithValue(context, "err.code", errcode.ErrorCodeUnknown)
			c = ctxu.WithValue(c, "err.message", e1.Error())
		}

		c = ctxu.WithLogger(c, ctxu.GetLogger(c,
			"err.code",
			"err.message",
			"err.detail"))
		ctxu.GetResponseLogger(c).Errorf("response completed with error")
	}
}

// context constructs the context object for the application. This only be
// called once per request.
func (app *App) context(w http.ResponseWriter, r *http.Request) *Context {
	ctx := defaultContextManager.context(app, w, r)
	ctx = ctxu.WithVars(ctx, r)
	ctx = ctxu.WithLogger(ctx, ctxu.GetLogger(ctx,
		"vars.name",
		"vars.reference",
		"vars.digest",
		"vars.uuid"))

	context := &Context{
		App:        app,
		Context:    ctx,
		urlBuilder: v2.NewURLBuilderFromRequest(r),
	}

	return context
}

// authorized checks if the request can proceed with access to the requested
// repository. If it succeeds, the context may access the requested
// repository. An error will be returned if access is not available.
func (app *App) authorized(w http.ResponseWriter, r *http.Request, context *Context) error {
	ctxu.GetLogger(context).Debug("authorizing request")
	repo := getName(context)

	if app.accessController == nil {
		return nil // access controller is not enabled.
	}

	var accessRecords []auth.Access

	if repo != "" {
		accessRecords = appendAccessRecords(accessRecords, r.Method, repo)
	} else {
		// Only allow the name not to be set on the base route.
		if app.nameRequired(r) {
			// For this to be properly secured, repo must always be set for a
			// resource that may make a modification. The only condition under
			// which name is not set and we still allow access is when the
			// base route is accessed. This section prevents us from making
			// that mistake elsewhere in the code, allowing any operation to
			// proceed.
			if err := errcode.ServeJSON(w, v2.ErrorCodeUnauthorized); err != nil {
				ctxu.GetLogger(context).Errorf("error serving error json: %v (from %v)", err, context.Errors)
			}
			return fmt.Errorf("forbidden: no repository name")
		}
		accessRecords = appendCatalogAccessRecord(accessRecords, r)
	}

	ctx, err := app.accessController.Authorized(context.Context, accessRecords...)
	if err != nil {
		switch err := err.(type) {
		case auth.Challenge:
			// Add the appropriate WWW-Auth header
			err.SetHeaders(w)

			if err := errcode.ServeJSON(w, v2.ErrorCodeUnauthorized.WithDetail(accessRecords)); err != nil {
				ctxu.GetLogger(context).Errorf("error serving error json: %v (from %v)", err, context.Errors)
			}
		default:
			// This condition is a potential security problem either in
			// the configuration or whatever is backing the access
			// controller. Just return a bad request with no information
			// to avoid exposure. The request should not proceed.
			ctxu.GetLogger(context).Errorf("error checking authorization: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}

		return err
	}

	// TODO(stevvooe): This pattern needs to be cleaned up a bit. One context
	// should be replaced by another, rather than replacing the context on a
	// mutable object.
	context.Context = ctx
	return nil
}

// eventBridge returns a bridge for the current request, configured with the
// correct actor and source.
func (app *App) eventBridge(ctx *Context, r *http.Request) notifications.Listener {
	actor := notifications.ActorRecord{
		Name: getUserName(ctx, r),
	}
	request := notifications.NewRequestRecord(ctxu.GetRequestID(ctx), r)

	return notifications.NewBridge(ctx.urlBuilder, app.events.source, actor, request, app.events.sink)
}

// nameRequired returns true if the route requires a name.
func (app *App) nameRequired(r *http.Request) bool {
	route := mux.CurrentRoute(r)
	routeName := route.GetName()
	return route == nil || (routeName != v2.RouteNameBase && routeName != v2.RouteNameCatalog)
}

// apiBase implements a simple yes-man for doing overall checks against the
// api. This can support auth roundtrips to support docker login.
func apiBase(w http.ResponseWriter, r *http.Request) {
	const emptyJSON = "{}"
	// Provide a simple /v2/ 200 OK response with empty json response.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprint(len(emptyJSON)))

	fmt.Fprint(w, emptyJSON)
}

// appendAccessRecords checks the method and adds the appropriate Access records to the records list.
func appendAccessRecords(records []auth.Access, method string, repo string) []auth.Access {
	resource := auth.Resource{
		Type: "repository",
		Name: repo,
	}

	switch method {
	case "GET", "HEAD":
		records = append(records,
			auth.Access{
				Resource: resource,
				Action:   "pull",
			})
	case "POST", "PUT", "PATCH":
		records = append(records,
			auth.Access{
				Resource: resource,
				Action:   "pull",
			},
			auth.Access{
				Resource: resource,
				Action:   "push",
			})
	case "DELETE":
		// DELETE access requires full admin rights, which is represented
		// as "*". This may not be ideal.
		records = append(records,
			auth.Access{
				Resource: resource,
				Action:   "*",
			})
	}
	return records
}

// Add the access record for the catalog if it's our current route
func appendCatalogAccessRecord(accessRecords []auth.Access, r *http.Request) []auth.Access {
	route := mux.CurrentRoute(r)
	routeName := route.GetName()

	if routeName == v2.RouteNameCatalog {
		resource := auth.Resource{
			Type: "registry",
			Name: "catalog",
		}

		accessRecords = append(accessRecords,
			auth.Access{
				Resource: resource,
				Action:   "*",
			})
	}
	return accessRecords
}

// applyRegistryMiddleware wraps a registry instance with the configured middlewares
func applyRegistryMiddleware(ctx context.Context, registry distribution.Namespace, middlewares []configuration.Middleware) (distribution.Namespace, error) {
	for _, mw := range middlewares {
		rmw, err := registrymiddleware.Get(ctx, mw.Name, mw.Options, registry)
		if err != nil {
			return nil, fmt.Errorf("unable to configure registry middleware (%s): %s", mw.Name, err)
		}
		registry = rmw
	}
	return registry, nil

}

// applyRepoMiddleware wraps a repository with the configured middlewares
func applyRepoMiddleware(ctx context.Context, repository distribution.Repository, middlewares []configuration.Middleware) (distribution.Repository, error) {
	for _, mw := range middlewares {
		rmw, err := repositorymiddleware.Get(ctx, mw.Name, mw.Options, repository)
		if err != nil {
			return nil, err
		}
		repository = rmw
	}
	return repository, nil
}

// applyStorageMiddleware wraps a storage driver with the configured middlewares
func applyStorageMiddleware(driver storagedriver.StorageDriver, middlewares []configuration.Middleware) (storagedriver.StorageDriver, error) {
	for _, mw := range middlewares {
		smw, err := storagemiddleware.Get(mw.Name, mw.Options, driver)
		if err != nil {
			return nil, fmt.Errorf("unable to configure storage middleware (%s): %v", mw.Name, err)
		}
		driver = smw
	}
	return driver, nil
}

// uploadPurgeDefaultConfig provides a default configuration for upload
// purging to be used in the absence of configuration in the
// confifuration file
func uploadPurgeDefaultConfig() map[interface{}]interface{} {
	config := map[interface{}]interface{}{}
	config["enabled"] = true
	config["age"] = "168h"
	config["interval"] = "24h"
	config["dryrun"] = false
	return config
}

func badPurgeUploadConfig(reason string) {
	panic(fmt.Sprintf("Unable to parse upload purge configuration: %s", reason))
}

// startUploadPurger schedules a goroutine which will periodically
// check upload directories for old files and delete them
func startUploadPurger(ctx context.Context, storageDriver storagedriver.StorageDriver, log ctxu.Logger, config map[interface{}]interface{}) {
	if config["enabled"] == false {
		return
	}

	var purgeAgeDuration time.Duration
	var err error
	purgeAge, ok := config["age"]
	if ok {
		ageStr, ok := purgeAge.(string)
		if !ok {
			badPurgeUploadConfig("age is not a string")
		}
		purgeAgeDuration, err = time.ParseDuration(ageStr)
		if err != nil {
			badPurgeUploadConfig(fmt.Sprintf("Cannot parse duration: %s", err.Error()))
		}
	} else {
		badPurgeUploadConfig("age missing")
	}

	var intervalDuration time.Duration
	interval, ok := config["interval"]
	if ok {
		intervalStr, ok := interval.(string)
		if !ok {
			badPurgeUploadConfig("interval is not a string")
		}

		intervalDuration, err = time.ParseDuration(intervalStr)
		if err != nil {
			badPurgeUploadConfig(fmt.Sprintf("Cannot parse interval: %s", err.Error()))
		}
	} else {
		badPurgeUploadConfig("interval missing")
	}

	var dryRunBool bool
	dryRun, ok := config["dryrun"]
	if ok {
		dryRunBool, ok = dryRun.(bool)
		if !ok {
			badPurgeUploadConfig("cannot parse dryrun")
		}
	} else {
		badPurgeUploadConfig("dryrun missing")
	}

	go func() {
		rand.Seed(time.Now().Unix())
		jitter := time.Duration(rand.Int()%60) * time.Minute
		log.Infof("Starting upload purge in %s", jitter)
		time.Sleep(jitter)

		for {
			storage.PurgeUploads(ctx, storageDriver, time.Now().Add(-purgeAgeDuration), !dryRunBool)
			log.Infof("Starting upload purge in %s", intervalDuration)
			time.Sleep(intervalDuration)
		}
	}()
}
