# 📦 CI/CD TOÀN TẬP: TỪ DRONE CI TRÊN SERVER LINUX ĐẾN DEPLOY FIREBASE (GITHUB REPO)

## 🧰 MÔI TRƯỜNG SỬ DỤNG

- ✅ Local: Windows/macOS
- ✅ Server: Linux (Ubuntu)
- ✅ Git provider: GitHub
- ✅ CI/CD: Drone CI (Self-hosted)
- ✅ Deploy: Firebase Hosting hoặc Firebase App Distribution

---

## 🚀 PHẦN 1: CÀI ĐẶT DRONE CI TRÊN SERVER LINUX

### 1. Cài Docker và Docker Compose
```bash
sudo apt update
sudo apt install docker.io docker-compose -y
```

### 2. Tạo cấu hình Drone (ở `/opt/drone/docker-compose.yml`)
```yaml
version: '3'
services:
  drone-server:
    image: drone/drone:2
    ports:
      - 8080:80
    volumes:
      - ./data:/data
    restart: always
    environment:
      - DRONE_GITHUB_CLIENT_ID=<your-client-id>
      - DRONE_GITHUB_CLIENT_SECRET=<your-client-secret>
      - DRONE_RPC_SECRET=supersecret123
      - DRONE_SERVER_HOST=ci.yourdomain.com
      - DRONE_SERVER_PROTO=http
      - DRONE_USER_CREATE=username:yourgithubusername,admin:true

  drone-runner:
    image: drone/drone-runner-docker:1
    restart: always
    depends_on:
      - drone-server
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DRONE_RPC_PROTO=http
      - DRONE_RPC_HOST=drone-server
      - DRONE_RPC_SECRET=supersecret123
      - DRONE_RUNNER_CAPACITY=2
```

> Đăng ký GitHub OAuth App: https://github.com/settings/developers
> - Homepage: `http://ci.yourdomain.com`
> - Authorization callback URL: `http://ci.yourdomain.com/login`

### 3. Chạy Drone
```bash
cd /opt/drone
sudo docker-compose up -d
```

---

## 🔗 PHẦN 2: KẾT NỐI GITHUB VỚI DRONE

1. Mở `http://<your-server-ip>:8080`
2. Login bằng GitHub OAuth
3. Activate repository bạn muốn CI/CD

---

## 📄 PHẦN 3: THÊM FILE `.drone.yml` VÀO CODEBASE

### Option 1: Deploy Firebase Hosting (Flutter Web)
```yaml
kind: pipeline
type: docker
name: firebase-deploy

steps:
  - name: build flutter web
    image: cirrusci/flutter:3.19
    commands:
      - flutter pub get
      - flutter build web

  - name: deploy to firebase hosting
    image: node:18
    environment:
      FIREBASE_TOKEN:
        from_secret: firebase_token
    commands:
      - npm install -g firebase-tools
      - firebase deploy --only hosting --token $FIREBASE_TOKEN

trigger:
  branch:
    - main
```

### Option 2: Deploy Firebase App Distribution (Android APK)
```yaml
kind: pipeline
type: docker
name: firebase-app-distribution

steps:
  - name: build apk
    image: cirrusci/flutter:3.19
    commands:
      - flutter pub get
      - flutter build apk --release

  - name: upload to Firebase App Distribution
    image: node:18
    environment:
      FIREBASE_TOKEN:
        from_secret: firebase_token
    commands:
      - npm install -g firebase-tools
      - firebase appdistribution:distribute build/app/outputs/flutter-apk/app-release.apk \
          --app "<your-firebase-app-id>" \
          --groups "testers" \
          --token $FIREBASE_TOKEN

trigger:
  branch:
    - main
```

---

## 🔐 PHẦN 4: CẤU HÌNH FIREBASE

### 1. Cài Firebase CLI trên local (để tạo token)
```bash
npm install -g firebase-tools
firebase login:ci
```
Copy `FIREBASE_TOKEN`, rồi lên Drone UI → Repo → Settings → Secrets → Thêm:
```
firebase_token = <TOKEN_FROM_FIREBASE>
```

### 2. Tạo `firebase.json`
```json
{
  "hosting": {
    "public": "build/web",
    "ignore": [
      "firebase.json",
      "**/.*",
      "**/node_modules/**"
    ]
  }
}
```

### 3. Tạo `.firebaserc`
```json
{
  "projects": {
    "default": "your-firebase-project-id"
  }
}
```

---

## 🎯 PHẦN 5: CI/CD WORKFLOW

```bash
# Trên local Windows/macOS:
git add .
git commit -m "feat: update build"
git push

# Ngay sau đó:
Drone CI sẽ tự động:
→ Build → Deploy Firebase 🎉
```

---

## ✅ CHECKLIST NHANH

| Thành phần | Đã cài | Notes |
|------------|--------|-------|
| Docker     | ✅     | `apt install docker.io` |
| Docker Compose | ✅ | `apt install docker-compose` |
| Drone Server | ✅ | Qua `docker-compose.yml` |
| Drone Runner | ✅ | Cùng file compose |
| GitHub OAuth App | ✅ | Tạo app tại GitHub settings |
| Firebase CLI | ✅ | `npm install -g firebase-tools` |
| Firebase Token | ✅ | `firebase login:ci` |
| `.drone.yml` | ✅ | Commit vào repo |
| `firebase.json` | ✅ | Khai báo thư mục build |

---

> Bất kỳ khi nào cần build web, apk hoặc deploy production, chỉ cần `git push` là Drone tự chạy ✨

---

**Maintainer**: khiemnd777  
**Last update**: 2025-04-14

