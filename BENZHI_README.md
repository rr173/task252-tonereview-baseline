基于 Go 实现的濒危语言声调证据复核 Web 项目，一款后端服务，完成发音片段导入、说话人基线归一化、声调轮廓最小对立证据聚类、反例排除裁决与不可变分析版本冻结。

# BENZHI 评测说明

## 1. 项目类型

语言声学证据复核后端服务（非 OA/工单/预约，非数据看板消费类应用）。提供 JSON 形态的 `/api` 复核接口，供研究者以程序化方式提交田野录音词条、比对归一化基频轮廓、裁决声调最小对立并发布不可变分析版本。

## 2. 标准命令

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
go run ./cmd/tonereview --smoke-test
go run ./cmd/tonereview --addr :8080 --db tonoreview.db
```

- `--addr`：HTTP 监听地址，默认 `:8080`
- `--db`：SQLite 数据库文件路径，默认 `tonoreview.db`
- `--smoke-test`：不常驻；跑完端到端场景后关闭并重新打开数据库，退出码 0 表示通过

## 3. 评测镜像

`Dockerfile` 与 `benzhi.Dockerfile` 内容完全一致；使用 Go 1.26.3 Bookworm builder 和 Alpine 3.20 runtime 的多阶段构建，产物为 `/app/tonereview`。脚本第二个参数为目标平台。镜像不声明固定端口，服务监听地址由 `--addr` 指定。

```bash
./build_benzhi_docker.sh task252-tonereview:amd64 linux/amd64
docker run --rm task252-tonereview:amd64 --smoke-test

./build_benzhi_docker.sh task252-tonereview:arm64 linux/arm64
docker run --rm task252-tonereview:arm64 --smoke-test

docker run --rm -P task252-tonereview:amd64 --addr :8080 --db ./app.db
```

## 4. 冒烟自测契约（--smoke-test）

创建临时数据库 → 写入田野批次与发音片段 → 说话人基线归一化 → 最小对立证据聚类 → 裁决对立 → 冻结分析版本 → 走完批次状态机至封存 → 关闭并重新打开数据库，校验批次/片段/对立/版本全部持久化恢复后退出 0。

## 5. 核心 API（`/api` 前缀）

- 批次：`POST /api/batches`、`GET /api/batches`、`GET /api/batches/{id}`、`GET /api/batches/{id}/summary`、`POST /api/batches/{id}/review|publish|seal`
- 说话人：`POST /api/speakers`、`GET /api/speakers`、`GET /api/speakers/{id}`、`POST /api/speakers/{id}/baseline`
- 片段：`POST /api/batches/{id}/segments`、`GET /api/segments`、`GET /api/segments/{id}`、`POST /api/segments/{id}/verify|noise|exclude|restore|f0`、`GET /api/segments/{id}/f0`
- 对立：`POST /api/oppositions`、`GET /api/oppositions`、`GET /api/oppositions/{id}`、`POST /api/oppositions/{id}/evidence|recompute|adjudicate`、`GET /api/oppositions/{id}/cluster`、`POST /api/compare`
- 版本：`POST /api/versions`、`POST /api/versions/{id}/share|freeze`、`GET /api/versions`、`GET /api/versions/{id}`
- 自检：`GET /api/health`、`GET /api/stats`、`GET /api/selfcheck`

## 6. 环境与组件

- Go 1.26.3（GOTOOLCHAIN=local，CGO_ENABLED=0）
- SQLite 3.46.1（modernc.org/sqlite v1.52.0，CGO 无关）
