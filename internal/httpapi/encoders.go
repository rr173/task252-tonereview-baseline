// Package httpapi 暴露以 /api 为前缀的 JSON HTTP 接口，调用 service 编排层。
// 路径参数使用 Go 1.22+ 的内置模式匹配（{id}）。
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"task252-tonereview/internal/model"
	"task252-tonereview/internal/service"
)

// API 聚合服务与路由。
type API struct {
	svc *service.Service
}

// writeJSON 以给定状态码写出 JSON 响应。
func writeJSON(w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// readJSON 解析请求体到目标结构；任何解析失败（语法/未知字段/空体）均包装为
// model.ErrBadRequest，使错误映射统一返回 400。
func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%w: %v", model.ErrBadRequest, err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("%w: request body must contain one JSON value", model.ErrBadRequest)
		}
		return fmt.Errorf("%w: trailing data: %v", model.ErrBadRequest, err)
	}
	return nil
}

// writeError 将领域错误映射为 HTTP 状态码。
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrBadRequest):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrConflict),
		errors.Is(err, model.ErrFrozen),
		errors.Is(err, model.ErrSealed),
		errors.Is(err, model.ErrInvalidState),
		errors.Is(err, model.ErrDuplicate):
		status = http.StatusConflict
	}
	writeJSON(w, map[string]any{"error": err.Error(), "status": status}, status)
}
