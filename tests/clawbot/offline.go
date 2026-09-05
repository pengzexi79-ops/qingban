package main

// 离线接口格式自检:在本地起 httptest mock「微信 ClawBot(iLink)服务端」,
// 用本包 client.go 的协议层真实打一遍,断言请求/响应线格式与官方一致。
// 不需要联网、不需要 bot_token,任何机器都能跑;真机联调失败时可先回到这里排查是
// "格式错了"还是"环境/授权问题"。

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"
)

// offlineChecks:全部离线断言(顺序执行;mock 服务全部本地起停)。
func offlineChecks() {
	mock := serveAPIMock()
	defer mock.Close()

	// 登录两个端点改指本地 mock(实网域名不可在离线自检里访问)。
	oldQR, oldStatus := qrCodeURL, qrStatusURL
	defer func() { qrCodeURL, qrStatusURL = oldQR, oldStatus }()
	qrCodeURL = mock.URL + "/ilink/bot/get_bot_qrcode?bot_type=3"
	qrStatusURL = mock.URL + "/ilink/bot/get_qrcode_status?qrcode="

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// ---- 0) 公共头/X-WECHAT-UIN 生成格式 ----
	uin := newUIN()
	raw, err := base64.StdEncoding.DecodeString(uin)
	check("wire.uin.base64", err == nil && uin != "", fmt.Sprint(err))
	check("wire.uin.digits", err == nil && digitsOnly(string(raw)) && len(raw) <= 10, string(raw))
	check("wire.uin.vary", newUIN() != uin, "两次随机应不同")

	// ---- 1) 扫码登录流程(申请二维码 → wait/scaned/confirmed 轮询) ----
	qr, err := fetchQRCode(ctx)
	check("login.qr.fetch", err == nil && qr.QRCode == "qr-mock-1" && qr.QRCodeImgContent != "", fmt.Sprint(err, qr))
	var statusLog []string
	creds, err := pollQRStatus(ctx, qr.QRCode, func(s string) { statusLog = append(statusLog, s) })
	check("login.qr.statuses", err == nil && strings.Join(statusLog, ",") == "wait,scaned,confirmed",
		fmt.Sprint(statusLog, err))
	check("login.creds", err == nil && creds.BotToken == "mock-bot-token" &&
		creds.ILinkBotID == "bot_mock@im.bot" && creds.ILinkUserID == "wxid-bot-1" &&
		creds.BaseURL == mock.URL, fmt.Sprint(creds))
	// 登录返回的 baseurl 必须被客户端采用(线上网关未必是默认域名)。
	c := newClient(creds)
	check("login.baseurl.used", c.BaseURL == mock.URL, c.BaseURL)

	// ---- 2) getupdates:空游标首轮 + 游标回传第二轮拿到消息推送 ----
	r1, err := c.GetUpdates(ctx, "")
	check("getupdates.poll.empty", err == nil && r1.Ret == 0 && len(r1.Msgs) == 0 && r1.GetUpdatesBuf == "cur-1",
		fmt.Sprint(err, r1))
	check("getupdates.request.json", mockLast["getupdates"] ==
		`{"get_updates_buf":"","base_info":{"channel_version":"1.0.0"}}`, mockLast["getupdates"])
	check("getupdates.headers", mockHdr["getupdates"].Get("AuthorizationType") == "ilink_bot_token" &&
		mockHdr["getupdates"].Get("Authorization") == "Bearer mock-bot-token" &&
		mockHdr["getupdates"].Get("Content-Type") == "application/json" &&
		mockHdr["getupdates"].Get("X-WECHAT-UIN") == c.UIN,
		fmt.Sprint(mockHdr["getupdates"]))
	r2, err := c.GetUpdates(ctx, r1.GetUpdatesBuf)
	check("getupdates.request.buf", mockLast["getupdates"] ==
		`{"get_updates_buf":"cur-1","base_info":{"channel_version":"1.0.0"}}`, mockLast["getupdates"])
	check("getupdates.poll.msg", err == nil && len(r2.Msgs) == 1 && r2.GetUpdatesBuf == "cur-2",
		fmt.Sprint(err, r2))

	// ---- 3) 收到消息解析(夹具故意多带 create_time_ms/session_id,应被容忍) ----
	m := r2.Msgs[0]
	check("msg.fields", m.Seq == 17 && m.MessageID == 9912 && m.FromUserID == "wx_u88@im.wechat" &&
		m.ToUserID == "bot_mock@im.bot" && m.MessageType == msgTypeUser && m.MessageState == msgStateNew &&
		m.ContextToken == "ctx-9", fmt.Sprint(m))
	check("msg.text", msgText(m) == "你好 ClawBot", msgText(m))

	// ---- 4) 回音:sendmessage 线格式 + context_token 原样回带 ----
	err = c.SendText(ctx, m.FromUserID, "收到!", m.ContextToken)
	check("sendmessage.ret0", err == nil, fmt.Sprint(err))
	wantSend := `{"msg":{"from_user_id":"bot_mock@im.bot","to_user_id":"wx_u88@im.wechat",` +
		`"client_id":"<uuid>","message_type":2,"message_state":2,` +
		`"item_list":[{"type":1,"text_item":{"text":"收到!"}}],"context_token":"ctx-9"},"base_info":{}}`
	got := normalizeUUID(mockLast["sendmessage"])
	check("sendmessage.request.json", got == wantSend, "got:  "+got+"\nwant: "+wantSend)
	check("sendmessage.headers", mockHdr["sendmessage"].Get("Authorization") == "Bearer mock-bot-token",
		mockHdr["sendmessage"].Get("Authorization"))
	err = c.SendText(ctx, m.FromUserID, "trigger-error", m.ContextToken)
	check("sendmessage.retNonzero", err != nil && strings.Contains(err.Error(), "ret=3"), fmt.Sprint(err))

	// ---- 5) 打字指示:getconfig 取 ticket → sendtyping(1 开 / 2 停;失败上抛) ----
	cfg, err := c.GetConfig(ctx, m.FromUserID, m.ContextToken)
	check("getconfig.response", err == nil && cfg.TypingTicket == "tk-1", fmt.Sprint(err, cfg))
	check("getconfig.request.json", mockLast["getconfig"] ==
		`{"ilink_user_id":"wx_u88@im.wechat","context_token":"ctx-9","base_info":{}}`, mockLast["getconfig"])
	err = c.SendTyping(ctx, m.FromUserID, m.ContextToken, typingOn)
	check("sendtyping.on", err == nil, fmt.Sprint(err))
	check("sendtyping.request.json", mockLast["sendtyping"] ==
		`{"ilink_user_id":"wx_u88@im.wechat","typing_ticket":"tk-1","status":1,"base_info":{}}`,
		mockLast["sendtyping"])
	err = c.SendTyping(ctx, "fail-user", m.ContextToken, typingOff)
	check("sendtyping.retNonzero", err != nil && strings.Contains(err.Error(), "ret=1"), fmt.Sprint(err))

	// ---- 6) getuploadurl:媒体上传预签名(仅校验请求线格式;CDN 密文上传不做) ----
	up, err := c.GetUploadURL(ctx, getUploadURLRequest{
		FileKey: "f-1", MediaType: 1, ToUserID: m.FromUserID,
		RawSize: 100, RawFileMD5: "d41d8cd98f00b204e9800998ecf8427e", FileSize: 112,
		AESKey: "aGVsbG8xMjM0NTY3OA==",
	})
	check("getuploadurl.response", err == nil && up.Ret == 0 && up.UploadParam == "upload-param-1",
		fmt.Sprint(err, up))
	check("getuploadurl.request.json", mockLast["getuploadurl"] ==
		`{"filekey":"f-1","media_type":1,"to_user_id":"wx_u88@im.wechat","rawsize":100,`+
			`"rawfilemd5":"d41d8cd98f00b204e9800998ecf8427e","filesize":112,`+
			`"aeskey":"aGVsbG8xMjM0NTY3OA==","no_need_thumb":false,"base_info":{}}`,
		mockLast["getuploadurl"])

	// ---- 7) 会话过期 errcode=-14:字段透出;清空游标后可恢复 ----
	expired := serveExpiredMock()
	defer expired.Close()
	ec := newClient(Credentials{BotToken: "bt-e", ILinkBotID: "b@x", BaseURL: expired.URL})
	re, err := ec.GetUpdates(ctx, "cur-boom")
	check("expired.errcode", err == nil && re.Ret == 0 && re.ErrCode == errCodeSessionExpired && len(re.Msgs) == 0,
		fmt.Sprint(err, re))
	rOK, err := ec.GetUpdates(ctx, "")
	check("expired.recover", err == nil && rOK.Ret == 0 && rOK.GetUpdatesBuf == "cur-ok", fmt.Sprint(err, rOK))
}

// ---- mock 记录(单线程顺序调用,无需加锁) ----

var (
	// mockLast:各端点最近一次收到的请求体原文(键=端点短名)。
	mockLast = map[string]string{}
	// mockHdr:各端点最近一次请求的完整头。
	mockHdr = map[string]http.Header{}
)

// record:记录请求体与头(各业务端点统一走这里)。
func record(ep string, r *http.Request) string {
	raw, _ := io.ReadAll(r.Body)
	mockLast[ep] = string(raw)
	mockHdr[ep] = r.Header.Clone()
	return string(raw)
}

// writeJSON:按 JSON 写响应。
func writeJSON(rw http.ResponseWriter, v any) {
	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(v)
}

// serveAPIMock:模拟 ClawBot 服务端场景 A(登录 + 全部业务端点 + 失败分支)。
func serveAPIMock() *httptest.Server {
	var srv *httptest.Server
	qrCalls, upCalls := 0, 0
	mux := http.NewServeMux()

	mux.HandleFunc("/ilink/bot/get_bot_qrcode", func(rw http.ResponseWriter, r *http.Request) {
		writeJSON(rw, map[string]any{
			"qrcode":             "qr-mock-1",
			"qrcode_img_content": "https://weixin.qq.com/q/mock-1",
		})
	})
	mux.HandleFunc("/ilink/bot/get_qrcode_status", func(rw http.ResponseWriter, r *http.Request) {
		qrCalls++
		switch qrCalls {
		case 1:
			writeJSON(rw, map[string]any{"status": "wait"})
		case 2:
			writeJSON(rw, map[string]any{"status": "scaned"})
		default:
			// confirmed 时下发 baseurl=本 mock(验证客户端采用登录下发的网关)。
			writeJSON(rw, map[string]any{"status": "confirmed", "bot_token": "mock-bot-token",
				"ilink_bot_id": "bot_mock@im.bot", "ilink_user_id": "wxid-bot-1", "baseurl": srv.URL})
		}
	})
	mux.HandleFunc("/ilink/bot/getupdates", func(rw http.ResponseWriter, r *http.Request) {
		upCalls++
		body := record("getupdates", r)
		if mockHdr["getupdates"].Get("Authorization") != "Bearer mock-bot-token" {
			http.Error(rw, "bad auth", http.StatusUnauthorized)
			return
		}
		switch upCalls {
		case 1:
			writeJSON(rw, map[string]any{"ret": 0, "msgs": []any{},
				"get_updates_buf": "cur-1", "longpolling_timeout_ms": 35000})
		case 2:
			writeJSON(rw, map[string]any{"ret": 0,
				"msgs":            []any{userMsgJSON("wx_u88@im.wechat", "你好 ClawBot", "ctx-9", 17, 9912)},
				"get_updates_buf": "cur-2"})
		default:
			_ = body
			writeJSON(rw, map[string]any{"ret": 0,
				"msgs":            []any{userMsgJSON("wx_u88@im.wechat", "第二条", "ctx-10", 18, 9913)},
				"get_updates_buf": "cur-3"})
		}
	})
	mux.HandleFunc("/ilink/bot/sendmessage", func(rw http.ResponseWriter, r *http.Request) {
		raw := record("sendmessage", r)
		if strings.Contains(raw, "trigger-error") {
			writeJSON(rw, map[string]any{"ret": 3, "errmsg": "ticket invalid"})
			return
		}
		writeJSON(rw, map[string]any{"ret": 0, "errmsg": ""})
	})
	mux.HandleFunc("/ilink/bot/getconfig", func(rw http.ResponseWriter, r *http.Request) {
		record("getconfig", r)
		writeJSON(rw, map[string]any{"ret": 0, "typing_ticket": "tk-1"})
	})
	mux.HandleFunc("/ilink/bot/sendtyping", func(rw http.ResponseWriter, r *http.Request) {
		raw := record("sendtyping", r)
		if strings.Contains(raw, "fail-user") {
			writeJSON(rw, map[string]any{"ret": 1, "errmsg": "bad ticket"})
			return
		}
		writeJSON(rw, map[string]any{"ret": 0})
	})
	mux.HandleFunc("/ilink/bot/getuploadurl", func(rw http.ResponseWriter, r *http.Request) {
		record("getuploadurl", r)
		writeJSON(rw, map[string]any{"ret": 0, "upload_param": "upload-param-1",
			"upload_full_url": "https://cdn.example/up"})
	})

	srv = httptest.NewServer(mux)
	return srv
}

// serveExpiredMock:模拟场景 B(errcode=-14 会话过期;空游标即恢复)。
func serveExpiredMock() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ilink/bot/getupdates", func(rw http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if strings.Contains(string(raw), "cur-boom") {
			writeJSON(rw, map[string]any{"ret": 0, "errcode": -14,
				"errmsg": "session timeout", "msgs": []any{}})
			return
		}
		writeJSON(rw, map[string]any{"ret": 0, "msgs": []any{}, "get_updates_buf": "cur-ok"})
	})
	return httptest.NewServer(mux)
}

// userMsgJSON:按官方 WeixinMessage 结构造的"用户消息"夹具;故意多带
// create_time_ms/session_id 两个本客户端未建模字段,验证"未知字段向后兼容"。
func userMsgJSON(from, text, ctxToken string, seq, msgID int) map[string]any {
	return map[string]any{
		"seq":            seq,
		"message_id":     msgID,
		"create_time_ms": 1700000000000 + int64(seq),
		"session_id":     "session-" + ctxToken,
		"from_user_id":   from,
		"to_user_id":     "bot_mock@im.bot",
		"message_type":   msgTypeUser,
		"message_state":  msgStateNew,
		"item_list":      []any{map[string]any{"type": itemTypeText, "text_item": map[string]any{"text": text}}},
		"context_token":  ctxToken,
	}
}

// uuidRe:client_id 形态(UUID v4)。
var uuidRe = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}`)

// normalizeUUID:把随机 client_id 归一为占位符,便于整串比对。
func normalizeUUID(s string) string {
	return uuidRe.ReplaceAllString(s, "<uuid>")
}

// digitsOnly:是否纯 ASCII 数字。
func digitsOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
