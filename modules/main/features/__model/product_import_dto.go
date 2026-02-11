package model

type ProductExcelRow struct {
	Code string `json:"code"`
	Name string `json:"name"`

	CategoryLV1 string `json:"category_lv1"`
	CategoryLV2 string `json:"category_lv2"`
	CategoryLV3 string `json:"category_lv3"`

	BrandNames           string `json:"brand_names,omitempty"`
	RawMaterialNames     string `json:"raw_material_names,omitempty"`
	TechniqueNames       string `json:"technique_names,omitempty"`
	RestorationTypeNames string `json:"restoration_type_names,omitempty"`
	ProcessNames         string `json:"process_names,omitempty"`
	RetailPrice          string `json:"retail_price,omitempty"`

	HasBrandField           bool `json:"-"`
	HasRawMaterialField     bool `json:"-"`
	HasTechniqueField       bool `json:"-"`
	HasRestorationTypeField bool `json:"-"`
	HasProcessField         bool `json:"-"`
}

type ProductImportResult struct {
	TotalRows int      `json:"totalRows"`
	Added     int      `json:"added"`
	Updated   int      `json:"updated"`
	Skipped   int      `json:"skipped"`
	Errors    []string `json:"errors,omitempty"`
}
