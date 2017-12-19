package model

import "github.com/jinzhu/gorm"

type TemplateCategory struct {
	TemplateId uint `sql:"type:integer REFERENCES catalog_template(id) ON DELETE CASCADE"`
	CategoryId uint `sql:"type:integer REFERENCES catalog_category(id) ON DELETE CASCADE"`
}

type TemplateCategoryModel struct {
	Base
	TemplateCategory
}

func lookupTemplateCategories(db *gorm.DB, templateId uint) []string {
	var categoryModels []CategoryModel
	db.Raw(`
SELECT *
FROM catalog_template_category, catalog_category
WHERE catalog_template_category.template_id = ?
AND catalog_template_category.category_id = catalog_category.id
	`, templateId).Scan(&categoryModels)

	var categories []string
	for _, categoryModel := range categoryModels {
		categories = append(categories, categoryModel.Name)
	}
	return categories
}
