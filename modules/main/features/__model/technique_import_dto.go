package model

type TechniqueExcelRow struct {
	CategoryName string `json:"category_name"`
	Name         string `json:"name"`
}

type TechniqueImportResult struct {
	TotalRows int      `json:"totalRows"`
	Added     int      `json:"added"`
	Skipped   int      `json:"skipped"`
	Errors    []string `json:"errors,omitempty"`
}
