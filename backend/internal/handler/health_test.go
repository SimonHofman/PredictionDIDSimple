// Package handler 健康检查单元测试
package handler

// 导入依赖
import (
	"encoding/json"     // JSON 解码
	"net/http"          // HTTP 常量
	"net/http/httptest" // 测试用 HTTP 工具
	"testing"           // 测试框架
)

// TestHealth_endpoint 测试 /health 接口返回 200 和 ok 状态
func TestHealth_endpoint(t *testing.T) {
	h := &Health{} // 无任何依赖
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.health(rec, req) // 调用处理函数

	// 校验状态码
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	// 解析 JSON
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	// 校验 status 字段
	if body["status"] != "ok" {
		t.Fatalf("status field = %q", body["status"])
	}
}

// TestReady_withoutDB 测试当无 DB 时 /ready 仍返回 200
func TestReady_withoutDB(t *testing.T) {
	h := &Health{} // DB 为 nil
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	h.ready(rec, req)

	// DB 为 nil 时不应报错
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 when DB nil", rec.Code)
	}
}
