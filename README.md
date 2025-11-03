# 🏗️ Backend - Hệ Thống Quản Lý Sản Xuất Luca

## 🔧 Công nghệ sử dụng
- **Ngôn ngữ:** Golang
- **Kiến trúc:** Modular Monolith (mỗi module có thể chạy độc lập hoặc tích hợp với API Gateway)
- **Database:** PostgreSQL + Ent (ORM) + Flyway (migration)
- **Redis:** Cache, Pub/Sub, Circuit Breaker Backup
- **Authentication:** JWT + Refresh Token
- **Authorization:** RBAC + Permission-based Access
- **Circuit Breaker:** sony/gobreaker
- **Logger:** zap custom
- **Cấu hình:** YAML
- **CI/CD:** Drone CI (dev: Windows, server: Linux)

## 📁 Cấu trúc thư mục
```
- modules/
  - product/
    - main.go
    - config.yaml
    - handler/
    - service/
    - repository/
    - ent/
- shared/
  - app/
  - config/
  - redis/
  - logger/
  - db/ent/
- scripts/
  - module_manager.go
  - init_roles/
```

## 🚀 Bắt đầu sau khi clone repo

Sau khi clone lần đầu, hãy chạy lệnh sau để:

- Cài đặt dependencies (`go mod tidy`)  
- Generate toàn bộ schema Ent  
- Build thử toàn bộ module để kiểm tra lỗi

```bash
go run ./scripts/init_project.go
```

> Lệnh này sẽ giúp bạn đảm bảo project hoạt động đúng ngay từ đầu, tránh lỗi thiếu thư mục `generated` hoặc Ent chưa được generate.

## ▶️ Khởi chạy module
```bash
go run ./main.go
```

## 🧪 Migration
- Ent generate schema tại `shared/db/ent/schema`
- Flyway migration scripts tại `migrations/<module>`

## 🛡️ Bảo mật
- Hỗ trợ Access Token ngắn hạn và Refresh Token dài hạn
- Middleware kiểm tra Role & Permission tự động

## 🧰 Giao tiếp module
- HTTP API
- Event Bus
- Dependency Injection