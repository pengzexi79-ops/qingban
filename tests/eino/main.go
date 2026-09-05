// tests/eino:Eino(ADK)集成验收程序(真实可运行)。
//
// 用法:
//   离线验收(默认):go run .                                  —— echo 模型,无需网络
//   OpenAI 兼容协议实测(DeepSeek 等,测缓存命中率):
//     $env:DEEPSEEK_API_KEY="sk-..."          # key 经环境变量传入,不进 shell 历史/代码
//     go run . --api $env:DEEPSEEK_API_KEY [--base-url https://api.deepseek.com] [--model deepseek-chat] [--rounds 4]
//
// 验收目标:
//  ① model.BaseChatModel 组件(echo 离线 + rawChatModel 走 OpenAI 兼容 HTTP)
//  ② adk.NewChatModelAgent + Instruction;adk.NewTypedRunner:Query / Run
//  ③ AgentEvent 流消费(adk.GetMessage)+ 流正常结束
//  ④ 缓存命中率统计(本文件新增重点):
//     - OpenAI 协议标准字段:usage.prompt_tokens_details.cached_tokens
//     - DeepSeek 扩展字段:usage.prompt_cache_hit_tokens / prompt_cache_miss_tokens
//       (eino-ext 的 go-openai 客户端不解析 DeepSeek 扩展字段 → rawChatModel 自行解析保留)
//     命中率:openaiRate = cached/prompt;deepseekRate = hit/(hit+miss)
// 说明:密钥不做任何默认值/文件落盘;只经命令行或环境变量传入。

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ================== 1. echoChatModel(离线链路验收) ==================

// echoChatModel:实现 model.BaseChatModel;把最后一条消息回显(带 [echo] 前缀)。
type echoChatModel struct{}

// Generate:非流式补全。
func (m *echoChatModel) Generate(_ context.Context, input []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	last := ""
	if len(input) > 0 {
		last = input[len(input)-1].Content
	}
	return &schema.Message{Role: schema.Assistant, Content: "[echo] " + last}, nil
}

// Stream:流式补全(单帧下发)。
func (m *echoChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, err := m.Generate(ctx, input, opts...)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
}

// 编译期断言:满足 BaseChatModel 接口。
var _ model.BaseChatModel = (*echoChatModel)(nil)

// ================== 2. rawChatModel(OpenAI 兼容协议 + 缓存字段原始解析) ==================

// usageReport:一次真实调用的用量与两套缓存统计。
type usageReport struct {
	// PromptTokens:输入总 token。
	PromptTokens int
	// CompletionTokens:输出 token。
	CompletionTokens int
	// OpenAICachedTokens:OpenAI 标准字段(prompt_tokens_details.cached_tokens)。
	OpenAICachedTokens int
	// DSHitTokens/DSMissTokens:DeepSeek 扩展字段(prompt_cache_hit/miss_tokens)。
	DSHitTokens  int
	DSMissTokens int
	// LatencyMs:端到端耗时(ms)。
	LatencyMs int64
}

// openAIUsageRaw:usage 最小结构——同时容纳两套缓存字段,额外字段无损解析。
type openAIUsageRaw struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	// OpenAI 协议:cached_tokens 嵌套在 prompt_tokens_details
	PromptTokensDetails *struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
	// DeepSeek 协议:扩展字段平铺在 usage 顶层
	PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
}

// openAIRespRaw:补全响应最小结构。
type openAIRespRaw struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage openAIUsageRaw `json:"usage"`
}

// rawChatModel:BaseChatModel 实现——直连 OpenAI 兼容端点(DeepSeek/OpenAI/Ollama 均可)。
// 与 eino-ext 组件的差异:自行解析 DeepSeek 扩展缓存字段(go-openai 客户端会丢弃)。
type rawChatModel struct {
	// BaseURL:端点根,如 https://api.deepseek.com(自动补 /chat/completions)。
	BaseURL string
	// APIKey:密钥(仅进程内存,来源:命令行/环境变量;勿落代码与仓库)。
	APIKey string
	// Model:模型名(deepseek-chat 等)。
	Model string
	// Client:HTTP 客户端(120s 超时,见 newRawChatModel)。
	Client *http.Client
	// lastUsage:最近一次调用用量(本验收顺序调用;并发场景应改经回调/上下文携带)。
	lastUsage usageReport
}

// newRawChatModel:构造组件(默认端点 https://api.deepseek.com)。
func newRawChatModel(baseURL, apiKey, modelName string) *rawChatModel {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		base = "https://api.deepseek.com"
	}
	return &rawChatModel{
		BaseURL: base,
		APIKey:  apiKey,
		Model:   modelName,
		Client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// chatReqMsg:单条消息载荷。
type chatReqMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatReq:补全请求体(本验收只发文本消息,关闭流式)。
type chatReq struct {
	Model    string       `json:"model"`
	Messages []chatReqMsg `json:"messages"`
	Stream   bool         `json:"stream"`
}

// callOnce:执行一次非流式补全,返回助手消息 + 完整用量(两套缓存字段)。
func (m *rawChatModel) callOnce(ctx context.Context, input []*schema.Message) (*schema.Message, usageReport, error) {
	// ① 组装 messages(schema 消息 → openai role/content)
	msgs := make([]chatReqMsg, 0, len(input))
	for _, mm := range input {
		if mm != nil {
			msgs = append(msgs, chatReqMsg{Role: string(mm.Role), Content: mm.Content})
		}
	}
	body, err := json.Marshal(chatReq{Model: m.Model, Messages: msgs})
	if err != nil {
		return nil, usageReport{}, err
	}

	// ② POST {base}/chat/completions
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, usageReport{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if m.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+m.APIKey)
	}
	resp, err := m.Client.Do(req)
	lat := time.Since(start).Milliseconds()
	if err != nil {
		return nil, usageReport{LatencyMs: lat}, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, usageReport{LatencyMs: lat},
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	// ③ 解析响应与用量
	var out openAIRespRaw
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, usageReport{LatencyMs: lat}, err
	}
	u := usageReport{PromptTokens: out.Usage.PromptTokens, CompletionTokens: out.Usage.CompletionTokens, LatencyMs: lat}
	if out.Usage.PromptTokensDetails != nil {
		u.OpenAICachedTokens = out.Usage.PromptTokensDetails.CachedTokens
	}
	u.DSHitTokens, u.DSMissTokens = out.Usage.PromptCacheHitTokens, out.Usage.PromptCacheMissTokens
	content := ""
	if len(out.Choices) > 0 {
		content = out.Choices[0].Message.Content
	}
	return &schema.Message{Role: schema.Assistant, Content: content}, u, nil
}

// Generate:BaseChatModel 非流式。
func (m *rawChatModel) Generate(ctx context.Context, input []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	msg, u, err := m.callOnce(ctx, input)
	m.lastUsage = u // 顺序调用场景读取统计
	return msg, err
}

// Stream:BaseChatModel 流式(单帧复用;真实流式解析在组件层实现,语义一致)。
func (m *rawChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, err := m.Generate(ctx, input, opts...)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
}

// ================== 3. 事件流消费(backend/AI/agent.go RunTurn 的雏形) ==================

// collectFrom:消费事件流,拼出完整助手回复(ev.Err 立即返回;!ok 流结束)。
func collectFrom(iter *adk.AsyncIterator[*adk.TypedAgentEvent[*schema.Message]]) (string, error) {
	var sb strings.Builder
	for {
		ev, ok := iter.Next()
		if !ok {
			return sb.String(), nil
		}
		if ev.Err != nil {
			return "", ev.Err
		}
		if ev.Output != nil && ev.Output.MessageOutput != nil {
			msg, _, err := adk.GetMessage(ev) // 自动拼接流式帧
			if err != nil {
				return "", err
			}
			if msg != nil {
				sb.WriteString(msg.Content)
			}
		}
	}
}

// runAndCollect:runner.Run(消息列表)→ 收集回复。
func runAndCollect(runner *adk.TypedRunner[*schema.Message], ctx context.Context, messages []*schema.Message) (string, error) {
	return collectFrom(runner.Run(ctx, messages))
}

// queryAndCollect:runner.Query(单串)→ 收集回复。
func queryAndCollect(runner *adk.TypedRunner[*schema.Message], ctx context.Context, q string) (string, error) {
	return collectFrom(runner.Query(ctx, q))
}

// ================== 4. 断言辅助 ==================

var fails int

func check(name string, cond bool, detail string) {
	if cond {
		fmt.Printf("[PASS] %s\n", name)
	} else {
		fails++
		fmt.Printf("[FAIL] %s: %s\n", name, detail)
	}
}

// ================== 5. 离线验收(echo) ==================

func runOfflineEcho(ctx context.Context) {
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "companion-echo",
		Description: "验收用回声 AI 好友",
		Instruction: "你是青伴的验收角色。收到消息请按原话回显。",
		Model:       &echoChatModel{},
	})
	if err != nil {
		fmt.Println("[FAIL] NewChatModelAgent:", err)
		os.Exit(1)
	}
	runner := adk.NewTypedRunner[*schema.Message](adk.TypedRunnerConfig[*schema.Message]{
		Agent:           agent,
		EnableStreaming: true,
	})

	// 单轮 Query
	reply, err := queryAndCollect(runner, ctx, "你好,请记住我喜欢雨天听歌")
	if err != nil {
		fmt.Println("[FAIL] Query 单轮:", err)
		os.Exit(1)
	}
	check("query.single", reply == "[echo] 你好,请记住我喜欢雨天听歌", fmt.Sprintf("reply=%q", reply))

	// 多轮 Run(调用方维护历史,官方 ch02 模式)
	history := []*schema.Message{schema.UserMessage("我的名字是小林")}
	r1, err := runAndCollect(runner, ctx, history)
	if err != nil {
		fmt.Println("[FAIL] Run 第一轮:", err)
		os.Exit(1)
	}
	check("run.turn1", r1 == "[echo] 我的名字是小林", fmt.Sprintf("r1=%q", r1))
	history = append(history, schema.AssistantMessage(r1, nil), schema.UserMessage("我叫什么名字?"))
	r2, err := runAndCollect(runner, ctx, history)
	if err != nil {
		fmt.Println("[FAIL] Run 第二轮:", err)
		os.Exit(1)
	}
	check("run.turn2.history", r2 == "[echo] 我叫什么名字?", fmt.Sprintf("r2=%q", r2))

	if fails > 0 {
		os.Exit(1)
	}
	fmt.Println("\neino:离线 ADK 链路验收通过(ChatModelAgent + TypedRunner + AgentEvent 流)")
}

// ================== 6. 真实模型缓存命中率实测 ==================

// runCacheBenchmark:同一 prompt 重复调用 rounds 轮,统计两套缓存命中率。
// 缓存语义说明:
//   - DeepSeek 自动上下文缓存:相同前缀(含 system)的第二次起命中,prompt_cache_hit_tokens 上升;
//   - OpenAI 标准字段 cached_tokens 对 DeepSeek 端点通常恒 0(DeepSeek 不回填该非标字段)→ 如实展示差异。
func runCacheBenchmark(ctx context.Context, m *rawChatModel, rounds int) {
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "companion-raw",
		Description: "OpenAI 兼容协议实测",
		Instruction: "你是青伴的测试角色。请用中文简要回答。", // 稳定 system 前缀:命中前提
		Model:       m,
	})
	if err != nil {
		fmt.Println("[FAIL] NewChatModelAgent:", err)
		os.Exit(1)
	}
	runner := adk.NewTypedRunner[*schema.Message](adk.TypedRunnerConfig[*schema.Message]{
		Agent:           agent,
		EnableStreaming: true,
	})

	// 固定长 prompt(内容一致 → DeepSeek 磁盘缓存才可能命中)
	prompt := strings.Repeat("青伴是一个本地优先的 AI 陪伴产品,强调离线可用、数据自持、长期记忆与情感陪伴。", 8) +
		"\n请用 50 字以内总结这段话的核心设计理念。"

	fmt.Printf("端点:%s 模型:%s 轮数:%d\n", m.BaseURL+"/chat/completions", m.Model, rounds)
	fmt.Println("轮次 | prompt | openaiCached | dsHit | dsMiss | openai命中 | deepseek命中 | 耗时ms | 回复预览")
	totalHit, totalMiss := 0, 0
	for i := 1; i <= rounds; i++ {
		msgs := []*schema.Message{schema.UserMessage(prompt)}
		reply, runErr := runAndCollect(runner, ctx, msgs)
		if runErr != nil {
			fmt.Printf("%3d  |   --   |      --      |  --  |   --   |     --    |      --     |   --   | 错误:%v\n", i, runErr)
			continue
		}
		u := m.lastUsage
		openaiRate := 0.0
		if u.PromptTokens > 0 {
			openaiRate = float64(u.OpenAICachedTokens) / float64(u.PromptTokens)
		}
		dsRate := 0.0
		if u.DSHitTokens+u.DSMissTokens > 0 {
			dsRate = float64(u.DSHitTokens) / float64(u.DSHitTokens+u.DSMissTokens)
		}
		totalHit += u.DSHitTokens
		totalMiss += u.DSMissTokens
		preview := reply
		if r := []rune(reply); len(r) > 12 {
			preview = string(r[:12]) + "…"
		}
		fmt.Printf("%3d  | %5d | %12d | %4d | %5d | %7.1f%%  | %10.1f%%   | %6d | %s\n",
			i, u.PromptTokens, u.OpenAICachedTokens, u.DSHitTokens, u.DSMissTokens,
			openaiRate*100, dsRate*100, u.LatencyMs, preview)
	}

	// 汇总结论
	if totalHit+totalMiss > 0 {
		fmt.Printf("\nDeepSeek 扩展字段合计命中率:%.1f%%(hit=%d miss=%d)\n",
			float64(totalHit)/float64(totalHit+totalMiss)*100, totalHit, totalMiss)
	}
}

func main() {
	ctx := context.Background()

	apiKey := flag.String("api", "", "OpenAI 兼容端点 API Key(建议经环境变量传入)")
	baseURL := flag.String("base-url", "", "端点根地址(默认 https://api.deepseek.com)")
	modelName := flag.String("model", "", "模型名(默认 deepseek-chat)")
	rounds := flag.Int("rounds", 4, "重复调用轮数(缓存命中率统计)")
	flag.Parse()

	if *apiKey == "" {
		runOfflineEcho(ctx) // 默认离线验收
		return
	}

	name := *modelName
	if name == "" {
		name = "deepseek-chat"
	}
	m := newRawChatModel(*baseURL, *apiKey, name)
	runCacheBenchmark(ctx, m, *rounds)
}
