package customfields

type FieldType string

const (
	TypeText        FieldType = "text"
	TypeNumber      FieldType = "number"
	TypeBool        FieldType = "bool"
	TypeDate        FieldType = "date" // ISO-8601 (YYYY-MM-DD) hoặc RFC3339 date-time tùy options
	TypeSelect      FieldType = "select"
	TypeMultiSelect FieldType = "multiselect"
	TypeJSON        FieldType = "json"
	TypeRichText    FieldType = "richtext"
	TypeImage       FieldType = "image"
	TypeRelation    FieldType = "relation" // giữ id / ids trong custom_fields
)

type FieldDef struct {
	Name         string         `json:"name"`
	Label        string         `json:"label"`
	Type         FieldType      `json:"type"`
	Required     bool           `json:"required"`
	Unique       bool           `json:"unique"`
	Table        bool           `json:"table"`
	Form         bool           `json:"form"`
	Search       bool           `json:"search"`
	DefaultValue any            `json:"default_value"`
	Options      map[string]any `json:"options"`    // choices/min/max/pattern/...
	Visibility   string         `json:"visibility"` // public/admin/internal
}

type Schema struct {
	Collection string     `json:"collection"`
	Fields     []FieldDef `json:"fields"`
}
