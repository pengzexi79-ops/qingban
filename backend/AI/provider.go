package ai

// 模型供应商(Provider)适配层:把"青伴要的能力"翻译成"各家协议的 HTTP 调用"。
// 架构:模型可替换,业务层只面向本包接口;本阶段实现 OpenAI 兼容协议
// (Ollama 本地 + 任意兼容端点),Anthropic/Ollama 原生按需追加。
// 默认本地链路:Ollama http://localhost:11434(openai-compatible,无密钥)。
// 伪代码草稿:HTTP 细节以函数体内伪代码注释占位。

import (
	"time"

	"qingban/model"
)

// ChatMessage:送入模型的单条消息(roles:system/user/assistant)。
type ChatMessage struct {
	// Role:system/user/assistant。
	Role string `json:"role"`
	// Content:文本内容(本阶段纯文本;多模态第二阶段追加 content 数组)。
	Content string `json:"content"`
}

// ChatRequest:一次对话补全请求(内部形态,与协议解耦)。
type ChatRequest struct {
	// Model:模型 id(来自 ApiProfile 的 chatModel 等字段)。
	Model string
	// Messages:完整消息序列(由调用方按预算裁剪)。
	Messages []ChatMessage
	// Temperature:采样温度(空=服务商默认)。
	Temperature *float64
	// Stream:是否流式。
	Stream bool
}

// Usage:用量回传(协议字段映射后的统一形态)。
type Usage struct {
	// InputTokens:输入 tokens。
	InputTokens int
	// OutputTokens:输出 tokens。
	OutputTokens int
	// CachedTokens:缓存命中 tokens(供应商支持才有)。
	CachedTokens int
}

// DeltaChunk:流式输出的一帧增量。
type DeltaChunk struct {
	// Content:本次增量正文(可为空,仅心跳/结束标记)。
	Content string
	// Done:是否结束帧(流结束后适配器发出)。
	Done bool
}

// ChatResult:非流式调用完整结果。
type ChatResult struct {
	// Content:模型完整回复。
	Content string
	// Usage:用量。
	Usage Usage
	// ProviderRequestID:供应商侧请求 id(落 usage_records 排障)。
	ProviderRequestID string
}

// Client:某条 ApiProfile 对应的协议客户端(每次真实调用构建,不常驻避免密钥滞留)。
type Client struct {
	// Profile:使用的配置(含解密后密钥,仅调用生命周期内存活)。
	Profile *model.ApiProfile
	// SecretKey:解密后的明文密钥(用完即弃;空=本地模型无密钥)。
	SecretKey string
}

// NewClient:构造协议客户端。
// secret:服务层用 utils.SecretBox 解密后传入(服务层是唯一持有 SecretBox 的地方)。
func NewClient(profile *model.ApiProfile, secret string) *Client {
	return &Client{Profile: profile, SecretKey: secret}
}

// ChatOnce:非流式对话补全(同步发送链路与群聊轮次单成员发言用)。
func (c *Client) ChatOnce(req ChatRequest) (*ChatResult, error) {
	// url, payload := c.route(req, stream: false)               // ① 协议分发:
	//     // openai-compatible → POST base/chat/completions(含 Authorization: Bearer key, 无 key 不带头)
	//     // ollama 原生 → POST base/api/chat;anthropic → /v1/messages(第二阶段)
	// ctx, cancel := timed(60s); defer cancel()                  // ② 超时 60s(可配)
	// resp, err := httpPost(ctx, url, payload)                   // ③ 非 2xx → PROVIDER_ERROR(截断响应体,不外泄密钥/原文)
	// out := ChatResult{Content: resp.choices[0].message.content,
	//                   Usage: mapUsage(resp.usage), ProviderRequestID: resp.id}
	// return &out, nil                                           // ④
	return nil, nil // TODO(实现):见函数注释 ①~④
}

// ChatStream:流式对话补全(onDelta 逐帧回调;结束后返回聚合结果)。
func (c *Client) ChatStream(req ChatRequest, onDelta func(delta DeltaChunk) error) (*ChatResult, error) {
	// resp := httpPostStream(c.route(req, stream: true))         // ① stream=true,Content-Type: text/event-stream
	// defer resp.Close()
	// scanner := bufio.NewScanner(resp.Body)
	// var content strings.Builder; var usage Usage
	// for scanner.Scan() {                                        // ② 逐行 "data: {json}"
	//     line := scanner.Text(); if !hasPrefix(line, "data:") { continue }
	//     data := trimPrefix(line, "data: ")
	//     if data == "[DONE]" { break }                           // ③ 结束帧
	//     var ch struct{ Choices []struct{ Delta struct{ Content string } }; Usage Usage }
	//     jsonUnmarshal(data, &ch)
	//     if ch.Choices[0].Delta.Content != "" {
	//         content.WriteString(d); if onDelta != nil { if err := onDelta({d, false}); err != nil { cancel; return } }
	//     }
	//     if ch.Usage != nil { usage = ch.Usage }                  // ④ openai 末帧带 usage;缺失则 0
	// }
	// if ctx 取消/读错误 { return nil, PROVIDER_ERROR }
	// return &ChatResult{content.String(), usage, respID}, nil
	return nil, nil // TODO(实现):见函数注释 ①~④
}

// ListModels:拉取模型目录并推断能力(供 /api-profiles/{id}/models 回显)。
// 推断规则(实现用):名称含 vision/qwen-vl/gpt-4o-mini → vision;tts/voice → tts;
// stt/whisper → hearing;embedding/bge → embedding;否则至少 chat(+streaming)。
func (c *Client) ListModels() ([]model.ModelInfo, error) {
	// if c.Profile.Protocol == ProtoOllama { GET base/api/tags → models[] }   // Ollama 原生
	// else { GET base/models → data[] }                                          // openai 兼容
	// // 逐项:capabilities = inferCaps(name)(上表);serving = Ollama 时查 /api/ps 已加载
	// return modelInfos, nil
	return nil, nil // TODO(实现):见函数注释
}

// Test:连通性测试(GET /api-profiles/{id}/test 用)。
// 探测:Ollama → GET /api/tags;兼容端点 → 1-token chat(或 GET /models);
// 记录往返延迟;返回 status/latencyMs/detail(失败原因摘要,脱敏)。
func (c *Client) Test() (status string, latencyMs int64, detail string) {
	// start := time.Now(); _, err := listModelsOrPing(c)
	// ms := time.Since(start).Milliseconds()
	// if err != nil { return "failed", ms, errMessage(err) }
	// return "success", ms, ""
	return "", 0, "" // TODO(实现)
}

// timed:超时上下文辅助(防止对不可达端点无限挂起)。
func timed(timeout time.Duration) (cancel func()) {
	// ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// return cancel
	return func() {}
}
