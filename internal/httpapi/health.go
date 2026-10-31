package httpapi

import (
	"context"
	"net/http"
	"time"
)

// Pinger 是健康检查需要的数据库探活接口。
type Pinger interface {
	Ping(ctx context.Context) error
}

// 健康检查接口
// @Summary 健康检查
// @Description 检查进程和数据库连通性。数据库不可用时返回 503。
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /healthz [get]
func (api *API) health(response http.ResponseWriter, request *http.Request) {
	pingContext, cancel := context.WithTimeout(request.Context(), 2*time.Second)
	defer cancel()
	if err := api.pinger.Ping(pingContext); err != nil {
		writeJSON(response, http.StatusServiceUnavailable, HealthResponse{Status: "unavailable"})
		return
	}
	writeJSON(response, http.StatusOK, HealthResponse{Status: "ok"})
}
