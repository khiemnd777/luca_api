package model

type ImportFieldProfile struct {
	ID          int     `json:"id"`
	Scope       string  `json:"scope"`       // clinics, dentists, orders...
	Code        string  `json:"code"`        // default, vn-template-2025...
	Name        string  `json:"name"`        // label hiển thị
	Description *string `json:"description"` // optional
	IsDefault   bool    `json:"is_default"`
}

type ImportFieldProfileInput struct {
	Scope       string  `json:"scope"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsDefault   bool    `json:"is_default"`
}

type ImportFieldMapping struct {
	ID        int `json:"id"`
	ProfileID int `json:"profile_id"`

	InternalKind  string `json:"internal_kind"`  // core | metadata | external
	InternalPath  string `json:"internal_path"`  // name, phone_number, tax_code...
	InternalLabel string `json:"internal_label"` // label hiển thị

	MetadataCollectionSlug *string `json:"metadata_collection_slug"`
	MetadataFieldName      *string `json:"metadata_field_name"`

	DataType string `json:"data_type"` // text, number, date...

	ExcelHeader *string `json:"excel_header"`
	ExcelColumn *int    `json:"excel_column"`

	Required bool `json:"required"`
	Unique   bool `json:"unique"`

	TransformHint *string `json:"transform_hint"`
}

type ImportFieldMappingInput struct {
	ProfileID int `json:"profile_id"`

	InternalKind  string `json:"internal_kind"`
	InternalPath  string `json:"internal_path"`
	InternalLabel string `json:"internal_label"`

	MetadataCollectionSlug *string `json:"metadata_collection_slug"`
	MetadataFieldName      *string `json:"metadata_field_name"`

	DataType string `json:"data_type"`

	ExcelHeader *string `json:"excel_header"`
	ExcelColumn *int    `json:"excel_column"`

	Required bool `json:"required"`
	Unique   bool `json:"unique"`

	TransformHint *string `json:"transform_hint"`
}
