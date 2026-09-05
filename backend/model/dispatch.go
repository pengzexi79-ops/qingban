package model

// 批次投喂/短期记忆 角色级配置:companions.dispatch_settings JSON 列。
// 缺省值集中声明于本文件常量(配置默认值随实体走,避免散落);
// 老数据(无本列)不迁移,读取处零值回落缺省(见 ApplyDefaults),按默认行为运行。
// 设计拍板见 docs/batch_dispatch_design.md(§6 决策记录 #1/#4)。

// 缺省配置常量(角色级可被 DispatchSettings 覆盖;日期因子曲线见 AI 包 SummaryDateFactor)。
const (
	// DefaultDebounceSeconds:攒批静默窗(秒),默认 12。
	DefaultDebounceSeconds = 12
	// DefaultMaxBatchCount:攒批硬顶(条),默认 20。
	DefaultMaxBatchCount = 20
	// DefaultMaxBatchChars:攒批硬顶(字),默认 3000。
	DefaultMaxBatchChars = 3000
	// DefaultKvTTLMinutes:厂商 KV 缓存保守 TTL(分钟),默认 10。
	DefaultKvTTLMinutes = 10
	// DefaultCharRatio:短期记忆换算因子(压缩率),默认 0.12。
	DefaultCharRatio = 0.12
)

// DispatchSettings:单个 AI 成员的批次对话调度参数(内嵌于 Companion)。
type DispatchSettings struct {
	// DebounceSeconds 攒批静默窗(秒):距最后一条用户消息达到该值且期间无新消息 → 投喂。
	DebounceSeconds int `json:"debounceSeconds"`
	// MaxBatchCount 攒批硬顶(条):未投喂消息达到该条数立即投喂(防积压)。
	MaxBatchCount int `json:"maxBatchCount"`
	// MaxBatchChars 攒批硬顶(字):未投喂消息源文本达到该字数立即投喂(防积压)。
	MaxBatchChars int `json:"maxBatchChars"`
	// KvTTLMinutes 厂商 KV 缓存保守 TTL(分钟):消费后在其前 80% 时刻排定总结任务。
	// 说明:厂商不公开精确 TTL,此为保守估计;宁可早总结不可晚总结。
	KvTTLMinutes int `json:"kvTTLMinutes"`
	// CharRatio 短期记忆换算因子(压缩率):目标字数 = 日期因子 × 原字数 × 本值。
	CharRatio float64 `json:"charRatio"`
}

// ApplyDefaults:零值字段回落缺省(返回新值,不改原对象)。
// 说明:逐字段 if <=0 回落上方 Default* 常量;CharRatio 回落 DefaultCharRatio。
func (d DispatchSettings) ApplyDefaults() DispatchSettings {
	// if d.DebounceSeconds <= 0 { d.DebounceSeconds = DefaultDebounceSeconds }
	// if d.MaxBatchCount <= 0 { d.MaxBatchCount = DefaultMaxBatchCount }
	// if d.MaxBatchChars <= 0 { d.MaxBatchChars = DefaultMaxBatchChars }
	// if d.KvTTLMinutes <= 0 { d.KvTTLMinutes = DefaultKvTTLMinutes }
	// if d.CharRatio <= 0 { d.CharRatio = DefaultCharRatio }
	// return d
	return d // TODO(实现):见函数注释
}
