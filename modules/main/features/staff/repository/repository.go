package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/role"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/section"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/staff"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/staffsection"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
	"golang.org/x/crypto/bcrypt"
)

type StaffRepository interface {
	Create(ctx context.Context, input model.StaffDTO) (*model.StaffDTO, error)
	Update(ctx context.Context, input model.StaffDTO) (*model.StaffDTO, error)
	ChangePassword(ctx context.Context, id int, newPassword string) error
	GetByID(ctx context.Context, id int) (*model.StaffDTO, error)
	CheckPhoneExists(ctx context.Context, userID int, phone string) (bool, error)
	CheckEmailExists(ctx context.Context, userID int, email string) (bool, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.StaffDTO], error)
	ListBySectionID(ctx context.Context, sectionID int, query table.TableQuery) (table.TableListResult[model.StaffDTO], error)
	ListByRoleName(ctx context.Context, roleName string, query table.TableQuery) (table.TableListResult[model.StaffDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.StaffDTO], error)
	Delete(ctx context.Context, id int) error
}

type staffRepo struct {
	db    *generated.Client
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewStaffRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) StaffRepository {
	return &staffRepo{db: db, deps: deps, cfMgr: cfMgr}
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

	refCode := uuid.NewString()
	qrCode := utils.GenerateQRCodeStringForUser(refCode)
	pwdHash, _ := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)

	userEnt, err := tx.User.Create().
		SetName(input.Name).
		SetPassword(string(pwdHash)).
		SetNillableEmail(&input.Email).
		SetNillablePhone(&input.Phone).
		SetNillableActive(&input.Active).
		SetNillableAvatar(&input.Avatar).
		SetNillableRefCode(&refCode).
		SetNillableQrCode(&qrCode).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	staffQ := tx.Staff.Create().
		SetUserID(userEnt.ID)

	// customfields
	_, err = customfields.PrepareCustomFields(ctx,
		r.cfMgr,
		[]string{"staff"},
		input.CustomFields,
		staffQ,
		false,
	)
	if err != nil {
		return nil, err
	}

	staffEnt, err := staffQ.Save(ctx)

	if err != nil {
		return nil, err
	}

	// Edge – Sections
	var sectionNames []string
	var sectionNamesStr string

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

			// get section names
			rows := make([]struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}, 0, len(sectionIDs))

			if err := tx.Section.
				Query().
				Where(section.IDIn(sectionIDs...)).
				Select(section.FieldID, section.FieldName).
				Scan(ctx, &rows); err != nil {
				return nil, err
			}

			// map id -> name
			nameByID := make(map[int]string, len(rows))
			for _, r := range rows {
				nameByID[r.ID] = r.Name
			}

			sectionNames = make([]string, 0, len(sectionIDs))
			for _, id := range sectionIDs {
				if n, ok := nameByID[id]; ok {
					sectionNames = append(sectionNames, n)
				}
			}

			sectionNamesStr = strings.Join(sectionNames, "|")
		}
	}

	_, err = tx.Staff.UpdateOneID(staffEnt.ID).
		SetNillableSectionNames(&sectionNamesStr).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	// Edge – Roles
	if input.RoleIDs != nil {
		roleIDs := utils.DedupInt(input.RoleIDs, -1)
		if len(roleIDs) > 0 {
			_, err = tx.User.UpdateOneID(userEnt.ID).
				AddRoleIDs(roleIDs...).
				Save(ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	dto := mapper.MapAs[*generated.User, *model.StaffDTO](userEnt)
	dto.SectionIDs = input.SectionIDs
	dto.SectionNames = sectionNames
	dto.RoleIDs = input.RoleIDs
	dto.CustomFields = input.CustomFields

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

	userQ := tx.User.UpdateOneID(input.ID).
		SetName(input.Name).
		SetNillableEmail(&input.Email).
		SetNillablePhone(&input.Phone).
		SetNillableActive(&input.Active).
		SetNillableAvatar(&input.Avatar)

	if input.Password != nil && *input.Password != "" {
		pwdHash, _ := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		userQ.SetPassword(string(pwdHash))
	}

	userEnt, err := userQ.Save(ctx)

	if err != nil {
		return nil, err
	}

	staffEnt, err := tx.Staff.
		Query().
		Where(staff.HasUserWith(user.IDEQ(input.ID))).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	var sectionNamesStr string
	var sectionNames []string

	// Edge – Sections
	if input.SectionIDs != nil {
		sectionIDs := utils.DedupInt(input.SectionIDs, -1)

		if _, err = tx.StaffSection.
			Delete().
			Where(staffsection.StaffIDEQ(staffEnt.ID)).
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

			// get section names
			rows := make([]struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}, 0, len(sectionIDs))

			if err := tx.Section.
				Query().
				Where(section.IDIn(sectionIDs...)).
				Select(section.FieldID, section.FieldName).
				Scan(ctx, &rows); err != nil {
				return nil, err
			}

			// map id -> name
			nameByID := make(map[int]string, len(rows))
			for _, r := range rows {
				nameByID[r.ID] = r.Name
			}

			sectionNames = make([]string, 0, len(sectionIDs))
			for _, id := range sectionIDs {
				if n, ok := nameByID[id]; ok {
					sectionNames = append(sectionNames, n)
				}
			}

			sectionNamesStr = strings.Join(sectionNames, "|")
		}
	}

	staffQ := tx.Staff.UpdateOneID(staffEnt.ID).
		SetNillableSectionNames(&sectionNamesStr)

	// customfields
	_, err = customfields.PrepareCustomFields(ctx,
		r.cfMgr,
		[]string{"staff"},
		input.CustomFields,
		staffQ,
		false,
	)
	if err != nil {
		return nil, err
	}

	_, err = staffQ.Save(ctx)

	if err != nil {
		return nil, err
	}

	// Edge – Roles
	if input.RoleIDs != nil {
		roleIDs := utils.DedupInt(input.RoleIDs, -1)

		upd := tx.User.UpdateOneID(userEnt.ID).ClearRoles()
		if len(roleIDs) > 0 {
			upd = upd.AddRoleIDs(roleIDs...)
		}
		if _, err = upd.Save(ctx); err != nil {
			return nil, err
		}
	}

	dto := mapper.MapAs[*generated.User, *model.StaffDTO](userEnt)
	dto.SectionIDs = input.SectionIDs
	dto.SectionNames = sectionNames
	dto.RoleIDs = input.RoleIDs
	dto.CustomFields = input.CustomFields

	return dto, nil
}

func (r *staffRepo) ChangePassword(ctx context.Context, id int, newPassword string) error {
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	const updateQuery = `UPDATE users SET password = $2 WHERE id = $1`
	_, err = r.deps.DB.ExecContext(ctx, updateQuery, id, string(newHash))
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (r *staffRepo) CheckPhoneExists(ctx context.Context, userID int, phone string) (bool, error) {
	return r.db.User.Query().
		Where(user.IDNEQ(userID), user.PhoneEQ(phone), user.DeletedAtIsNil()).
		Exist(ctx)
}

func (r *staffRepo) CheckEmailExists(ctx context.Context, userID int, email string) (bool, error) {
	return r.db.User.Query().
		Where(user.IDNEQ(userID), user.EmailEQ(email), user.DeletedAtIsNil()).
		Exist(ctx)
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
		Select(staffsection.FieldSectionID).
		Ints(ctx)
	if err != nil {
		return nil, err
	}

	roleIDs, err := userEnt.QueryRoles().IDs(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.User, *model.StaffDTO](userEnt)
	dto.SectionIDs = sectionIDs
	dto.RoleIDs = roleIDs

	if staffEnt.SectionNames != nil {
		sn := staffEnt.SectionNames
		sectionNames := strings.Split(*sn, "|")
		dto.SectionNames = sectionNames
	}

	if staffEnt.CustomFields != nil {
		dto.CustomFields = staffEnt.CustomFields
	}

	return dto, nil
}

func (r *staffRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.StaffDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.User.Query().
			Where(
				user.DeletedAtIsNil(),
				user.HasStaff(),
			).
			WithStaff(func(sq *generated.StaffQuery) {
				sq.WithSections(func(ssq *generated.StaffSectionQuery) {
					ssq.WithSection()
				})
			}),
		query,
		user.Table,
		user.FieldID,
		user.FieldID,
		func(src []*generated.User) []*model.StaffDTO {
			out := make([]*model.StaffDTO, 0, len(src))
			for _, u := range src {
				dto := mapper.MapAs[*generated.User, *model.StaffDTO](u)
				if u.Edges.Staff != nil {
					st := u.Edges.Staff

					for _, ss := range st.Edges.Sections {
						if ss.Edges.Section != nil {
							dto.SectionIDs = append(dto.SectionIDs, ss.SectionID)
							dto.SectionNames = append(dto.SectionNames, ss.Edges.Section.Name)
						}
					}

					// customfields
					dto.CustomFields = st.CustomFields
				}
				out = append(out, dto)
			}
			return out
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
		).
		WithStaff(func(sq *generated.StaffQuery) {
			sq.WithSections(func(ssq *generated.StaffSectionQuery) {
				ssq.WithSection()
			})
		})

	return table.TableList(
		ctx,
		q,
		query,
		user.Table,
		user.FieldID,
		user.FieldID,
		func(src []*generated.User) []*model.StaffDTO {
			out := make([]*model.StaffDTO, 0, len(src))
			for _, u := range src {
				dto := mapper.MapAs[*generated.User, *model.StaffDTO](u)
				if u.Edges.Staff != nil {
					st := u.Edges.Staff

					for _, ss := range st.Edges.Sections {
						if ss.Edges.Section != nil {
							dto.SectionIDs = append(dto.SectionIDs, ss.SectionID)
							dto.SectionNames = append(dto.SectionNames, ss.Edges.Section.Name)
						}
					}

					// customfields
					dto.CustomFields = st.CustomFields
				}
				out = append(out, dto)
			}
			return out
		},
	)
}

func (r *staffRepo) ListByRoleName(ctx context.Context, roleName string, query table.TableQuery) (table.TableListResult[model.StaffDTO], error) {
	q := r.db.User.
		Query().
		Where(
			user.DeletedAtIsNil(),
			user.HasRolesWith(role.RoleNameEQ(roleName)),
		).
		WithStaff(func(sq *generated.StaffQuery) {
			sq.WithSections(func(ssq *generated.StaffSectionQuery) {
				ssq.WithSection()
			})
		})

	return table.TableList(
		ctx,
		q,
		query,
		user.Table,
		user.FieldID,
		user.FieldID,
		func(src []*generated.User) []*model.StaffDTO {
			out := make([]*model.StaffDTO, 0, len(src))
			for _, u := range src {
				dto := mapper.MapAs[*generated.User, *model.StaffDTO](u)
				if u.Edges.Staff != nil {
					st := u.Edges.Staff

					for _, ss := range st.Edges.Sections {
						if ss.Edges.Section != nil {
							dto.SectionIDs = append(dto.SectionIDs, ss.SectionID)
							dto.SectionNames = append(dto.SectionNames, ss.Edges.Section.Name)
						}
					}

					// customfields
					dto.CustomFields = st.CustomFields
				}
				out = append(out, dto)
			}
			return out
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
