package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// writeJSON 统一写出 JSON 响应。
func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	if value != nil {
		_ = json.NewEncoder(response).Encode(value)
	}
}

// writeError 写出统一错误体。
func writeError(response http.ResponseWriter, status int, message string) {
	writeJSON(response, status, ErrorResponse{Error: message})
}

// readJSON 将请求体限制在 1 MB，并拒绝未知字段和多个 JSON 值。
func readJSON(response http.ResponseWriter, request *http.Request, target any) bool {
	request.Body = http.MaxBytesReader(response, request.Body, 1<<20)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(response, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(response, http.StatusBadRequest, "invalid JSON: expected a single object")
		return false
	}
	return true
}

// pathID 解析并校验路径里的 id。
func pathID(response http.ResponseWriter, request *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(response, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}
