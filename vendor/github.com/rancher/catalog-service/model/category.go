package model

type Category struct {
	Name string
}

type CategoryModel struct {
	Base
	Category
}

type CategoryAndTemplate struct {
	TemplateID int
	CategoryID int
	Category
}

type LabelAndTemplate struct {
	TemplateID string
	Key        string
	Value      string
}
