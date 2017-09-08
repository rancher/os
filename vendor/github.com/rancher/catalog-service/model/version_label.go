package model

import "github.com/jinzhu/gorm"

type VersionLabel struct {
	VersionId uint `sql:"type:integer REFERENCES catalog_version(id) ON DELETE CASCADE"`

	Key   string
	Value string
}

type VersionLabelModel struct {
	Base
	VersionLabel
}

func lookupVersionLabels(db *gorm.DB, versonId uint) map[string]string {
	var labelModels []VersionLabelModel
	db.Where(&VersionLabelModel{
		VersionLabel: VersionLabel{
			VersionId: versonId,
		},
	}).Find(&labelModels)

	labels := map[string]string{}
	for _, label := range labelModels {
		labels[label.Key] = label.Value
	}

	return labels
}
