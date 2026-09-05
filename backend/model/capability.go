package model

// 能力开关(Capabilities),用户级(globalCapabilities)与角色级(companion.capabilities)共用。
// 语义(BACKEND_HANDOFF §5):前端开关只表达"用户希望允许该能力";
// 后端仍需验证模型、供应商权限、额度与协议是否真实可执行——本阶段仅持久化+透传,能力执行属第二阶段。

// Capabilities:七类能力的启用开关集合。
type Capabilities struct {
	// Hearing:听觉(音频输入/ASR 转写)。第二阶段的 /capabilities/hearing 才真实生效。
	Hearing bool `json:"hearing"`
	// TTS:语音合成(文本转语音,返回可播放资源与时长)。
	TTS bool `json:"tts"`
	// VoiceClone:声音复刻(需单独授权/用途说明/可撤销,见 BACKEND_HANDOFF §5,不与 TTS 混为一个开关)。
	VoiceClone bool `json:"voiceClone"`
	// Vision:视觉(图片上传、理解结果与原图保留策略)。
	Vision bool `json:"vision"`
	// Video:视频(上传、转码、理解任务与异步状态)。
	Video bool `json:"video"`
	// ImageGeneration:文生图(内容审核、计费与结果资产)。
	ImageGeneration bool `json:"imageGeneration"`
	// WebSearch:联网搜索(来源、引用、时效与隐私边界)。
	WebSearch bool `json:"webSearch"`
	// ContentFilter:内容安全过滤(旧前端数据兼容字段,导入时若存在则保留)。
	ContentFilter bool `json:"contentFilter,omitempty"`
}
