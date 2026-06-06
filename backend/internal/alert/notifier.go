// Package alert 提供告警通知能力（日志 + Webhook）
package alert

// 导入依赖
import (
	"bytes"         // 字节缓冲
	"encoding/json" // JSON 编解码
	"log"           // 日志
	"net/http"      // HTTP 客户端
	"time"          // 超时控制
)

// Notifier 告警通知器，可向 Webhook 推送告警
type Notifier struct {
	webhookURL string       // 目标 webhook 地址（空则只输出日志）
	client     *http.Client // HTTP 客户端（带超时）
}

// New 创建一个新的 Notifier
func New(webhookURL string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,                             // 保存 webhook 地址
		client:     &http.Client{Timeout: 5 * time.Second}, // 5 秒超时的 HTTP 客户端
	}
}

// Send 发送一条告警
// event：事件名；message：详细消息
func (n *Notifier) Send(event, message string) {
	// 总是输出本地日志，便于排查
	log.Printf("ALERT [%s]: %s", event, message)
	// 未配置 webhook 时直接返回
	if n.webhookURL == "" {
		return
	}
	// 序列化为 JSON 负载
	body, _ := json.Marshal(map[string]string{"event": event, "message": message})
	// 发送 POST 请求到 webhook
	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		// 推送失败仅记录日志
		log.Printf("webhook error: %v", err)
		return
	}
	_ = resp.Body.Close() // 关闭响应体，忽略错误
}
