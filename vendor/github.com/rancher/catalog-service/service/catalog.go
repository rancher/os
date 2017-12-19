package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/go-rancher/api"
)

func getCatalogs(w http.ResponseWriter, r *http.Request, envId string) (int, error) {
	apiContext := api.GetApiContext(r)

	catalogs := model.LookupCatalogs(db, envId)

	resp := model.CatalogCollection{}
	for _, catalog := range catalogs {

		resp.Data = append(resp.Data, *catalogResource(catalog, apiContext, envId))
	}

	apiContext.Write(&resp)
	return 0, nil
}

func getCatalog(w http.ResponseWriter, r *http.Request, envId string) (int, error) {
	apiContext := api.GetApiContext(r)

	vars := mux.Vars(r)
	envId, err := getEnvironmentId(r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// TODO error checking
	catalogName := vars["catalog"]
	catalog := model.LookupCatalog(db, envId, catalogName)
	if catalog == nil {
		return http.StatusNotFound, errors.New("Catalog not found")
	}

	apiContext.Write(catalogResource(*catalog, apiContext, envId))
	return 0, nil
}

type CatalogRequest struct {
	Name   string
	URL    string
	Branch string
	Kind   string
}

func isDuplicateName(catalogModel *model.CatalogModel) bool {

	catalogs := []model.CatalogModel{}
	catalogsQuery := `
	SELECT *
	FROM catalog
	WHERE (environment_id = "global" OR environment_id = ?)
	AND name = ?`
	db.Raw(catalogsQuery, catalogModel.EnvironmentId, catalogModel.Name).Find(&catalogs)

	if len(catalogs) > 0 {
		return true
	}

	return false
}

func isDuplicateEnvName(catalogModel *model.CatalogModel, oldCatalogName string) bool {

	if oldCatalogName == catalogModel.Name {
		return false
	}

	catalogs := []model.CatalogModel{}
	catalogsQuery := `
	SELECT *
	FROM catalog
	WHERE (environment_id = ?)
	AND name = ?`
	db.Raw(catalogsQuery, catalogModel.EnvironmentId, catalogModel.Name).Find(&catalogs)

	if len(catalogs) > 0 {
		return true
	}

	return false
}

func isDuplicateGlobalName(catalogModel *model.CatalogModel) bool {
	catalogs := []model.CatalogModel{}
	catalogsQuery := `
	SELECT *
	FROM catalog
	WHERE environment_id = "global"
	AND name = ?`
	db.Raw(catalogsQuery, catalogModel.EnvironmentId, catalogModel.Name).Find(&catalogs)

	if len(catalogs) > 0 {
		return true
	}

	return false
}

func createCatalog(w http.ResponseWriter, r *http.Request, envId string) (int, error) {
	apiContext := api.GetApiContext(r)

	catalogModel, err := catalogModelFromRequest(r, envId)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if catalogModel.Name == "" {
		return http.StatusBadRequest, errors.New("Missing field 'name'")
	}
	if catalogModel.URL == "" {
		return http.StatusBadRequest, errors.New("Missing field 'url'")
	}

	if isDuplicateName(catalogModel) {
		return http.StatusUnprocessableEntity, fmt.Errorf("Catalog name %s already exists", catalogModel.Name)
	}

	if err := db.Create(catalogModel).Error; err != nil {
		return http.StatusBadRequest, err
	}

	apiContext.Write(catalogResource(catalogModel.Catalog, apiContext, ""))
	return 0, nil
}

func catalogExists(catalogModel *model.CatalogModel, envId string) bool {
	query := `
	SELECT *
	FROM catalog
	WHERE name = ?
	AND environment_id = ?`

	catalog := []model.CatalogModel{}

	db.Raw(query, catalogModel.Name, envId).Find(&catalog)

	if len(catalog) > 0 {
		return true
	}

	return false
}

func updateCatalog(w http.ResponseWriter, r *http.Request, envId string) (int, error) {
	apiContext := api.GetApiContext(r)

	catalogModel, err := catalogModelFromRequest(r, envId)
	if err != nil {
		return http.StatusBadRequest, err
	}

	vars := mux.Vars(r)
	oldCatalogName := vars["catalog"]

	oldCatalog := model.CatalogModel{
		Catalog: model.Catalog{
			Name:          oldCatalogName,
			EnvironmentId: envId,
		},
	}

	if isDuplicateGlobalName(catalogModel) {
		return http.StatusUnprocessableEntity, fmt.Errorf("Catalog name %s already exists", catalogModel.Name)
	}

	if isDuplicateEnvName(catalogModel, oldCatalogName) {
		return http.StatusUnprocessableEntity, fmt.Errorf("Catalog name %s already exists", catalogModel.Name)
	}

	if !catalogExists(&oldCatalog, envId) {
		return http.StatusNotFound, errors.New("Catalog not found")
	}

	if err := db.Model(&model.CatalogModel{}).Where(&oldCatalog).Update(catalogModel).Error; err != nil {
		return http.StatusBadRequest, err
	}

	apiContext.Write(catalogResource(catalogModel.Catalog, apiContext, ""))
	return 0, nil
}

func catalogModelFromRequest(r *http.Request, envId string) (*model.CatalogModel, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var catalogRequest CatalogRequest
	if err := json.Unmarshal(body, &catalogRequest); err != nil {
		return nil, err
	}

	return &model.CatalogModel{
		Catalog: model.Catalog{
			EnvironmentId: envId,
			Name:          catalogRequest.Name,
			URL:           catalogRequest.URL,
			Branch:        catalogRequest.Branch,
			Kind:          catalogRequest.Kind,
		},
	}, nil
}

func deleteCatalog(w http.ResponseWriter, r *http.Request, envId string) (int, error) {
	vars := mux.Vars(r)

	name, ok := vars["catalog"]
	if !ok {
		return http.StatusBadRequest, errors.New("Missing paramater catalog")
	}

	model.DeleteCatalog(db, envId, name)

	w.WriteHeader(http.StatusNoContent)
	return 0, nil
}

func getCatalogTemplates(w http.ResponseWriter, r *http.Request, envId string) (int, error) {
	apiContext := api.GetApiContext(r)
	vars := mux.Vars(r)

	catalogName, ok := vars["catalog"]
	if !ok {
		return http.StatusBadRequest, errors.New("Missing paramater catalog")
	}

	rancherVersion := r.URL.Query().Get("rancherVersion")
	templateBaseEq := r.URL.Query().Get("templateBase_eq")
	categories, _ := r.URL.Query()["category"]
	categoriesNe, _ := r.URL.Query()["category_ne"]

	templates := model.LookupTemplates(db, envId, catalogName, templateBaseEq, categories, categoriesNe)

	// TODO: this is duplicated
	resp := model.TemplateCollection{}
	for _, template := range templates {
		templateResource := templateResource(apiContext, catalogName, template, rancherVersion, envId)
		if len(templateResource.VersionLinks) > 0 {
			resp.Data = append(resp.Data, *templateResource)
		}
	}

	resp.Actions = map[string]string{
		"refresh": api.GetApiContext(r).UrlBuilder.ReferenceByIdLink("template", "") + "?action=refresh",
	}

	apiContext.Write(&resp)
	return 0, nil
}
