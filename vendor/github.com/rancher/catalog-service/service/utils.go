package service

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/catalog-service/parse"
	"github.com/rancher/catalog-service/utils"
	"github.com/rancher/go-rancher/api"
	"github.com/rancher/go-rancher/v2"
)

const (
	environmentIdHeader = "x-api-project-id"
)

func getEnvironmentId(r *http.Request) (string, error) {
	environment := r.Header.Get(environmentIdHeader)
	if environment == "" {
		environment = r.URL.Query().Get("projectId")
		if environment == "" {
			return "", fmt.Errorf("Request is missing environment header")
		}
	}
	return environment, nil
}

func ReturnHTTPError(w http.ResponseWriter, r *http.Request, httpStatus int, err error) {
	w.WriteHeader(httpStatus)

	catalogError := model.CatalogError{
		Resource: client.Resource{
			Type: "error",
		},
		Status:  strconv.Itoa(httpStatus),
		Message: err.Error(),
	}

	api.GetApiContext(r).Write(&catalogError)
}

// TODO: this should return an error
func URLEncoded(str string) string {
	u, err := url.Parse(str)
	if err != nil {
		log.Errorf("Error encoding the url: %s , error: %v", str, err)
		return str
	}
	return u.String()
}

func generateVersionId(catalogName string, template model.Template, version model.Version) string {
	versionId := generateTemplateId(catalogName, template)
	if version.Revision == nil {
		versionId += fmt.Sprintf(":%s", version.Version)
	} else {
		versionId += fmt.Sprintf(":%d", *version.Revision)
	}
	return versionId
}

func generateTemplateId(catalogName string, template model.Template) string {
	if template.Base == "" {
		return fmt.Sprintf("%s:%s", catalogName, template.FolderName)
	}
	return fmt.Sprintf("%s:%s*%s", catalogName, template.Base, template.FolderName)
}

func catalogResource(catalog model.Catalog, apiContext *api.ApiContext, envId string) *model.CatalogResource {
	selfLink := apiContext.UrlBuilder.ReferenceByIdLink("catalogs", catalog.Name)
	projectID := envId
	if projectID != "" {
		selfLink = selfLink + "?projectId=" + projectID
	}

	return &model.CatalogResource{
		Resource: client.Resource{
			Id:    catalog.Name,
			Type:  "catalog",
			Links: map[string]string{"self": selfLink},
		},
		Catalog: catalog,
	}
}

func templateDefaultVersion(template model.Template, catalogName string, apiContext *api.ApiContext) (string, string) {
	var defaultTemplateVersionId string
	for _, version := range template.Versions {
		if version.Version == template.DefaultVersion {
			defaultTemplateVersionId = generateVersionId(catalogName, template, version)
			selfLink := apiContext.UrlBuilder.ReferenceByIdLink("templates", defaultTemplateVersionId)
			return selfLink, defaultTemplateVersionId
		}
	}

	sort.Sort(model.Versions(template.Versions))
	if len(template.Versions) != 0 {
		defaultTemplateVersionId = generateVersionId(catalogName, template, template.Versions[len(template.Versions)-1])
		selfLink := apiContext.UrlBuilder.ReferenceByIdLink("templates", defaultTemplateVersionId)
		return selfLink, defaultTemplateVersionId
	}

	return "", ""
}

func templateResource(apiContext *api.ApiContext, catalogName string, template model.Template, rancherVersion string, envId string) *model.TemplateResource {
	templateId := generateTemplateId(catalogName, template)

	versionLinks := map[string]string{}
	for _, version := range template.Versions {
		if utils.VersionBetween(version.MinimumRancherVersion, rancherVersion, version.MaximumRancherVersion) {
			route := generateVersionId(catalogName, template, version)
			link := apiContext.UrlBuilder.ReferenceByIdLink("template", route)
			versionLinks[version.Version] = URLEncoded(link)
		}
	}

	links := map[string]string{}

	links["icon"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", fmt.Sprintf("%s?image&projectId=%s", templateId, envId)))
	if template.Readme != "" {
		links["readme"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", fmt.Sprintf("%s?readme", templateId)))
	}
	if template.ProjectURL != "" {
		links["project"] = template.ProjectURL
	}

	defaultVersion, defaultVersionId := templateDefaultVersion(template, catalogName, apiContext)
	if defaultVersion != "" {
		links["defaultVersion"] = defaultVersion
	}

	return &model.TemplateResource{
		Resource: client.Resource{
			Id:    templateId,
			Type:  "template",
			Links: links,
		},
		Template:                 template,
		VersionLinks:             versionLinks,
		DefaultTemplateVersionId: defaultVersionId,
	}
}

func defaultUpgradeVersionLink(upgradeVersions []model.Version, catalogName string, template model.Template, apiContext *api.ApiContext) string {
	sort.Sort(model.Versions(upgradeVersions))
	if len(upgradeVersions) != 0 {
		route := generateVersionId(catalogName, template, upgradeVersions[len(upgradeVersions)-1])
		link := apiContext.UrlBuilder.ReferenceByIdLink("template", route)
		defaultUpgradeVersionLink := URLEncoded(link)
		return defaultUpgradeVersionLink
	}

	return ""
}

func versionResource(apiContext *api.ApiContext, catalogName string, template model.Template, version model.Version, rancherVersion string, envId string) (*model.TemplateVersionResource, error) {
	templateId := generateTemplateId(catalogName, template)
	versionId := generateVersionId(catalogName, template, version)

	filesMap := map[string]string{}
	for _, file := range version.Files {
		filesMap[file.Name] = file.Contents
	}

	var questions []model.Question
	templateVersion, templateVersionExists := filesMap["template-version.yml"]
	compose, composeExists := filesMap["compose.yml"]
	rancherCompose, rancherComposeExists := filesMap["rancher-compose.yml"]
	if templateVersionExists {
		catalogInfo, err := parse.CatalogInfoFromTemplateVersion([]byte(templateVersion))
		if err != nil {
			return nil, err
		}
		questions = catalogInfo.Questions
	} else if composeExists {
		catalogInfo, err := parse.CatalogInfoFromCompose([]byte(compose))
		if err != nil {
			return nil, err
		}
		questions = catalogInfo.Questions
	} else if rancherComposeExists {
		catalogInfo, err := parse.CatalogInfoFromRancherCompose([]byte(rancherCompose))
		if err != nil {
			return nil, err
		}
		questions = catalogInfo.Questions
	}

	links := map[string]string{}
	links["icon"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", fmt.Sprintf("%s?image&projectId=%s", templateId, envId)))

	if version.Readme != "" {
		links["readme"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", fmt.Sprintf("%s?readme", versionId)))
	} else if template.Readme != "" {
		links["readme"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", fmt.Sprintf("%s?readme", templateId)))
	}
	if template.ProjectURL != "" {
		links["project"] = template.ProjectURL
	}

	links["template"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", templateId))

	upgradeVersionLinks := map[string]string{}
	upgradeVersions := []model.Version{}
	for _, upgradeVersion := range template.Versions {
		if showUpgradeVersion(version, upgradeVersion, rancherVersion) {
			route := generateVersionId(catalogName, template, upgradeVersion)
			link := apiContext.UrlBuilder.ReferenceByIdLink("template", route)
			upgradeVersionLinks[upgradeVersion.Version] = URLEncoded(link)
			upgradeVersions = append(upgradeVersions, upgradeVersion)

			if template.DefaultVersion == upgradeVersion.Version {
				links["defaultUpgradeVersion"] = URLEncoded(link)
			}
		}
	}

	if _, ok := links["defaultUpgradeVersion"]; !ok {
		defaultUpgradeVersionLink := defaultUpgradeVersionLink(upgradeVersions, catalogName, template, apiContext)
		if defaultUpgradeVersionLink != "" {
			links["defaultUpgradeVersion"] = defaultUpgradeVersionLink
		}
	}

	return &model.TemplateVersionResource{
		Resource: client.Resource{
			Id:    versionId,
			Type:  "templateVersion",
			Links: links,
		},
		Version:             version,
		Files:               filesMap,
		Questions:           questions,
		UpgradeVersionLinks: upgradeVersionLinks,
		TemplateId:          templateId,
	}, nil
}

func showUpgradeVersion(version, upgradeVersion model.Version, rancherVersion string) bool {
	if !utils.VersionGreaterThan(upgradeVersion.Version, version.Version) {
		return false
	}
	if !utils.VersionBetween(upgradeVersion.MinimumRancherVersion, rancherVersion, upgradeVersion.MaximumRancherVersion) {
		return false
	}
	if upgradeVersion.UpgradeFrom != "" {
		satisfiesRange, err := utils.VersionSatisfiesRange(version.Version, upgradeVersion.UpgradeFrom)
		if err != nil {
			return false
		}
		return satisfiesRange
	}
	return true
}
