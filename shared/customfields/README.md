# Usage

Lưu ý các bảng có sử dụng cơ chế metadata driven, thì phải tạo một bảng custom_fields.

Ví dụ bảng nghiệp vụ có cột custom_fields

```postgres
ALTER TABLE products
ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;
```

Index GIN cho tìm kiếm động

```postgres
CREATE INDEX IF NOT EXISTS idx_products_custom_gin ON products USING GIN (custom_fields);
```

Index biểu thức cho key hay lọc nhiều (ví dụ: color)

```postgres
CREATE INDEX IF NOT EXISTS idx_products_cf_color ON products ((custom_fields->>'color'));
```

- Create

```go
func (r *ProductRepository) Create(ctx context.Context, coreName string, corePrice *float64, custom map[string]any) (*generated.Product, error) {
    // 1) Validate custom theo metadata
    vr, err := r.cfMgr.Validate(ctx, "products", custom, false)
    if err != nil { return nil, err }
    if len(vr.Errs) > 0 { return nil, fmt.Errorf("validation errors: %v", vr.Errs) }

    // 2) Lưu
    return r.db.Product.
        Create().
        SetName(coreName).
        SetNillablePrice(corePrice).
        SetCustomFields(vr.Clean).
        Save(ctx)
}
```

- Update

```go
func (r *ProductRepository) Patch(ctx context.Context, id int, patch map[string]any) (*generated.Product, error) {
    // Lấy hiện trạng custom
    cur, err := r.db.Product.Get(ctx, id)
    if err != nil { return nil, err }

    // Validate chỉ các field gửi lên
    vr, err := r.cfMgr.Validate(ctx, "products", patch, true)
    if err != nil { return nil, err }
    if len(vr.Errs) > 0 { return nil, fmt.Errorf("validation errors: %v", vr.Errs) }

    merged := customfields.MergePatch(cur.CustomFields, vr.Clean)

    return r.db.Product.
        UpdateOneID(id).
        SetCustomFields(merged).
        Save(ctx)
}
```

- List & filter custom

```go
func (r *ProductRepository) List(ctx context.Context, customFilters map[string]any, limit, offset int) ([]*generated.Product, error) {
    q := r.db.Product.Query()

    // Ent predicate ví dụ:
    if v, ok := customFilters["color"]; ok {
        q = q.Where(customfields.JSONEq("color", v))
        delete(customFilters, "color")
    }
    if v, ok := customFilters["min_weight"]; ok {
        q = q.Where(customfields.JSONNumOp("weight", ">=", v))
        delete(customFilters, "min_weight")
    }
    if v, ok := customFilters["kw"]; ok {
        kw := fmt.Sprintf("%%%v%%", v)
        q = q.Where(customfields.JSONILike("title", kw))
        delete(customFilters, "kw")
    }

    return q.
        Limit(limit).
        Offset(offset).
        All(ctx)
}
```
