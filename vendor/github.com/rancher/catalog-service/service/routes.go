package service

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/rancher/catalog-service/manager"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/go-rancher/api"
	"github.com/rancher/go-rancher/client"
)

// MuxWrapper is a wrapper over the mux router that returns 503 until catalog is ready
type MuxWrapper struct {
	IsReady bool
	Router  *mux.Router
}

func (httpWrapper *MuxWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpWrapper.Router.ServeHTTP(w, r)
}

// TODO
var schemas *client.Schemas

// TODO
var m *manager.Manager
var db *gorm.DB

func handler(schemas *client.Schemas, envIdRequired bool, f func(http.ResponseWriter, *http.Request, string) (int, error)) http.Handler {
	return api.ApiHandler(schemas, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		envId, err := getEnvironmentId(r)
		if err != nil {
			if envIdRequired {
				ReturnHTTPError(w, r, http.StatusBadRequest, err)
				return
			}
			envId = "global"
		}
		if code, err := f(w, r, envId); err != nil {
			ReturnHTTPError(w, r, code, err)
			return
		}
	}))
}

func NewRouter(manager *manager.Manager, gormDb *gorm.DB) *mux.Router {
	// TODO
	m = manager
	db = gormDb

	schemas := &client.Schemas{}

	apiVersion := schemas.AddType("apiVersion", client.Resource{})
	apiVersion.CollectionMethods = []string{}

	schemas.AddType("schema", client.Schema{})

	schemas.AddType("catalog", model.CatalogResource{})

	template := schemas.AddType("template", model.TemplateResource{})
	template.CollectionActions = map[string]client.Action{
		"refresh": {},
	}
	delete(template.ResourceFields, "icon")
	delete(template.ResourceFields, "readme")
	delete(template.ResourceFields, "projectURL")

	question := schemas.AddType("question", model.Question{})
	question.CollectionMethods = []string{}

	templateVersion := schemas.AddType("templateVersion", model.TemplateVersionResource{})
	delete(templateVersion.ResourceFields, "readme")

	templateVersionQuestions := templateVersion.ResourceFields["questions"]
	templateVersionQuestions.Type = "array[question]"
	templateVersion.ResourceFields["questions"] = templateVersionQuestions

	err := schemas.AddType("error", model.CatalogError{})
	err.CollectionMethods = []string{}

	// API framework routes
	router := mux.NewRouter().StrictSlash(true)

	router.Methods("GET").Path("/").Handler(api.VersionsHandler(schemas, "v1-catalog"))
	router.Methods("GET").Path("/v1-catalog/schemas").Handler(api.SchemasHandler(schemas))
	router.Methods("GET").Path("/v1-catalog/schemas/{id}").Handler(api.SchemaHandler(schemas))
	router.Methods("GET").Path("/v1-catalog").Handler(api.VersionHandler(schemas, "v1-catalog"))

	router.Methods("GET").Path("/v1-catalog/catalogs").Name("GetCatalogs").Handler(handler(schemas, false, getCatalogs))
	router.Methods("POST").Path("/v1-catalog/catalogs").Name("CreateCatalog").Handler(handler(schemas, true, createCatalog))
	router.Methods("GET").Path("/v1-catalog/catalogs/{catalog}").Name("GetCatalog").Handler(handler(schemas, false, getCatalog))
	router.Methods("PUT").Path("/v1-catalog/catalogs/{catalog}").Name("UpdateCatalog").Handler(handler(schemas, false, updateCatalog))
	router.Methods("DELETE").Path("/v1-catalog/catalogs/{catalog}").Name("DeleteCatalog").Handler(handler(schemas, true, deleteCatalog))

	router.Methods("GET").Path("/v1-catalog/templates").Name("GetTemplates").Handler(handler(schemas, false, getTemplates))
	router.Methods("GET").Path("/v1-catalog/templates/{catalog_template_version}").Name("GetTemplate").Handler(handler(schemas, false, getTemplate))
	router.Methods("GET").Path("/v1-catalog/templateversions/{catalog_template_version}").Name("GetTemplate").Handler(handler(schemas, false, getTemplate))
	router.Methods("POST").Path("/v1-catalog/templates").Name("RefreshTemplates").Handler(handler(schemas, false, refreshTemplates))
	router.GetRoute("RefreshTemplates").Queries("action", "refresh")

	return router
}
