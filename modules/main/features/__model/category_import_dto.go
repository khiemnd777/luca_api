package model

type CategoryExcelRow struct {
	LV1    string `json:"lv1"`
	LV2    string `json:"lv2"`
	LV3    string `json:"lv3"`
	Field1 string `json:"field1,omitempty"`
	Field2 string `json:"field2,omitempty"`
	Field3 string `json:"field3,omitempty"`
	Field4 string `json:"field4,omitempty"`
	Field5 string `json:"field5,omitempty"`
}

type CategoryImportRowResult struct {
	RowIndex int    `json:"row_index"`
	Status   string `json:"status"`
	Reason   string `json:"reason,omitempty"`
}

type CategoryImportResult struct {
	TotalRows int                       `json:"totalRows"`
	AddedLV1  int                       `json:"addedLV1"`
	AddedLV2  int                       `json:"addedLV2"`
	AddedLV3  int                       `json:"addedLV3"`
	Skipped   int                       `json:"skipped"`
	Errors    []string                  `json:"errors,omitempty"`
	Rows      []CategoryImportRowResult `json:"rows,omitempty"`
}
