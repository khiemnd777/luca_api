package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/modules/auth_apple/config"
	"github.com/khiemnd777/andy_api/modules/auth_apple/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/user"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type AuthAppleRepository struct {
	ent  *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewAuthAppleRepository(ent *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) *AuthAppleRepository {
	return &AuthAppleRepository{
		ent:  ent,
		deps: deps,
	}
}

// FindOrCreateUserByApple tìm theo provider_id (sub). Nếu chưa có thì tạo user mới.
//   - Apple có thể KHÔNG trả email ở lần đăng nhập sau → dùng fallback email để đảm bảo unique & not-null.
//   - Nếu email bị unique conflict (email đã tồn tại do đăng ký trước đó), ta trả về user theo email.
//     (Không ép link/ghi đè provider để tránh làm mất liên kết Google/Facebook hiện có.)
func (r *AuthAppleRepository) FindOrCreateUserByApple(ctx context.Context, info *model.AppleUserInfo) (*generated.User, bool, error) {
	// 1) Nếu đã có provider_id=Apple Sub → dùng luôn
	if u, err := r.ent.User.Query().
		Where(user.ProviderIDEQ(info.Sub), user.ProviderEQ("apple"), user.Active(true)).
		First(ctx); err == nil {
		return u, false, nil
	}

	// 2) Chuẩn hóa email: nếu Apple không trả, tạo fallback email để không vi phạm unique/not-null
	email := info.Email
	if email == "" {
		domain := "apple.honvang.com"
		if r.deps != nil && r.deps.Config.AuthApple.FallbackEmailDomain != "" {
			domain = r.deps.Config.AuthApple.FallbackEmailDomain
		}
		email = fmt.Sprintf("%s@%s", info.Sub, domain)
	}

	dummyPassword := utils.GenerateOAuthDummyPassword()

	// 3) Cố gắng tạo user mới
	newUser, err := r.ent.User.Create().
		SetEmail(email).
		SetName(info.Name).
		SetAvatar(info.Picture).
		SetPassword(dummyPassword).
		SetProvider("apple").
		SetProviderID(info.Sub).
		Save(ctx)
	if err == nil {
		return newUser, true, nil
	}

	// 4) Nếu email đã tồn tại → trả về user theo email (không ép link provider)
	//    Bạn có thể mở comment "link an toàn" phía dưới nếu muốn tự động set provider apple
	//    khi tài khoản hiện tại chưa có provider hoặc provider đã là apple.
	existed, qerr := r.ent.User.Query().
		Where(user.EmailEQ(email), user.Active(true)).
		First(ctx)
	if qerr == nil {
		// --- (Tùy chọn) Link an toàn — chỉ set nếu chưa có provider hoặc đã là apple ---
		// if existed.Provider == "" || existed.Provider == "apple" {
		// 	if existed.ProviderID != info.Sub {
		// 		_, _ = existed.Update().SetProvider("apple").SetProviderID(info.Sub).Save(ctx)
		// 	}
		// }
		return existed, false, nil
	}

	// 5) Nếu không tìm được theo email, thử nhượng bộ: tìm user theo provider_id (không filter active)
	if u2, e2 := r.ent.User.Query().
		Where(user.ProviderIDEQ(info.Sub), user.ProviderEQ("apple")).
		First(ctx); e2 == nil {
		return u2, false, nil
	}

	// 6) Trả lỗi gốc tạo user
	return nil, false, err
}

func (r *AuthAppleRepository) CreateRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	_, err := r.ent.RefreshToken.Create().
		SetUserID(userID).
		SetToken(token).
		SetExpiresAt(expiresAt).
		Save(ctx)
	return err
}
