// tests/clawbot:微信 ClawBot(官方 iLink bot API)接入复刻。
// 形态与 fastclaw-ai/weclaw 一致:扫码登录后【常驻】长轮询监听 getupdates,
// 微信里一对一给「微信ClawBot」发消息 → 进程调 OpenAI 协议端点(配置硬编码在本文件)
// → 把模型回复 sendmessage 回微信,形成完整对话闭环。
//
// 使用(按包运行,勿 go run 单文件——类型在 client.go/offline.go):
//     cd D:\开源项目\青伴\tests\clawbot && go run .            # 常驻对话
//     go run . -login                                        # 仅换绑/重新扫码登录
// 首次运行无凭据会自动弹二维码;微信扫码后先在微信给「微信ClawBot」发一条消息激活。
// 凭据存 .creds/auth.json(0600,含 bot_token,已被 .gitignore 忽略)。
//
// 线格式依据:微信官方插件 @tencent-weixin/openclaw-weixin(Backend API Protocol)
// 与 Go 参考实现 fastclaw-ai/weclaw(ilink 包);细节见同目录 client.go 顶部说明。

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
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/mdp/qrterminal/v3"
)

// ---- OpenAI 协议端点配置(硬编码,按需改成你自己的) ----
// 复刻的"大脑":任何 OpenAI 兼容 /chat/completions 服务都能接。
// 注意:下面 APIKey 属个人密钥,若本仓库要公开/提交务必先脱敏替换。
const (
	// chatBaseURL:openai 兼容端点基址(不含 /chat/completions 后缀)。
	chatBaseURL = "https://api.deepseek.com/v1"
	// chatAPIKey:端点鉴权密钥;无需鉴权(如本地 Ollama)留空。
	chatAPIKey = "sk-a0dd284eedd34dfc8099966c7007fd37"
	// chatModel:模型名(需已开通)。
	chatModel = "deepseek-v4-flash"
	// chatSystem:人设,每轮请求注入 system 消息。
	chatSystem = "你是我的微信 AI 助手,回复保持简洁口语化,能用一两句说完就别写小作文。"
	// chatTimeout:单次模型调用超时(模型慢时可调大)。
	chatTimeout = 180 * time.Second
	// convoTurns:每个会话保留的最近轮数(一轮=一问一答)。
	convoTurns = 12
)

// llmMsg:OpenAI 协议里的单条消息(与后端 AI/provider.go 的 ChatMessage 同构)。
type llmMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatOnce:调用 OpenAI 兼容 /chat/completions(非流式),返回整段回复文本。
func chatOnce(messages []llmMsg) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":    chatModel,
		"messages": messages,
		"stream":   false,
	})
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), chatTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(chatBaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if chatAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+chatAPIKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读响应: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("解析响应: %w", err)
	}
	if len(out.Choices) == 0 || out.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("choices 为空(响应原文截断:%s)", truncate(string(raw), 200))
	}
	return out.Choices[0].Message.Content, nil
}

// fails:累计失败数(与其它 tests/* 子包一致的断言约定)。
var fails int

// check:断言辅助;cond 为 false 时记失败。
func check(name string, cond bool, detail string) {
	if cond {
		fmt.Printf("[PASS] %s\n", name)
	} else {
		fails++
		fmt.Printf("[FAIL] %s: %s\n", name, detail)
	}
}

// pkgDir:本包所在目录(runtime.Caller 定位,不受 go run 的 cwd 影响)。
func pkgDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	return filepath.Dir(file)
}

// credsFile:联调凭据文件路径(与测试代码同目录的 .creds 下)。
func credsFile() string {
	return filepath.Join(pkgDir(), ".creds", "auth.json")
}

// loadCreds:读取已保存凭据。
func loadCreds() (Credentials, error) {
	raw, err := os.ReadFile(credsFile())
	if err != nil {
		return Credentials{}, err
	}
	var c Credentials
	if err := json.Unmarshal(raw, &c); err != nil {
		return Credentials{}, fmt.Errorf("凭据解析失败: %w", err)
	}
	if c.BotToken == "" || c.ILinkBotID == "" {
		return Credentials{}, fmt.Errorf("凭据文件 %s 字段不完整", credsFile())
	}
	return c, nil
}

// saveCreds:保存凭据(0600;.creds 目录不入库)。
func saveCreds(c Credentials) error {
	path := credsFile()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return err
	}
	fmt.Printf("凭据已保存:%s(含 bot_token,勿外泄)\n", path)
	return nil
}

// main:离线自检 → 确保已登录 → 常驻对话。
func main() {
	fLogin := flag.Bool("login", false, "仅重新扫码登录/换绑账号(完成后退出,不常驻)")
	flag.Parse()

	fmt.Println("== 1) 离线接口格式自检(本地 mock,不联网)==")
	offlineChecks()
	if fails > 0 {
		fmt.Printf("\n%d 项失败,先修格式再联调\n", fails)
		os.Exit(1)
	}
	fmt.Println("clawbot:离线格式自检全部通过")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 确保有登录凭据;无凭据时自动走扫码登录。
	creds, err := loadCreds()
	switch {
	case *fLogin:
		fmt.Println("== 扫码登录(换绑账号)==")
		creds, err = loginAndSave(ctx)
		if err != nil {
			fatalf("登录失败: %v", err)
		}
		fmt.Printf("登录成功 bot_id=%s;运行 go run . 进入常驻对话\n", creds.ILinkBotID)
		return
	case err != nil:
		fmt.Println("未检测到登录凭据,先扫码登录一次...")
		creds, err = loginAndSave(ctx)
		if err != nil {
			fatalf("登录失败: %v", err)
		}
	}

	// 常驻:收微信消息 → OpenAI 协议 → 回微信。
	fmt.Printf("== 2) 常驻监听中(bot=%s,Ctrl+C 退出)==\n", creds.ILinkBotID)
	fmt.Println("提示:先在微信里给「微信ClawBot」发一条消息激活(官方要求,否则收不到推送),之后即可一对一对话。")
	fmt.Printf("模型端点:%s 模型:%s\n", chatBaseURL, chatModel)
	serve(ctx, creds)
	fmt.Println("已退出(再次对话请重新运行 go run .)")
}

// loginAndSave:扫码登录并把凭据落盘。
func loginAndSave(ctx context.Context) (Credentials, error) {
	creds, err := liveLogin(ctx)
	if err != nil {
		return Credentials{}, err
	}
	if err := saveCreds(creds); err != nil {
		return Credentials{}, err
	}
	return creds, nil
}

// ---- 常驻对话 ----

// convo:一个"会话"(按 from_user_id 隔离)的近期上下文。
type convo struct {
	msgs []llmMsg
}

// add:追加一轮对话并裁剪到 convoTurns 轮。
func (cv *convo) add(role, content string) {
	cv.msgs = append(cv.msgs, llmMsg{Role: role, Content: content})
	total := convoTurns * 2
	if len(cv.msgs) > total {
		cv.msgs = cv.msgs[len(cv.msgs)-total:]
	}
}

// buildRequest:组装送模型的完整消息(system + 近期历史)。
func (cv *convo) buildRequest() []llmMsg {
	out := make([]llmMsg, 0, len(cv.msgs)+1)
	out = append(out, llmMsg{Role: "system", Content: chatSystem})
	return append(out, cv.msgs...)
}

// serve:常驻主循环(getupdates 长轮询;断线指数退避;-14 重置游标)。
func serve(ctx context.Context, creds Credentials) {
	c := newClient(creds)
	buf := ""
	backoff := 3 * time.Second
	// peers:按对方 ilink_user_id 隔离的会话上下文(常驻期间保留;清内存即重置)。
	peers := map[string]*convo{}

	for {
		resp, err := c.GetUpdates(ctx, buf)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			fmt.Printf("[warn] getupdates 失败,%s 后重试: %v\n", backoff, err)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return
			}
			if backoff < 60*time.Second {
				backoff *= 2
			}
			continue
		}
		backoff = 3 * time.Second

		// 会话过期:-14。游标非空→清空重连;游标本就空→凭据已失效,需重新 -login。
		if resp.ErrCode == errCodeSessionExpired {
			if buf != "" {
				buf = ""
				fmt.Println("[info] errcode=-14 会话过期,已重置游标重连")
			} else {
				fmt.Println("[warn] 凭据会话已失效,请运行 go run . -login 重新扫码")
				select {
				case <-time.After(60 * time.Second):
				case <-ctx.Done():
					return
				}
			}
			continue
		}
		if resp.Ret != 0 {
			fmt.Printf("[warn] getupdates ret=%d errmsg=%s(继续)\n", resp.Ret, resp.ErrMsg)
			continue
		}
		if resp.GetUpdatesBuf != "" {
			buf = resp.GetUpdatesBuf
		}

		for _, m := range resp.Msgs {
			if m.MessageType != msgTypeUser {
				continue // 只处理用户消息(bot 自己下发的消息不回环)
			}
			cv := peers[m.FromUserID]
			if cv == nil {
				cv = new(convo)
				peers[m.FromUserID] = cv
			}
			handleMsg(ctx, c, m, cv)
		}
	}
}

// handleMsg:单条用户消息 → 打字中 → 模型调用 → 回复(计入该会话上下文)。
func handleMsg(ctx context.Context, c *Client, m WeixinMessage, cv *convo) {
	fmt.Printf("[recv] %s\n", msgSummary(m))

	text := msgText(m)
	if text == "" {
		text = "…"
	}
	cv.add("user", text)

	_ = c.SendTyping(ctx, m.FromUserID, m.ContextToken, typingOn)

	reply, err := chatOnce(cv.buildRequest())
	if err != nil {
		reply = "（模型调用失败,本地日志见报错）" + truncate(err.Error(), 120)
		fmt.Printf("[FAIL] chat 失败: %v\n", err)
	}

	if err := c.SendText(ctx, m.FromUserID, reply, m.ContextToken); err != nil {
		fmt.Printf("[FAIL] 回复发送失败: %v\n", err)
		return
	}
	cv.add("assistant", reply)
	fmt.Printf("[sent] -> %s: %q\n", m.FromUserID, truncate(reply, 100))
	_ = c.SendTyping(ctx, m.FromUserID, m.ContextToken, typingOff)
}

// ---- 扫码登录 ----

// liveLogin:真实扫码登录流程(取码 → 终端渲染二维码 → 轮询确认 → 返回凭据)。
func liveLogin(ctx context.Context) (Credentials, error) {
	fctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	qr, err := fetchQRCode(fctx)
	cancel()
	if err != nil {
		return Credentials{}, err
	}

	fmt.Println("请用手机微信「扫一扫」下面的二维码完成授权(Ctrl+C 退出):")
	fmt.Println()
	renderQR(qr.QRCodeImgContent)
	fmt.Printf("\n二维码内容(兜底,可自行转码): %s\n\n", truncate(qr.QRCodeImgContent, 220))

	last := ""
	return pollQRStatus(ctx, qr.QRCode, func(s string) {
		if s != last {
			last = s
			switch s {
			case qrStatusScanned:
				fmt.Println("已扫码,请在手机上确认...")
			case qrStatusConfirmed:
				fmt.Println("已在手机确认!")
			case qrStatusExpired:
				fmt.Println("二维码已过期,请重新运行")
			}
		}
	})
}

// renderQR:把二维码内容文本渲染成终端二维码(半块字符,Windows Terminal/
// 各主流终端均可用;与 weclaw 相同的 qrterminal v3 渲染参数)。
func renderQR(content string) {
	qrterminal.GenerateWithConfig(content, qrterminal.Config{
		Level:          qrterminal.L,
		Writer:         os.Stdout,
		HalfBlocks:     true,
		BlackChar:      qrterminal.BLACK_BLACK,
		WhiteBlackChar: qrterminal.WHITE_BLACK,
		WhiteChar:      qrterminal.WHITE_WHITE,
		BlackWhiteChar: qrterminal.BLACK_WHITE,
		QuietZone:      1,
	})
}

// fatalf:打印错误并以非零码退出。
func fatalf(format string, args ...any) {
	fmt.Printf("[FAIL] "+format+"\n", args...)
	os.Exit(1)
}
