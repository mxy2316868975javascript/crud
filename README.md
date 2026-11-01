# CET-4 词汇量测试 API

用 Go + PostgreSQL 实现的词汇量测试后端：维护 CET-4 词库，匿名用户创建四选一测试、提交答案，服务端判分并按「正确率 × 词库总量」估算词汇量。

> 前端开发者先看 [docs/frontend-guide.md](docs/frontend-guide.md)。

## 技术栈

- Go 1.27，标准库 `net/http`，无 Web 框架
- PostgreSQL 16 + `pgx/v5`
- Docker Compose 启动数据库
- `swaggo/swag` 从 handler 注解生成 Swagger 2.0，再转换为 OpenAPI 3.0

## 目录结构

```text
cmd/
  api/                 服务启动与依赖装配
  seed/                种子导入入口
  openapi3/            Swagger 2.0 -> OpenAPI 3.0
internal/
  config/              环境变量
  domain/              领域模型、纯规则与错误
  repository/postgres/ SQL、事务、schema、批量导入
  service/             用例编排、仓储接口
  httpapi/             handler、DTO、路由、Swagger 注解
  httpapi/docs/        生成的 swagger.json / openapi.json / docs.html
  seed/                内嵌 cet4.csv 与解析
  integration/         黑盒集成测试
docker/                PostgreSQL 初始化脚本
```

依赖方向：`httpapi → service → domain`，`repository/postgres` 实现 `service` 的仓储接口；`cmd` 只负责装配。

没有前端页面。调接口先用 `/docs` 的 Swagger UI，或 [docs/requests.http](docs/requests.http) 里的请求。

## 快速开始

```bash
docker compose up -d --wait   # 启动 PostgreSQL，并自动创建 vocab_test
go run ./cmd/seed             # 导入内嵌的 3,849 个 CET-4 词汇
go run ./cmd/api              # 监听 :8080
```

验证：

```bash
curl -s localhost:8080/healthz
curl -s 'localhost:8080/api/v1/words?level=1&limit=2'
```

如果本机已有 PostgreSQL，也可以跳过 Docker（注意本机端口可能与 Compose 的 5433 不同）：

```bash
createdb vocab
DATABASE_URL='postgres://<user>@localhost:5432/vocab?sslmode=disable' go run ./cmd/seed
DATABASE_URL='postgres://<user>@localhost:5432/vocab?sslmode=disable' go run ./cmd/api
```

环境变量：

| 变量 | 默认值 |
|---|---|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5433/vocab?sslmode=disable` |
| `PORT` | `8080` |

## API

所有业务接口前缀为 `/api/v1`，请求和响应均为 JSON；错误响应为 `{"error":"..."}`。

| 方法 | 路径 | 说明 | 成功状态 |
|---|---|---|---|
| GET | `/healthz` | 进程和数据库健康检查 | 200 / 503 |
| POST | `/api/v1/words` | 新增词汇 | 201 |
| GET | `/api/v1/words?level=&limit=&offset=` | 分页列出词汇 | 200 |
| GET | `/api/v1/words/{id}` | 查询词汇 | 200 |
| PUT | `/api/v1/words/{id}` | 全量替换 | 200 |
| PATCH | `/api/v1/words/{id}` | 局部更新 | 200 |
| DELETE | `/api/v1/words/{id}` | 删除词汇 | 204 |
| POST | `/api/v1/tests` | 创建测试并抽题 | 201 |
| GET | `/api/v1/tests/{id}` | 查询测试 | 200 |
| POST | `/api/v1/tests/{id}/answers` | 提交答案并判分 | 200 |

状态码：400 参数不合法，404 资源不存在，409 单词重复或测试已提交或词汇被测试引用，500 服务端错误，503 数据库不可用。

### 接口文档

文档由 handler 上的 swaggo 注解生成，不需要手写规范。`go generate ./...` 会先生成 `internal/httpapi/docs/swagger.json`（Swagger 2.0），再用 `cmd/openapi3` 转成 `internal/httpapi/docs/openapi.json`（OpenAPI 3.0.3），Swagger UI 加载后者。通用 API 注解在 `internal/httpapi/router.go`。

```bash
curl -s localhost:8080/openapi.json   # OpenAPI 3.0.3
curl -s localhost:8080/swagger.json   # Swagger 2.0 源文件
```

生成时可以用 `PUBLIC_URL` 指定 `servers` 地址，默认 `http://localhost:8080`：

```bash
PUBLIC_URL=https://api.example.com go generate ./...
```

生成结果随代码提交。`swag` 已作为工具依赖固定在 `go.mod`，`go generate` 使用 `go tool swag`，不需要预装 CLI。转换器会给 3.0 的 `correct_count`、`estimated_vocabulary`、`completed_at` 补 `nullable`。Swagger UI 的 CSS/JS 从 unpkg CDN 加载；离线环境只能访问 JSON 规范，页面不会渲染。

### 词汇

```bash
curl -s -X POST localhost:8080/api/v1/words \
  -H 'Content-Type: application/json' \
  -d '{"word":"abandon","meaning":"v. 放弃；抛弃","level":1}'
```

```json
{"id":1,"word":"abandon","meaning":"v. 放弃；抛弃","level":1}
```

`PUT` 必须提供 `word`、`meaning`、`level`；`PATCH` 至少提供一个字段。`level` 范围 1–5。

列表响应：

```json
{"items":[{"id":1,"word":"abandon","meaning":"v. 放弃；抛弃","level":1}],"total":3849}
```

`limit` 默认 50，最大 200；`offset` 默认 0；`level` 可选 1–5。

### 测试流程

创建测试，`count` 默认 20，范围 1–100：

```bash
curl -s -X POST localhost:8080/api/v1/tests \
  -H 'Content-Type: application/json' -d '{"count":2}'
```

```json
{
  "id": 1,
  "question_count": 2,
  "completed": false,
  "questions": [
    {"word_id": 1, "word": "abandon", "options": ["v. 放弃；抛弃", "n. 能力", "adj. 抽象的", "v. 适应"]}
  ]
}
```

每个题目有 4 个选项，接口不会返回正确答案。词库不足时返回 400，不会静默缩短题数。

提交答案时必须覆盖全部题目：

```bash
curl -s -X POST localhost:8080/api/v1/tests/1/answers \
  -H 'Content-Type: application/json' \
  -d '{"answers":[{"word_id":1,"selected_index":0},{"word_id":2,"selected_index":3}]}'
```

```json
{
  "id": 1,
  "question_count": 2,
  "completed": true,
  "correct_count": 2,
  "estimated_vocabulary": 3849,
  "completed_at": "2026-10-31T12:00:00Z"
}
```

估算公式：`round(correct_count / question_count × total_words_at_start)`。`total_words_at_start` 是创建测试时的词库总量快照，之后增删词不会改变已创建测试的结果。

测试是匿名的：创建时返回 `id`，凭这个 ID 提交答案和查询结果，没有用户表和鉴权。

## 词表来源

`internal/seed/cet4.csv` 来自 [ECDICT](https://github.com/skywind3000/ECDICT)（MIT 协议，见 `internal/seed/ECDICT-LICENSE`）：

- 过滤 `tag` 包含 `cet4` 的条目，共 3,849 词
- `translation` 中的换行替换为 `; `，作为 `meaning`
- 按 `frq` 词频升序五等分生成 `level`：1 最常用，5 最生僻；缺失词频的 61 个词归入 level 5
- 少于官方四级大纲的约 4,500 词

CSV 格式为 `word,meaning,level`，已通过 `//go:embed` 打进 `cmd/seed`。可重复执行 `go run ./cmd/seed` 导入；同名词会被更新，会覆盖手工修改。

## 测试

测试会清空 `vocab_test` 库中的表，不要指向开发库：

```bash
TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5433/vocab_test?sslmode=disable' go test ./...
```

Docker Compose 首次初始化时会自动创建 `vocab_test`。如果数据库已经存在但缺少该库，手动创建：

```bash
createdb vocab_test
```

连接不上测试库时，测试会跳过而不是失败。

## 已知限制

- 词库管理接口没有鉴权，任何人可增删改
- `ORDER BY random()` 抽题是 O(n)，当前 3,849 词足够；词库上万时再换 `TABLESAMPLE`
- schema 用 `CREATE TABLE IF NOT EXISTS` 创建，没有迁移工具
- 没有测试历史列表/删除、CORS、限流、优雅停机
- 词汇量估算是粗略值，不是统计意义上的能力估计
