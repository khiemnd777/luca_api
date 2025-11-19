package relation

type Config struct {
	// Schema
	MainTable string // ví dụ: "materials"
	RefTable  string // ví dụ: "suppliers"

	// Runtime: cách lấy ID từ entity + input
	// entity: struct chứa ID chính (vd: material)
	// input:  struct chứa danh sách IDs (vd: SupplierIDs)
	GetMainID func(entity any) (int, error)
	GetIDs    func(input any) ([]int, error)
	SetResult func(output any, res []string) error
}
