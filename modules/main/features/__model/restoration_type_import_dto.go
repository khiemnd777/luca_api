package model

type RestorationTypeExcelRow struct {
	CategoryName string `json:"category_name"`
	Name         string `json:"name"`
}

type RestorationTypeImportResult struct {
	TotalRows int      `json:"totalRows"`
	Added     int      `json:"added"`
	Skipped   int      `json:"skipped"`
	Errors    []string `json:"errors,omitempty"`
}
