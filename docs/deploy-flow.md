# Flow: từ tạo service đến chạy trên CapRover (monorepo)

`test-codex-env` (`swe-squad/test-codex-env`) là **monorepo**:

- Mỗi service nằm trong một thư mục `services/<tên>/`.
- Tất cả service dùng chung bộ workflow ở `.github/workflows/`.
- Mỗi lần push hay mở PR, **chỉ những service có thay đổi** mới chạy CI và deploy.

Guide này viết cho server CapRover hiện tại, lấy theo bản backup `captain.json`. Ví dụ là 2 service đang có trong repo:

| Service | Template | Endpoint kiểm tra | App CapRover (dev) | Port |
|---|---|---|---|---|
| `ping-api` | go-gin | `GET /api/v1/ping` | `ping-api-dev` | 8080 |
| `ping-web` | react-vite | `GET /ping` | `ping-web-dev` | 80 |

Mặc định mỗi service **chỉ có app dev**. Production là tuỳ chọn, xem [Thêm Production](#thêm-production-cho-một-service).

## Server CapRover hiện tại

| Mục | Giá trị |
|---|---|
| Dashboard / `CAPROVER_SERVER` | `https://captain.cr.ngochoanielts.com` |
| URL mặc định của app | `https://<app>.cr.ngochoanielts.com` |
| Registry GHCR | `ghcr.io` (registry mặc định), hiện dùng PAT của `ielts-ngoc-hoan` |
| Gọi nội bộ giữa các app | `srv-captain--<app>:<port>`, ví dụ `srv-captain--ielts-db:5432` |

> ⚠️ **Bảo mật:**
>
> - `captain.json` chứa secret thật và đã được thêm vào `.gitignore`. **Không commit file này.**
> - `db-test` và `redis-test` đang mở port `54322` và `63791` ra Internet.

---

## Cách mọi thứ nối với nhau

```mermaid
flowchart LR
    subgraph REPO["swe-squad/test-codex-env"]
        S1[services/ping-api]
        S2[services/ping-web]
        WF[.github/workflows]
    end
    subgraph GH["GitHub"]
        GA[Actions]
        E1[(env ping-api-develop<br/>3 secrets)]
        E2[(env ping-web-develop<br/>3 secrets + VITE_API_BASE_URL)]
    end
    GHCR[(ghcr.io/swe-squad/ping-api<br/>ghcr.io/swe-squad/ping-web)]
    subgraph CR["CapRover"]
        A1[ping-api-dev]
        A2[ping-web-dev]
    end

    REPO -- git push develop --> GA
    E1 -. secrets .-> GA
    E2 -. secrets .-> GA
    GA -- build + push --> GHCR
    GA -- "API deploy (App Token)" --> CR
    GHCR -- pull --> A1
    GHCR -- pull --> A2
```

- **CapRover không tự đọc GitHub.** GitHub Actions là bên gọi API của CapRover, dùng 3 secrets lưu trong GitHub environment của từng service:
  - `CAPROVER_SERVER`: gọi vào server nào.
  - `APP_NAME`: app nào.
  - `CAPROVER_TOKEN`: token chứng minh có quyền deploy.
- **Tên environment theo service:** mỗi service có environment riêng, dạng `<service>-develop`, ví dụ `ping-api-develop`. Vì vậy mỗi service trỏ được tới app riêng của nó.
- **CapRover pull image từ GHCR** bằng PAT đã cấu hình trong registry.

---

## Checklist làm 1 lần

### A. Registry GHCR (một lần cho cả server)

Registry hiện tại dùng PAT của `ielts-ngoc-hoan`, nhiều khả năng không đọc được package của `swe-squad`. Tài khoản `minhvip08` là admin của cả 2 org, nên dùng PAT của tài khoản này:

- [ ] **Tạo PAT:** mở https://github.com/settings/tokens/new?scopes=read:packages&description=caprover-ghcr
  - Loại classic, chỉ tick scope `read:packages`.
  - Chọn expiration dài và đặt nhắc lịch thay token. Token hết hạn thì **mọi** app sẽ không pull được image.
- [ ] **Kiểm tra token** (chạy trong terminal của bạn):

  ```bash
  read -s PAT
  GH_TOKEN=$PAT gh api "orgs/ielts-ngoc-hoan/packages?package_type=container" -q '.[].name'   # thấy ielts-ngoc-hoan-be/fe
  GH_TOKEN=$PAT gh api "orgs/swe-squad/packages?package_type=container" -q '.[].name'         # rỗng là được, miễn không 403
  ```

- [ ] **Cập nhật registry:** vào CapRover → **Cluster → Docker Registry Configuration**, sửa registry `ghcr.io`:
  - Username: `minhvip08`
  - Password: PAT vừa tạo

  Giữ registry này là default, và chỉ để **một** registry cho `ghcr.io`.
- [ ] **Kiểm tra app cũ:** redeploy thử app `ielts-api`, để chắc image cũ vẫn pull được.

### B. CapRover: tạo app dev cho mỗi service

Mở `https://captain.cr.ngochoanielts.com` → **Apps** → tạo `ping-api-dev` và `ping-web-dev`. Với **mỗi app**:

- [ ] **HTTP Settings:** bấm **Enable HTTPS**, rồi bật **Force HTTPS**.
- [ ] **App Configs:** đặt **Container HTTP Port** bằng giá trị `port` trong `service.json` của service đó:
  - `ping-api-dev` → `8080`
  - `ping-web-dev` → `80`

  Bấm **Save & Update**. Env vars runtime (DB, key...) cũng thêm ở đây.
- [ ] **Deployment:** bật **App Token** rồi copy token. Không lưu token vào file.

### C. GitHub: tạo environment và secrets cho mỗi service

Chạy trong terminal của bạn, từ root của repo. Gõ token trực tiếp, đừng dán vào chat:

```bash
cd /home/simi/workspace/github/agentic/test-codex-env

CAPROVER_SERVER=https://captain.cr.ngochoanielts.com \
APP_NAME_DEVELOP=ping-api-dev CAPROVER_TOKEN_DEVELOP=<token ping-api-dev> \
  scripts/setup-service.sh ping-api

CAPROVER_SERVER=https://captain.cr.ngochoanielts.com \
APP_NAME_DEVELOP=ping-web-dev CAPROVER_TOKEN_DEVELOP=<token ping-web-dev> \
VITE_API_BASE_URL_DEVELOP=https://ping-api-dev.cr.ngochoanielts.com \
  scripts/setup-service.sh ping-web
```

- [ ] Script tạo environment `ping-api-develop` và `ping-web-develop` trên repo `swe-squad/test-codex-env`, và điền secrets vào đó.

**Muốn làm tay trên web thay vì chạy script:** vào repo → **Settings → Environments → New environment**.

1. Đặt tên `ping-api-develop`. Thêm 3 secrets:

   | Secret | Giá trị |
   |---|---|
   | `CAPROVER_SERVER` | `https://captain.cr.ngochoanielts.com` |
   | `APP_NAME` | `ping-api-dev` |
   | `CAPROVER_TOKEN` | App Token của `ping-api-dev` |

2. Làm tương tự cho `ping-web-develop`, thêm biến (tab **Variables**): `VITE_API_BASE_URL` = `https://ping-api-dev.cr.ngochoanielts.com`.

### D. Push code (lần đầu)

Repo hiện chưa có commit nào. Lần push đầu sẽ deploy **mọi** service:

```bash
cd /home/simi/workspace/github/agentic/test-codex-env
git add -A && git commit -m "feat: monorepo with ping-api and ping-web"
git push -u origin main
git checkout -b develop && git push -u origin develop
```

- [ ] **Push `main`:** chưa deploy service nào, vì chưa có service nào trong `PRODUCTION_SERVICES`.
- [ ] **Push `develop`:** deploy `ping-api` → `ping-api-dev` và `ping-web` → `ping-web-dev`.

### E. Kiểm tra

- [ ] **GitHub → Actions → Deploy** (lần chạy của `develop`): job `changes` liệt kê `["ping-api","ping-web"]`. Hai job `ping-api` và `ping-web` đều xanh.
- [ ] **CapRover → `ping-api-dev` → Deployment:** đang chạy `ghcr.io/swe-squad/ping-api:<sha7>`. Kiểm tra tương tự với `ping-web-dev`.
- [ ] **Terminal:**

  ```bash
  curl https://ping-api-dev.cr.ngochoanielts.com/api/v1/ping   # {"message":"pong","service":"ping-api"}
  curl https://ping-web-dev.cr.ngochoanielts.com/ping          # {"message":"pong","service":"ping-web"}
  ```

### F. AI review (một lần cho cả repo)

- [ ] **Codex settings → Code review:** bật `swe-squad/test-codex-env`.
  - Codex đọc `AGENTS.md` ở root (luật chung) và `services/<x>/AGENTS.md` (luật riêng của từng stack).

---

## Một lần deploy diễn ra thế nào

```mermaid
sequenceDiagram
    autonumber
    participant Dev as Bạn
    participant GA as GitHub Actions
    participant GHCR as GHCR
    participant CR as CapRover

    Dev->>GA: git push develop (sửa services/ping-api)
    GA->>GA: job changes: changed-services.sh → ["ping-api"]
    Note over GA: ping-web không đổi → không chạy
    GA->>GA: _ci.yml: setup Go từ service.json → make ci
    alt CI fail
        GA-->>Dev: ❌ dừng, không deploy
    end
    GA->>GA: docker build services/ping-api
    GA->>GHCR: push ghcr.io/swe-squad/ping-api:<sha7> + :latest-develop
    GA->>CR: deploy (secrets của env ping-api-develop)
    CR->>GHCR: pull image
    CR->>CR: start container mới, chuyển traffic, tắt container cũ
    GA-->>Dev: ✅ xanh
```

| Trigger | Service nào | Deploy lên | Image tags |
|---|---|---|---|
| Push / merge vào `develop` | Service có thay đổi | `<svc>-develop` | `<sha7>`, `latest-develop` |
| Push / merge vào `main` | Service có thay đổi **và** có tên trong biến `PRODUCTION_SERVICES` | `<svc>-production` | `<sha7>`, `latest-main` |
| Comment `/deploy` trên PR | Service có thay đổi trong PR | `<svc>-develop` | `<sha7>`, `pr-<số>` |
| Comment `/deploy ping-web` | Chỉ service được ghi tên | `<svc>-develop` | `<sha7>`, `pr-<số>` |
| Actions → *Deploy* → Run workflow | Danh sách nhập vào (để trống = tất cả) | `main` → production, nhánh khác → develop | `<sha7>`, `latest-<branch>` |

Khi nào một service được tính là "có thay đổi": có file thay đổi bên trong `services/<svc>/`. Thay đổi file ở root, như README hay docs, không deploy service nào.

## Thêm service mới

```
/new-service python-fastapi report-api       # Claude Code (hoặc $new-service ... trong Codex)
```

Hoặc chạy tay:

1. `scripts/new-service.sh python-fastapi report-api`: tạo `services/report-api/`.
2. CapRover: tạo app `report-api-dev` (port `8000`), bật HTTPS và App Token.
3. `CAPROVER_SERVER=... APP_NAME_DEVELOP=report-api-dev CAPROVER_TOKEN_DEVELOP=... scripts/setup-service.sh report-api`
4. Commit, mở PR vào `develop`, rồi merge. Lần merge này deploy `report-api`.

Không cần sửa workflow nào. Workflow tự nhận service mới nhờ file `service.json`.

## Thêm Production cho một service

1. **CapRover:** tạo app `ping-api` (không có `-dev`), cấu hình giống app dev. Thường sẽ gắn domain thật.
2. **Terminal:** chạy lại script với giá trị production:

   ```bash
   CAPROVER_SERVER=https://captain.cr.ngochoanielts.com \
   APP_NAME_DEVELOP=ping-api-dev    CAPROVER_TOKEN_DEVELOP=<token dev> \
   APP_NAME_PRODUCTION=ping-api     CAPROVER_TOKEN_PRODUCTION=<token prod> \
     scripts/setup-service.sh ping-api
   ```

   Script tạo environment `ping-api-production` và thêm `ping-api` vào repo variable `PRODUCTION_SERVICES`.
3. **GitHub:** vào Actions → *Deploy* → **Run workflow** trên `main`, ô services nhập `ping-api`, để deploy lần đầu. Từ đó trở đi, mỗi lần merge vào `main` có sửa `ping-api` là tự deploy production.

Muốn tắt production cho một service thì xoá tên service khỏi `PRODUCTION_SERVICES` trong **Settings → Variables**.

## Làm việc hằng ngày

```mermaid
flowchart LR
    A[git checkout -b feature/x] --> B[Sửa services/ping-api, mở PR vào develop]
    B --> C{CI cho ping-api + Codex review}
    C -- muốn thử --> D["comment /deploy → ping-api-dev"]
    C -- ok --> E[Merge → deploy ping-api-dev]
    D --> E
    E --> F[PR develop → main]
    F --> G["Merge → production (nếu có trong PRODUCTION_SERVICES)"]
```

- **PR:** chỉ service có thay đổi mới chạy CI. Trong tab Checks, job hiện tên theo service, ví dụ `ci / ping-api`.
- **Comment `/deploy`:**
  - Bot thả reaction 🚀, rồi comment ✅/❌ kèm danh sách service đã deploy.
  - App dev giữ bản của PR cho tới lần merge vào `develop` tiếp theo.
- **Rollback:**
  - Vào Actions → *Deploy* → Run workflow trên commit cũ, ô services nhập tên service.
  - Hoặc vào CapRover → app → **Deployment**, chọn version cũ rồi deploy lại.

## Lỗi thường gặp

| Triệu chứng | Nguyên nhân hay gặp | Cách sửa |
|---|---|---|
| Job `changes` ra `[]` nên không deploy gì | Push không đụng tới file nào trong `services/` | Bình thường. Muốn deploy thì vào Actions → Run workflow |
| Fail ở *Deploy image to CapRover* (auth/token) | Environment `<svc>-develop` thiếu hoặc sai secret | Kiểm tra **Settings → Environments**, rồi **Re-run** |
| CapRover báo `unauthorized`/`denied` khi pull | PAT của registry không đọc được `ghcr.io/swe-squad/*` | Làm lại mục A |
| Deploy xanh nhưng mở web báo **502** | Container HTTP Port khác `port` trong `service.json` | Sửa trong App Configs |
| Lỗi SSL hoặc không vào được | Chưa bấm Enable HTTPS cho app | Vào HTTP Settings của app |
| `ping-web` gọi sai API | Biến `VITE_API_BASE_URL` của env `ping-web-develop` sai hoặc trống | Sửa biến rồi deploy lại. Biến này nhúng vào lúc build, nên phải build lại |
| Fail ở job `ci` | `make ci` của service fail | Chạy `make ci` trong `services/<svc>` ở local để xem lỗi |

---

Tên các mục menu CapRover trong tài liệu này có thể hơi khác tuỳ phiên bản.
