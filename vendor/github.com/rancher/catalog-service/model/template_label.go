package model

import "github.com/jinzhu/gorm"

type TemplateLabel struct {
	TemplateId uint `sql:"type:integer REFERENCES catalog_template(id) ON DELETE CASCADE"`

	Key   string
	Value string
}

type TemplateLabelModel struct {
	Base
	TemplateLabel
}

func lookupTemplateLabels(db *gorm.DB, templateId uint) map[string]string {
	var labelModels []TemplateLabelModel
	db.Where(&TemplateLabelModel{
		TemplateLabel: TemplateLabel{
			TemplateId: templateId,
		},
	}).Find(&labelModels)

	labels := map[string]string{}
	for _, label := range labelModels {
		labels[label.Key] = label.Value
	}

	return labels
}
