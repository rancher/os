package model

import (
	"fmt"
	"strings"

	"github.com/docker/libcompose/utils"
	"github.com/jinzhu/gorm"
	"github.com/rancher/go-rancher/v2"
)

type Template struct {
	EnvironmentId string `json:"environmentId"`
	CatalogId     uint   `sql:"type:integer REFERENCES catalog(id) ON DELETE CASCADE"`

	Name           string `json:"name"`
	IsSystem       string `json:"isSystem"`
	Description    string `json:"description"`
	DefaultVersion string `json:"defaultVersion" yaml:"default_version"`
	Path           string `json:"path"`
	Maintainer     string `json:"maintainer"`
	License        string `json:"license"`
	ProjectURL     string `json:"projectURL" yaml:"project_url"`
	UpgradeFrom    string `json:"upgradeFrom"`
	FolderName     string `json:"folderName"`
	Catalog        string `json:"catalogId"`
	Base           string `json:"templateBase"`
	Icon           string `json:"icon"`
	IconFilename   string `json:"iconFilename"`
	Readme         string `json:"readme"`

	Categories []string          `sql:"-" json:"categories"`
	Labels     map[string]string `sql:"-" json:"labels"`

	Versions []Version `sql:"-"`
	Category string    `sql:"-"`
}

type TemplateModel struct {
	Base
	Template
}

type TemplateResource struct {
	client.Resource
	Template

	VersionLinks             map[string]string `json:"versionLinks"`
	DefaultTemplateVersionId string            `json:"defaultTemplateVersionId"`
}

type TemplateCollection struct {
	client.Collection
	Data []TemplateResource `json:"data,omitempty"`
}

func LookupTemplate(db *gorm.DB, environmentId, catalog, folderName, base string) *Template {
	var templateModel TemplateModel
	if err := db.Raw(`
SELECT catalog_template.*
FROM catalog_template, catalog
WHERE (catalog_template.environment_id = ? OR catalog_template.environment_id = ?)
AND catalog_template.catalog_id = catalog.id
AND catalog.name = ?
AND catalog_template.base = ?
AND catalog_template.folder_name = ?
`, environmentId, "global", catalog, base, folderName).Scan(&templateModel).Error; err == gorm.ErrRecordNotFound {
		return nil
	}

	fillInTemplate(db, &templateModel)
	return &templateModel.Template
}

func fillInTemplate(db *gorm.DB, templateModel *TemplateModel) {
	catalog := GetCatalog(db, templateModel.CatalogId)
	if catalog != nil {
		templateModel.Catalog = catalog.Name
	}

	templateModel.Categories = lookupTemplateCategories(db, templateModel.ID)
	templateModel.Labels = lookupTemplateLabels(db, templateModel.ID)
	templateModel.Versions = lookupVersions(db, templateModel.ID)
}

func templateCategoryMap(db *gorm.DB, templateIDList []int) map[int][]string {
	categoriesQuery := `
	SELECT template_id, category_id, name
	FROM catalog_template_category tc
	JOIN catalog_category c ON (tc.category_id = c.id)
	WHERE tc.template_id IN ( ? )`

	catagoryAndTemplateList := []CategoryAndTemplate{}
	db.Raw(categoriesQuery, templateIDList).Find(&catagoryAndTemplateList)

	// make map of template (key) to category name (value)

	var catagoryAndTemplateMap map[int][]string
	catagoryAndTemplateMap = make(map[int][]string)
	for _, catagoryAndTemplate := range catagoryAndTemplateList {

		catagoryAndTemplateMap[catagoryAndTemplate.TemplateID] = append(catagoryAndTemplateMap[catagoryAndTemplate.TemplateID], catagoryAndTemplate.Name)

	}

	return catagoryAndTemplateMap
}

func templateLabelMap(db *gorm.DB, templateIDList []int) map[int]map[string]string {
	labelsQuery := "SELECT template_id, `key`, value FROM catalog_label cl WHERE cl.template_id IN ( ? )"

	var labelAndTemplateList []TemplateLabelModel
	db.Raw(labelsQuery, templateIDList).Find(&labelAndTemplateList)

	var labelAndTemplateMap map[int]map[string]string
	labelAndTemplateMap = make(map[int]map[string]string)

	for _, labelAndTemplate := range labelAndTemplateList {

		if _, ok := labelAndTemplateMap[int(labelAndTemplate.TemplateId)]; !ok {
			labels := map[string]string{labelAndTemplate.Key: labelAndTemplate.Value}
			labelAndTemplateMap[int(labelAndTemplate.TemplateId)] = labels
		} else {
			labelAndTemplateMap[int(labelAndTemplate.TemplateId)][labelAndTemplate.Key] = labelAndTemplate.Value
		}
	}

	return labelAndTemplateMap
}

func templateVersionMap(db *gorm.DB, templateIDList []int) map[int][]Version {
	var versionList []VersionModel

	// all versions with list of template IDs
	versionsQuery := `
	SELECT *
	FROM catalog_version
	WHERE catalog_version.template_id IN ( ? )`
	db.Raw(versionsQuery, templateIDList).Find(&versionList)

	// look up version based on version id
	versionMap := map[uint]VersionModel{}
	var versionIDs []int

	for _, version := range versionList {

		versionIDs = append(versionIDs, int(version.ID))

		versionMap[version.ID] = version
	}

	var versionLabelList []VersionLabelModel

	versionLabelsQuery := `
	SELECT *
	FROM catalog_version_label
	WHERE catalog_version_label.version_id IN (?)`

	db.Raw(versionLabelsQuery, versionIDs).Find(&versionLabelList)

	for _, label := range versionLabelList {

		if versionMap[label.VersionId].Labels == nil {
			version := versionMap[label.VersionId]
			version.Labels = make(map[string]string)

			versionMap[label.VersionId] = version
		}

		versionMap[label.VersionId].Labels[label.Key] = label.Value
	}

	var versionFiles []FileModel

	versionFilesQuery := `
	SELECT *
	FROM catalog_file
	WHERE catalog_file.version_id IN ( ? )`
	db.Raw(versionFilesQuery, versionIDs).Find(&versionFiles)

	for _, file := range versionFiles {
		version := versionMap[file.VersionId]

		version.Files = append(versionMap[file.VersionId].Files, file.File)

		versionMap[file.VersionId] = version
	}

	templateVersionMap := map[int][]Version{}

	for _, version := range versionMap {

		templateVersionMap[int(version.TemplateId)] = append(templateVersionMap[int(version.TemplateId)], version.Version)
	}

	return templateVersionMap
}

func catalogMap(db *gorm.DB, templateIDList []int) map[uint]string {
	catalogQuery := `
	SELECT catalog.*
    FROM catalog
    JOIN catalog_template ON (catalog.id = catalog_template.catalog_id)
    WHERE catalog_template.id IN ( ? )`

	var catalogs []CatalogModel

	db.Raw(catalogQuery, templateIDList).Find(&catalogs)

	// make map of catalog id to catalog name

	catalogMap := map[uint]string{}

	for _, catalog := range catalogs {

		catalogMap[catalog.ID] = catalog.Name

	}

	return catalogMap

}

func LookupTemplates(db *gorm.DB, environmentId, catalog, templateBaseEq string, categories, categoriesNe []string) []Template {
	var templateModels []TemplateModel

	params := []interface{}{environmentId, "global"}
	if catalog != "" {
		params = append(params, catalog)
	}
	if templateBaseEq != "" {
		params = append(params, templateBaseEq)
	}

	query := `
	SELECT catalog_template.*
	FROM catalog_template, catalog
	WHERE (catalog_template.environment_id = ? OR catalog_template.environment_id = ?)
	AND catalog_template.catalog_id = catalog.id`

	if catalog != "" {
		query += `
AND catalog.name = ?`
	}
	if templateBaseEq != "" {
		query += `
AND catalog_template.base = ?`
	}

	db.Raw(query, params...).Find(&templateModels)

	var templateIDList []int
	for _, template := range templateModels {
		templateIDList = append(templateIDList, int(template.ID))
	}

	templateCategoryMap := templateCategoryMap(db, templateIDList)
	templateLabelMap := templateLabelMap(db, templateIDList)
	templateVersionMap := templateVersionMap(db, templateIDList)
	catalogMap := catalogMap(db, templateIDList)

	var templates []Template
	for _, templateModel := range templateModels {
		templateModel.Categories = templateCategoryMap[int(templateModel.ID)]
		templateModel.Labels = templateLabelMap[int(templateModel.ID)]
		templateModel.Versions = templateVersionMap[int(templateModel.ID)]
		templateModel.Catalog = catalogMap[templateModel.CatalogId]

		skip := false
		for _, category := range categories {
			if !utils.Contains(templateModel.Categories, category) {
				skip = true
				break
			}
		}
		for _, categoryNe := range categoriesNe {
			if utils.Contains(templateModel.Categories, categoryNe) {
				skip = true
				break
			}
		}
		if !skip {
			templates = append(templates, templateModel.Template)
		}
	}
	return templates
}

func listQuery(size int) string {
	var query string
	for i := 0; i < size; i++ {
		query += " ? ,"
	}
	return fmt.Sprintf("(%s)", strings.TrimSuffix(query, ","))
}
