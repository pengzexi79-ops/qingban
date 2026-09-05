package main

// 微信 ClawBot(iLink)协议层:HTTP + JSON,把"青伴要做的 bot 后端"与微信侧打通。
// 格式依据(两份同源,均已真机验证):
//   1. 微信官方插件 @tencent-weixin/openclaw-weixin 的 README(Backend API Protocol 一节);
//   2. Go 参考实现 fastclaw-ai/weclaw 的 ilink 包(client.go/types.go/auth.go/monitor.go)。
// 线格式要点(写新代码/排查联调时对照):
//   - 登录前(GET,无鉴权):get_bot_qrcode 取二维码;get_qrcode_status 长轮询等待扫码确认;
//     confirmed 后返回 bot_token/ilink_bot_id/baseurl/ilink_user_id。
//   - 登录后(POST,JSON):baseurl + /ilink/bot/{getupdates|sendmessage|getconfig|sendtyping|getuploadurl};
//     所有请求带固定头 AuthorizationType: ilink_bot_token、Authorization: Bearer <bot_token>、
//     X-WECHAT-UIN(随机 uint32 十进制串的 base64);body 统一带 base_info 字段。
//   - getupdates 是"长轮询":带上次下发的 get_updates_buf 游标;服务端返回新游标待下次回传;
//     errcode=-14 表示会话过期,需清空游标重连(monitor.go 语义)。
//   - 回消息(msg 内 message_type=2 bot、message_state=2 完成态)必须原样回带对方消息的 context_token。

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// defaultBaseURL:未在登录凭据中拿到 baseurl 时的兜底网关(weclaw 同值)。
	defaultBaseURL = "https://ilinkai.weixin.qq.com"
	// 长轮询时间(服务端建议值 35s,请求侧再多给 5s 余量)。
	longPollTimeout = 40 * time.Second
	// sendMessage/getconfig/sendtyping 等普通调用超时。
	callTimeout = 15 * time.Second
)

// 登录两步走的端点(变量化:离线自检可改指向本地 mock 服务端)。
var (
	// qrCodeURL:申请登录二维码(bot_type=3 为 ClawBot 渠道,weclaw 实测值)。
	qrCodeURL = defaultBaseURL + "/ilink/bot/get_bot_qrcode?bot_type=3"
	// qrStatusURL:轮询扫码状态,后接 ?qrcode=<申请返回的 qrcode>。
	qrStatusURL = defaultBaseURL + "/ilink/bot/get_qrcode_status?qrcode="

	// bot_type 值域之外,消息/条目/状态常量均为线格式固定值(见官方文档)。
	msgTypeUser    = 1 // message_type:用户消息
	msgTypeBot     = 2 // message_type:bot 消息
	msgStateNew    = 0 // message_state:新消息
	msgStateFinish = 2 // message_state:完成(整条回复已就绪)
	itemTypeText   = 1 // item_list[].type:文本
	itemTypeImage  = 2
	itemTypeVoice  = 3
	itemTypeFile   = 4
	itemTypeVideo  = 5

	typingOn  = 1 // sendtyping status:正在输入
	typingOff = 2 // sendtyping status:停止输入

	// errCodeSessionExpired:getupdates 返回 errcode=-14 = 会话过期(清空游标重连)。
	errCodeSessionExpired = -14
)

// Credentials:扫码登录成功后获得的 bot 身份(bot_token 即请求鉴权密钥,勿外泄)。
type Credentials struct {
	BotToken    string `json:"bot_token"`
	ILinkBotID  string `json:"ilink_bot_id"`
	BaseURL     string `json:"baseurl"`
	ILinkUserID string `json:"ilink_user_id"`
}

// baseInfo:body 公共字段(官方/weclaw 均带;仅 getupdates 填 channel_version)。
type baseInfo struct {
	ChannelVersion string `json:"channel_version,omitempty"`
}

// ---- getupdates:长轮询拉取新消息 ----

type getUpdatesRequest struct {
	GetUpdatesBuf string   `json:"get_updates_buf"`
	BaseInfo      baseInfo `json:"base_info"`
}

type getUpdatesResponse struct {
	Ret     int    `json:"ret"`
	ErrCode int    `json:"errcode,omitempty"`
	ErrMsg  string `json:"errmsg,omitempty"`
	// Msgs:服务端推送的消息(见下方 WeixinMessage 结构)。
	Msgs []WeixinMessage `json:"msgs"`
	// GetUpdatesBuf:同步游标,下次 getupdates 原样回传(空=重置)。
	GetUpdatesBuf string `json:"get_updates_buf"`
	// LongPollingTimeoutMs:服务端建议的下一次长轮询超时(可忽略)。
	LongPollingTimeoutMs int `json:"longpolling_timeout_ms,omitempty"`
}

// WeixinMessage:会话里的一条消息(官方文档 WeixinMessage)。
type WeixinMessage struct {
	Seq          int           `json:"seq,omitempty"`
	MessageID    int64         `json:"message_id,omitempty"`
	FromUserID   string        `json:"from_user_id"`
	ToUserID     string        `json:"to_user_id"`
	MessageType  int           `json:"message_type"`  // 1=USER 2=BOT
	MessageState int           `json:"message_state"` // 0=NEW 1=GENERATING 2=FINISH
	ItemList     []MessageItem `json:"item_list"`
	// ContextToken:会话上下文令牌——回复时必须原样带回 sendmessage.msg.context_token。
	ContextToken string `json:"context_token"`
}

// MessageItem:单条消息的内容条目(一条消息可含多项)。
type MessageItem struct {
	Type      int        `json:"type"` // 1 文本 2 图片 3 语音 4 文件 5 视频
	TextItem  *TextItem  `json:"text_item,omitempty"`
	ImageItem *ImageItem `json:"image_item,omitempty"`
	VoiceItem *VoiceItem `json:"voice_item,omitempty"`
	FileItem  *FileItem  `json:"file_item,omitempty"`
	VideoItem *VideoItem `json:"video_item,omitempty"`
}

type TextItem struct {
	Text string `json:"text"`
}

// CDNMedia:媒体统一经 CDN 分发,内容 AES-128-ECB 加密(aes_key 为 base64)。
type CDNMedia struct {
	EncryptQueryParam string `json:"encrypt_query_param,omitempty"`
	AESKey            string `json:"aes_key,omitempty"`
	EncryptType       int    `json:"encrypt_type,omitempty"` // 1=AES-128-ECB
}

type ImageItem struct {
	URL     string    `json:"url,omitempty"`
	Media   *CDNMedia `json:"media,omitempty"`
	MidSize int       `json:"mid_size,omitempty"` // 密文尺寸
}

type VoiceItem struct {
	Media      *CDNMedia `json:"media,omitempty"`
	EncodeType int       `json:"encode_type,omitempty"` // 6=silk(微信语音常用)
	Text       string    `json:"text,omitempty"`        // 微信侧语音转写文本(免自行 STT)
}

type FileItem struct {
	Media    *CDNMedia `json:"media,omitempty"`
	FileName string    `json:"file_name,omitempty"`
	Len      string    `json:"len,omitempty"` // 明文尺寸(字符串形式)
}

type VideoItem struct {
	Media     *CDNMedia `json:"media,omitempty"`
	VideoSize int       `json:"video_size,omitempty"`
}

// ---- sendmessage:发送文本/媒体 ----

type sendMessageRequest struct {
	Msg      SendMsg  `json:"msg"`
	BaseInfo baseInfo `json:"base_info"`
}

type SendMsg struct {
	FromUserID   string        `json:"from_user_id"`  // 通常 = bot 自身 id(ilink_bot_id)
	ToUserID     string        `json:"to_user_id"`    // 对方 ilink_user_id
	ClientID     string        `json:"client_id"`     // 本次发送的客户端侧唯一 id(幂等/对账)
	MessageType  int           `json:"message_type"`  // bot 回话 = 2
	MessageState int           `json:"message_state"` // 完成态 = 2
	ItemList     []MessageItem `json:"item_list"`
	ContextToken string        `json:"context_token"` // 原样回带收到消息的令牌
}

type sendMessageResponse struct {
	Ret    int    `json:"ret"`
	ErrMsg string `json:"errmsg,omitempty"`
}

// ---- getconfig / sendtyping:输入状态(打字中)指示 ----

type getConfigRequest struct {
	ILinkUserID  string   `json:"ilink_user_id"`
	ContextToken string   `json:"context_token,omitempty"`
	BaseInfo     baseInfo `json:"base_info"`
}

type getConfigResponse struct {
	Ret          int    `json:"ret"`
	ErrMsg       string `json:"errmsg,omitempty"`
	TypingTicket string `json:"typing_ticket,omitempty"`
}

type sendTypingRequest struct {
	ILinkUserID  string   `json:"ilink_user_id"`
	TypingTicket string   `json:"typing_ticket"`
	Status       int      `json:"status"` // 1=正在输入 2=停止
	BaseInfo     baseInfo `json:"base_info"`
}

type sendTypingResponse struct {
	Ret    int    `json:"ret"`
	ErrMsg string `json:"errmsg,omitempty"`
}

// ---- getuploadurl:媒体上传前取 CDN 预签名参数(本子包仅做格式自检,真机联调先不发媒体) ----

type getUploadURLRequest struct {
	FileKey     string   `json:"filekey"`
	MediaType   int      `json:"media_type"` // 1=图片 2=视频 3=文件
	ToUserID    string   `json:"to_user_id"`
	RawSize     int      `json:"rawsize"`    // 明文尺寸
	RawFileMD5  string   `json:"rawfilemd5"` // 明文 MD5
	FileSize    int      `json:"filesize"`   // AES 加密后尺寸
	AESKey      string   `json:"aeskey"`     // 调用方自生成(base64 AES-128 key)
	NoNeedThumb bool     `json:"no_need_thumb"`
	BaseInfo    baseInfo `json:"base_info"`
}

type getUploadURLResponse struct {
	Ret           int    `json:"ret"`
	ErrMsg        string `json:"errmsg,omitempty"`
	UploadParam   string `json:"upload_param"`
	UploadFullURL string `json:"upload_full_url,omitempty"`
}

// ---- HTTP 客户端 ----

// Client:登录态 bot 客户端(每次调用生命周期内使用,复用同一个即可)。
type Client struct {
	BaseURL  string
	BotToken string
	BotID    string
	UIN      string // X-WECHAT-UIN(base64(random uint32 十进制))
	hc       *http.Client
}

// newClient:凭 creds 构造;baseurl 为空时兜底默认网关。
func newClient(creds Credentials) *Client {
	base := creds.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	return &Client{
		BaseURL:  base,
		BotToken: creds.BotToken,
		BotID:    creds.ILinkBotID,
		UIN:      newUIN(),
		hc:       &http.Client{},
	}
}

// GetUpdates:长轮询一次(超时 ~40s;网络超时属正常,调用方继续下一轮)。
func (c *Client) GetUpdates(ctx context.Context, buf string) (*getUpdatesResponse, error) {
	body := getUpdatesRequest{
		GetUpdatesBuf: buf,
		BaseInfo:      baseInfo{ChannelVersion: "1.0.0"},
	}
	ctx, cancel := context.WithTimeout(ctx, longPollTimeout)
	defer cancel()
	var out getUpdatesResponse
	if err := c.post(ctx, "/ilink/bot/getupdates", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendText:给某用户发一条文本(bot 回话;contextToken 传收到消息的令牌)。
func (c *Client) SendText(ctx context.Context, toUserID, text, contextToken string) error {
	req := sendMessageRequest{
		Msg: SendMsg{
			FromUserID:   c.BotID,
			ToUserID:     toUserID,
			ClientID:     newClientID(),
			MessageType:  msgTypeBot,
			MessageState: msgStateFinish,
			ItemList: []MessageItem{{
				Type:     itemTypeText,
				TextItem: &TextItem{Text: text},
			}},
			ContextToken: contextToken,
		},
		BaseInfo: baseInfo{},
	}
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	var out sendMessageResponse
	if err := c.post(ctx, "/ilink/bot/sendmessage", req, &out); err != nil {
		return err
	}
	if out.Ret != 0 {
		return fmt.Errorf("sendmessage 失败 ret=%d errmsg=%s", out.Ret, out.ErrMsg)
	}
	return nil
}

// GetConfig:取用户侧会话配置(typing_ticket 在此下发)。
func (c *Client) GetConfig(ctx context.Context, userID, contextToken string) (*getConfigResponse, error) {
	req := getConfigRequest{
		ILinkUserID:  userID,
		ContextToken: contextToken,
		BaseInfo:     baseInfo{},
	}
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	var out getConfigResponse
	if err := c.post(ctx, "/ilink/bot/getconfig", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendTyping:发"正在输入"状态;ticket 为空时先 getconfig 取。
func (c *Client) SendTyping(ctx context.Context, userID, contextToken string, status int) error {
	ticket := ""
	if cfg, err := c.GetConfig(ctx, userID, contextToken); err == nil {
		ticket = cfg.TypingTicket
	}
	if ticket == "" {
		// 拿不到 ticket 时打字指示可跳过,不阻塞主流程。
		return nil
	}
	req := sendTypingRequest{
		ILinkUserID:  userID,
		TypingTicket: ticket,
		Status:       status,
		BaseInfo:     baseInfo{},
	}
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	var out sendTypingResponse
	if err := c.post(ctx, "/ilink/bot/sendtyping", req, &out); err != nil {
		return err
	}
	if out.Ret != 0 {
		return fmt.Errorf("sendtyping 失败 ret=%d errmsg=%s", out.Ret, out.ErrMsg)
	}
	return nil
}

// GetUploadURL:取媒体 CDN 预签名(需要先自备 AES-128-ECB 密钥与密文尺寸)。
func (c *Client) GetUploadURL(ctx context.Context, req getUploadURLRequest) (*getUploadURLResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	var out getUploadURLResponse
	if err := c.post(ctx, "/ilink/bot/getuploadurl", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// post:JSON POST 并解析;非 200 或 ret!=0 之外的协议错误在此返回。
func (c *Client) post(ctx context.Context, path string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	c.setHeaders(req)
	return doJSON(c.hc, req, out)
}

// setHeaders:固定鉴权头(线格式要求,勿遗漏)。
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AuthorizationType", "ilink_bot_token")
	req.Header.Set("Authorization", "Bearer "+c.BotToken)
	req.Header.Set("X-WECHAT-UIN", c.UIN)
}

// doGet:无鉴权 GET(仅登录流程用,URL 为全量)。
func doGet(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	return doJSON(http.DefaultClient, req, out)
}

// doJSON:发请求并解析 JSON 响应(2xx 之外报错,脱敏:不回显请求原文)。
func doJSON(hc *http.Client, req *http.Request, out any) error {
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}
	return nil
}

// newUIN:随机 uint32 → 十进制串 → base64(请求头 X-WECHAT-UIN)。
func newUIN() string {
	var n uint32
	_ = binary.Read(rand.Reader, binary.LittleEndian, &n)
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", n)))
}

// newClientID:发送侧唯一 id(UUID v4 形态,随机;用于消息幂等/对账)。
func newClientID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("clawbot-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// msgText:取消息中的首段可见文本(文本条目;语音取微信侧转写)。
func msgText(m WeixinMessage) string {
	for _, it := range m.ItemList {
		switch {
		case it.Type == itemTypeText && it.TextItem != nil:
			return it.TextItem.Text
		case it.Type == itemTypeVoice && it.VoiceItem != nil && it.VoiceItem.Text != "":
			return "[语音] " + it.VoiceItem.Text
		case it.Type == itemTypeImage:
			return "[图片]"
		case it.Type == itemTypeFile && it.FileItem != nil:
			return "[文件] " + it.FileItem.FileName
		case it.Type == itemTypeVideo:
			return "[视频]"
		}
	}
	return ""
}

// msgSummary:消息摘要(日志/联调展示用)。
func msgSummary(m WeixinMessage) string {
	who := m.FromUserID
	if m.MessageType == msgTypeBot {
		who = "bot(" + m.ToUserID + ")"
	}
	return fmt.Sprintf("from=%s type=%d state=%d seq=%d ctx=%q text=%q",
		who, m.MessageType, m.MessageState, m.Seq, m.ContextToken, truncate(msgText(m), 60))
}

// ---- 登录(扫码授权) ----

type qrCodeResponse struct {
	// QRCode:二维码标识,轮询状态时带回。
	QRCode string `json:"qrcode"`
	// QRCodeImgContent:二维码内容文本(通常为 URL),用于渲染可扫图片。
	QRCodeImgContent string `json:"qrcode_img_content"`
}

type qrStatusResponse struct {
	Status      string `json:"status"` // wait/scaned/confirmed/expired
	BotToken    string `json:"bot_token"`
	ILinkBotID  string `json:"ilink_bot_id"`
	BaseURL     string `json:"baseurl"`
	ILinkUserID string `json:"ilink_user_id"`
}

const (
	qrStatusWait      = "wait"
	qrStatusScanned   = "scaned" // 官方拼写即 scaned(少一个 n)
	qrStatusConfirmed = "confirmed"
	qrStatusExpired   = "expired"
)

// fetchQRCode:申请一个登录二维码(第 1 步)。
func fetchQRCode(ctx context.Context) (*qrCodeResponse, error) {
	var out qrCodeResponse
	if err := doGet(ctx, qrCodeURL, &out); err != nil {
		return nil, fmt.Errorf("get_bot_qrcode: %w", err)
	}
	if out.QRCode == "" {
		return nil, fmt.Errorf("get_bot_qrcode 响应缺少 qrcode 字段")
	}
	return &out, nil
}

// pollQRStatus:轮询扫码状态直到 confirmed/expired(第 2 步;每轮是 ~40s 长轮询)。
func pollQRStatus(ctx context.Context, qrcode string, onStatus func(status string)) (Credentials, error) {
	endpoint := qrStatusURL + url.QueryEscape(qrcode)
	for {
		select {
		case <-ctx.Done():
			return Credentials{}, ctx.Err()
		default:
		}
		pollCtx, cancel := context.WithTimeout(ctx, longPollTimeout)
		var out qrStatusResponse
		err := doGet(pollCtx, endpoint, &out)
		cancel()
		if err != nil {
			if ctx.Err() != nil {
				return Credentials{}, ctx.Err()
			}
			continue // 超时属正常,继续轮
		}
		if onStatus != nil {
			onStatus(out.Status)
		}
		switch out.Status {
		case qrStatusConfirmed:
			return Credentials{
				BotToken:    out.BotToken,
				ILinkBotID:  out.ILinkBotID,
				BaseURL:     out.BaseURL,
				ILinkUserID: out.ILinkUserID,
			}, nil
		case qrStatusExpired:
			return Credentials{}, fmt.Errorf("二维码已过期")
		}
	}
}

// truncate:截断长文本(日志防刷屏)。
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
