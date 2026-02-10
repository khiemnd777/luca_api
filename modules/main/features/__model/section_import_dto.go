package model

type SectionExcelRow struct {
	DepartmentID int    `json:"department_id"`
	Name         string `json:"name"`
	Color        string `json:"color"`
	ProcessName  string `json:"process_name"`
}

type SectionImportResult struct {
	TotalRows int      `json:"totalRows"`
	Added     int      `json:"added"`
	Skipped   int      `json:"skipped"`
	Errors    []string `json:"errors,omitempty"`
}
