package httpapi

import (
	"net/http"

	"github.com/mxy2316868975javascript/crud/internal/service"
)

// API 持有 HTTP handler 需要的服务依赖。
type API struct {
	words  *service.WordService
	tests  *service.TestService
	pinger Pinger
}

// swaggerInfo 仅用于承载 Swagger 通用注解，本身不参与运行。
// @title CET-4 词汇量测试接口
// @version 1.0.0
// @description CET-4 词汇量测试后端：词库增删改查、四选一测试、判分和健康检查。测试为匿名，凭测试 ID 提交答案；估算公式为：正确题数 ÷ 题目总数 × 创建测试时的词库总量，结果四舍五入。
// @host localhost:8080
// @BasePath /
func swaggerInfo() {}

// NewMux 创建并返回注册了全部路由的 HTTP handler。
func NewMux(words *service.WordService, tests *service.TestService, pinger Pinger) http.Handler {
	api := &API{words: words, tests: tests, pinger: pinger}
	router := http.NewServeMux()
	router.HandleFunc("GET /healthz", api.health)
	router.HandleFunc("GET /swagger.json", api.swaggerSpec)
	router.HandleFunc("GET /openapi.json", api.openapiSpec)
	router.HandleFunc("GET /docs", api.swaggerUI)
	router.HandleFunc("POST /api/v1/words", api.createWord)
	router.HandleFunc("GET /api/v1/words", api.listWords)
	router.HandleFunc("GET /api/v1/words/{id}", api.getWord)
	router.HandleFunc("PUT /api/v1/words/{id}", api.putWord)
	router.HandleFunc("PATCH /api/v1/words/{id}", api.patchWord)
	router.HandleFunc("DELETE /api/v1/words/{id}", api.deleteWord)
	router.HandleFunc("POST /api/v1/tests", api.createTest)
	router.HandleFunc("GET /api/v1/tests/{id}", api.getTest)
	router.HandleFunc("POST /api/v1/tests/{id}/answers", api.submitAnswers)
	return router
}
