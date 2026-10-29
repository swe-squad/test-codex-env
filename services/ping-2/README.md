# ping-2

Service nội bộ Python/FastAPI từ template `python-fastapi`, dùng test CI/CD trên dev.
Image: `ghcr.io/swe-squad/ping-2`, container port `8000`.

```text
Browser → ping-web /api/v1/ping → ping-api /api/v1/ping → ping-2 /api/v1/ping
Browser ← ping-web              ← ping-api              ← {"message":"pong","service":"ping-2"}
```

`GET /health` trả `{"status":"ok"}`. `GET /api/v1/ping` trả pong ở trên.
`ping-api` chuyển tiếp pong từ `ping-2`, timeout sau 3 giây và trả `502` nếu
không kết nối được hoặc phản hồi không hợp lệ. `/health` của mỗi service độc lập
với các service khác.

## Chạy local

Mỗi lệnh chạy server dưới đây dùng một terminal riêng, từ repo root:

```sh
make -C services/ping-2 install
make -C services/ping-2 run
```

```sh
PING_2_URL=http://localhost:8000 make -C services/ping-api run
```

```sh
make -C services/ping-web install
make -C services/ping-web dev
```

Mở địa chỉ Vite in ra, bấm **Ping** để thấy `pong (ping-2)`.
Để `VITE_API_BASE_URL` rỗng hoặc không đặt: Vite proxy `/api/` tới localhost:8080;
nginx trong image proxy tới `PING_API_URL` khi deploy. Giá trị `VITE_API_BASE_URL`
khác rỗng sẽ bỏ qua proxy này và cần API hỗ trợ truy cập từ browser.

Kiểm tra API trực tiếp:

```sh
curl --fail http://localhost:8080/api/v1/ping
```

## Cấu hình CapRover dev

| App | Container port | Runtime env trong App Configs |
| --- | --- | --- |
| `ping-web-dev` | `80` | `PING_API_URL=http://srv-captain--ping-api-dev:8080` |
| `ping-api-dev` | `8080` | `PING_2_URL=http://srv-captain--ping-2-dev:8000` |
| `ping-2-dev` | `8000` | `PORT=8000` (mặc định trong image) |

Code/image không có giá trị mặc định: thiếu `PING_API_URL` hoặc `PING_2_URL` thì container
dừng ngay khi khởi động. Production dùng app không có `-dev`, ví dụ
`PING_2_URL=http://srv-captain--ping-2:8000`.

- Tạo app `ping-2-dev`, bật **App Token** trong Deployment.
- Bật **Do not expose as web app** cho `ping-2-dev`, không thêm Port Mapping.
  Có thể bật tương tự cho `ping-api-dev` vì web đã proxy qua mạng nội bộ.
  Đây là [cấu hình CapRover](https://caprover.com/docs/app-configuration),
  việc thêm service vào repo chưa tự bật tùy chọn này.
- `ping-web-dev` có domain public và HTTPS; `ping-2-dev` không cần domain public/HTTPS.
- Registry trong CapRover → Cluster cần quyền đọc `ghcr.io/swe-squad/*`
  (PAT có `read:packages`).
- Trong GitHub environment `codex-sea-dev`, để biến build `VITE_API_BASE_URL`
  rỗng hoặc bỏ biến để web gọi `/api/` cùng origin. Biến này được chia sẻ giữa
  các frontend; kiểm tra các service khác trước khi đổi giá trị có sẵn.

Chạy lệnh sau trong terminal riêng với giá trị thật, không gửi token vào chat:

```sh
CAPROVER_SERVER=... APP_NAME_DEVELOP=ping-2-dev CAPROVER_TOKEN_DEVELOP=... \
  scripts/setup-service.sh ping-2 dev
```

Script thiết lập secrets `CAPROVER_SERVER`, `PING_2_DEV_APP_NAME`,
`PING_2_DEV_CAPROVER_TOKEN` trong `codex-sea-dev`. Secrets deploy của `ping-api`
và `ping-web` vẫn cần tồn tại như trước.

## Test CI/CD

Chạy `make ci` trong từng service: `ping-2`, `ping-api`, `ping-web`.
Workflow dùng chung tự nhận service mới qua `service.json`; không cần sửa workflow.
Mở PR vào `develop`, CI sẽ kiểm tra cả ba service đã đổi. Sau khi thiết lập app/secrets,
có thể comment `/deploy ping-2 ping-api ping-web` trên PR, hoặc merge vào `develop`
để deploy. Các service deploy độc lập nên có thể gặp `502` tạm thời khi rollout;
sau khi cả ba đã sẵn sàng, mở web và bấm **Ping**.

Đổi riêng `ping-2` trong lần test sau sẽ chỉ chạy CI/deploy cho `ping-2`.
Không thêm service này vào `PRODUCTION_SERVICES` vì chỉ dùng dev.
