// Package config 配置单元测试
package config

// 导入依赖
import (
	"os"      // 环境变量操作
	"testing" // 测试框架
)

// TestLoad_requiresDatabaseURL 测试 DATABASE_URL 为空时应返回错误
func TestLoad_requiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "") // 设置为空
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is empty") // 应该报错
	}
}

// TestLoad_defaults 测试默认配置值是否正确
func TestLoad_defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test?sslmode=disable") // 提供必填值
	t.Setenv("HTTP_PORT", "")                                             // 清空以使用默认
	t.Setenv("CHAIN_ID", "31337")                                         // 默认链 ID

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	// 验证端口默认值
	if cfg.HTTPPort != "8080" {
		t.Fatalf("HTTPPort = %q, want 8080", cfg.HTTPPort)
	}
	// 验证 ChainID 默认值
	if cfg.ChainID != 31337 {
		t.Fatalf("ChainID = %d, want 31337", cfg.ChainID)
	}
}

// TestLoad_customPort 测试自定义端口配置
func TestLoad_customPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test?sslmode=disable") // 必填
	t.Setenv("HTTP_PORT", "9090")                                         // 自定义端口

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	// 验证自定义端口
	if cfg.HTTPPort != "9090" {
		t.Fatalf("HTTPPort = %q, want 9090", cfg.HTTPPort)
	}

	_ = os.Unsetenv("HTTP_PORT") // 清理环境变量
}
