package model

type MappedCell struct {
	InternalKind  string `json:"internal_kind"`  // core | metadata | external
	InternalPath  string `json:"internal_path"`  // name, phone_number, tax_code...
	InternalLabel string `json:"internal_label"` // để debug/log
	DataType      string `json:"data_type"`      // text, number, date...

	// Giá trị raw (read từ Excel)
	RawValue any `json:"raw_value"`

	// Giá trị sau khi parse/transform
	Value any `json:"value"`
}

type MappedRow struct {
	ProfileID int `json:"profile_id"`

	// core fields → gán thẳng vào struct entity
	CoreFields map[string]any `json:"core_fields"`

	// metadata → gán vào custom_fields
	MetadataFields map[string]any `json:"metadata_fields"`

	// external → field ảo: log, note...
	ExternalFields map[string]any `json:"external_fields"`

	// detail đầy đủ từng cell (để debug, hiển thị lỗi,...)
	Cells []MappedCell `json:"cells"`
}
