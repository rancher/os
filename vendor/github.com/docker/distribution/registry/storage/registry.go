package storage

import (
	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/storage/cache"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
)

// registry is the top-level implementation of Registry for use in the storage
// package. All instances should descend from this object.
type registry struct {
	blobStore                   *blobStore
	blobServer                  distribution.BlobServer
	statter                     distribution.BlobStatter // global statter service.
	blobDescriptorCacheProvider cache.BlobDescriptorCacheProvider
	deleteEnabled               bool
	resumableDigestEnabled      bool
}

// NewRegistryWithDriver creates a new registry instance from the provided
// driver. The resulting registry may be shared by multiple goroutines but is
// cheap to allocate. If redirect is true, the backend blob server will
// attempt to use (StorageDriver).URLFor to serve all blobs.
//
// TODO(stevvooe): This function signature is getting very out of hand. Move to
// functional options for instance configuration.
func NewRegistryWithDriver(ctx context.Context, driver storagedriver.StorageDriver, blobDescriptorCacheProvider cache.BlobDescriptorCacheProvider, deleteEnabled bool, redirect bool, isCache bool) distribution.Namespace {
	// create global statter, with cache.
	var statter distribution.BlobDescriptorService = &blobStatter{
		driver: driver,
		pm:     defaultPathMapper,
	}

	if blobDescriptorCacheProvider != nil {
		statter = cache.NewCachedBlobStatter(blobDescriptorCacheProvider, statter)
	}

	bs := &blobStore{
		driver:  driver,
		pm:      defaultPathMapper,
		statter: statter,
	}

	return &registry{
		blobStore: bs,
		blobServer: &blobServer{
			driver:   driver,
			statter:  statter,
			pathFn:   bs.path,
			redirect: redirect,
		},
		blobDescriptorCacheProvider: blobDescriptorCacheProvider,
		deleteEnabled:               deleteEnabled,
		resumableDigestEnabled:      !isCache,
	}
}

// Scope returns the namespace scope for a registry. The registry
// will only serve repositories contained within this scope.
func (reg *registry) Scope() distribution.Scope {
	return distribution.GlobalScope
}

// Repository returns an instance of the repository tied to the registry.
// Instances should not be shared between goroutines but are cheap to
// allocate. In general, they should be request scoped.
func (reg *registry) Repository(ctx context.Context, name string) (distribution.Repository, error) {
	if err := v2.ValidateRepositoryName(name); err != nil {
		return nil, distribution.ErrRepositoryNameInvalid{
			Name:   name,
			Reason: err,
		}
	}

	var descriptorCache distribution.BlobDescriptorService
	if reg.blobDescriptorCacheProvider != nil {
		var err error
		descriptorCache, err = reg.blobDescriptorCacheProvider.RepositoryScoped(name)
		if err != nil {
			return nil, err
		}
	}

	return &repository{
		ctx:             ctx,
		registry:        reg,
		name:            name,
		descriptorCache: descriptorCache,
	}, nil
}

// repository provides name-scoped access to various services.
type repository struct {
	*registry
	ctx             context.Context
	name            string
	descriptorCache distribution.BlobDescriptorService
}

// Name returns the name of the repository.
func (repo *repository) Name() string {
	return repo.name
}

// Manifests returns an instance of ManifestService. Instantiation is cheap and
// may be context sensitive in the future. The instance should be used similar
// to a request local.
func (repo *repository) Manifests(ctx context.Context, options ...distribution.ManifestServiceOption) (distribution.ManifestService, error) {
	ms := &manifestStore{
		ctx:        ctx,
		repository: repo,
		revisionStore: &revisionStore{
			ctx:        ctx,
			repository: repo,
			blobStore: &linkedBlobStore{
				ctx:           ctx,
				blobStore:     repo.blobStore,
				repository:    repo,
				deleteEnabled: repo.registry.deleteEnabled,
				blobAccessController: &linkedBlobStatter{
					blobStore:  repo.blobStore,
					repository: repo,
					linkPath:   manifestRevisionLinkPath,
				},

				// TODO(stevvooe): linkPath limits this blob store to only
				// manifests. This instance cannot be used for blob checks.
				linkPath: manifestRevisionLinkPath,
			},
		},
		tagStore: &tagStore{
			ctx:        ctx,
			repository: repo,
			blobStore:  repo.registry.blobStore,
		},
	}

	// Apply options
	for _, option := range options {
		err := option(ms)
		if err != nil {
			return nil, err
		}
	}

	return ms, nil
}

// Blobs returns an instance of the BlobStore. Instantiation is cheap and
// may be context sensitive in the future. The instance should be used similar
// to a request local.
func (repo *repository) Blobs(ctx context.Context) distribution.BlobStore {
	var statter distribution.BlobDescriptorService = &linkedBlobStatter{
		blobStore:  repo.blobStore,
		repository: repo,
		linkPath:   blobLinkPath,
	}

	if repo.descriptorCache != nil {
		statter = cache.NewCachedBlobStatter(repo.descriptorCache, statter)
	}

	return &linkedBlobStore{
		blobStore:            repo.blobStore,
		blobServer:           repo.blobServer,
		blobAccessController: statter,
		repository:           repo,
		ctx:                  ctx,

		// TODO(stevvooe): linkPath limits this blob store to only layers.
		// This instance cannot be used for manifest checks.
		linkPath:      blobLinkPath,
		deleteEnabled: repo.registry.deleteEnabled,
	}
}

func (repo *repository) Signatures() distribution.SignatureService {
	return &signatureStore{
		repository: repo,
		blobStore:  repo.blobStore,
		ctx:        repo.ctx,
	}
}
