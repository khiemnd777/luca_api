package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/staff"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/staffsection"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type StaffRepository interface {
	Create(ctx context.Context, input model.StaffDTO) (*model.StaffDTO, error)
	Update(ctx context.Context, input model.StaffDTO) (*model.StaffDTO, error)
	GetByID(ctx context.Context, id int) (*model.StaffDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.StaffDTO], error)
	ListBySectionID(ctx context.Context, sectionID int, query table.TableQuery) (table.TableListResult[model.StaffDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.StaffDTO], error)
	Delete(ctx context.Context, id int) error
}

type staffRepo struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewStaffRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) StaffRepository {
	return &staffRepo{db: db, deps: deps}
}

func (r *staffRepo) Create(ctx context.Context, input model.StaffDTO) (*model.StaffDTO, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	userEnt, err := tx.User.Create().
		SetName(input.Name).
		SetNillableEmail(input.Email).
		SetNillablePhone(input.Phone).
		SetNillableActive(input.Active).
		SetNillableAvatar(input.Avatar).
		SetNillableQrCode(input.QrCode).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	staffEnt, err := tx.Staff.Create().
		SetUserID(userEnt.ID).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	// Edge
	if input.SectionIDs != nil {
		sectionIDs := utils.DedupInt(input.SectionIDs, -1)
		if len(sectionIDs) > 0 {
			bulk := make([]*generated.StaffSectionCreate, 0, len(sectionIDs))
			for _, sid := range sectionIDs {
				bulk = append(bulk, tx.StaffSection.Create().
					SetStaffID(staffEnt.ID).
					SetSectionID(sid),
				)
			}
			if err = tx.StaffSection.CreateBulk(bulk...).Exec(ctx); err != nil {
				return nil, err
			}
		}
	}

	dto := mapper.MapAs[*generated.User, *model.StaffDTO](userEnt)
	dto.SectionIDs = input.SectionIDs
	return dto, nil
}

func (r *staffRepo) Update(ctx context.Context, input model.StaffDTO) (*model.StaffDTO, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	staffEnt, err := tx.Staff.
		Query().
		Where(staff.HasUserWith(user.IDEQ(input.ID))).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	userEnt, err := tx.User.UpdateOneID(staffEnt.ID).
		SetName(input.Name).
		SetNillableEmail(input.Email).
		SetNillablePhone(input.Phone).
		SetNillableActive(input.Active).
		SetNillableAvatar(input.Avatar).
		SetNillableQrCode(input.QrCode).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	if input.SectionIDs != nil {
		sectionIDs := utils.DedupInt(input.SectionIDs, -1)

		if _, err = tx.StaffSection.
			Delete().
			Where(staffsection.StaffIDEQ(input.ID)).
			Exec(ctx); err != nil {
			return nil, err
		}

		if len(sectionIDs) > 0 {
			bulk := make([]*generated.StaffSectionCreate, 0, len(sectionIDs))
			for _, sid := range sectionIDs {
				bulk = append(bulk, tx.StaffSection.Create().
					SetStaffID(staffEnt.ID).
					SetSectionID(sid),
				)
			}
			if err = tx.StaffSection.CreateBulk(bulk...).Exec(ctx); err != nil {
				return nil, err
			}
		}
	}

	dto := mapper.MapAs[*generated.User, *model.StaffDTO](userEnt)
	dto.SectionIDs = input.SectionIDs
	return dto, nil
}

func (r *staffRepo) GetByID(ctx context.Context, id int) (*model.StaffDTO, error) {
	userEnt, err := r.db.User.Query().
		Where(
			user.IDEQ(id),
			user.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	staffEnt, err := r.db.Staff.
		Query().
		Where(staff.HasUserWith(user.IDEQ(id))).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	sectionIDs, err := staffEnt.
		QuerySections().
		IDs(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.User, *model.StaffDTO](userEnt)
	dto.SectionIDs = sectionIDs
	return dto, nil
}

func (r *staffRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.StaffDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.User.Query().
			Where(user.DeletedAtIsNil()),
		query,
		staff.Table,
		staff.FieldID,
		staff.FieldID,
		func(src []*generated.User) []*model.StaffDTO {
			mapped := mapper.MapListAs[*generated.User, *model.StaffDTO](src)
			return mapped
		},
	)
	if err != nil {
		var zero table.TableListResult[model.StaffDTO]
		return zero, err
	}
	return list, nil
}

func (r *staffRepo) ListBySectionID(ctx context.Context, sectionID int, query table.TableQuery) (table.TableListResult[model.StaffDTO], error) {
	q := r.db.User.
		Query().
		Where(
			user.DeletedAtIsNil(),
			user.HasStaffWith(
				staff.HasSectionsWith(
					staffsection.SectionIDEQ(sectionID),
				),
			),
		)

	return table.TableList(
		ctx,
		q,
		query,
		user.Table,
		user.FieldID,
		user.FieldID,
		func(src []*generated.User) []*model.StaffDTO {
			return mapper.MapListAs[*generated.User, *model.StaffDTO](src)
		},
	)
}

func (r *staffRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.StaffDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.User.Query().
			Where(user.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(user.FieldName),
			dbutils.GetNormField(user.FieldPhone),
			dbutils.GetNormField(user.FieldEmail),
		},
		query,
		user.Table,
		user.FieldID,
		user.FieldID,
		user.Or,
		func(src []*generated.User) []*model.StaffDTO {
			mapped := mapper.MapListAs[*generated.User, *model.StaffDTO](src)
			return mapped
		},
	)
}

func (r *staffRepo) Delete(ctx context.Context, id int) error {
	return r.db.User.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
