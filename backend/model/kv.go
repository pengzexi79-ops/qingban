package model

// KV 存储实体:kvs 表。
// 定位(PHASE1 §4 表清单的 kv):迁移状态、引导状态、默认 API 配置 id 等零散键值,
// 以及可选的持久化幂等/事件序号。键以 k:<域>:<名字> 命名,值统一 JSON 文本。

import "time"

// KV 键常量(集中声明,避免字符串散落导致键冲突)。
const (
	// KVBootstrapDone:引导完成标记("1"=已初始化;GET /bootstrap 据此返回 firstRun)。
	KVBootstrapDone = "k:bootstrap:done"
	// KVDataVersion:当前数据版本(写入 qingban_api_v1;导入前端数据时记录来源版本)。
	KVDataVersion = "k:migrate:data_version"
	// KVDefaultProfileID:默认 API 配置 id(首次创建/种子 Ollama 配置时写入;删除保护用)。
	KVDefaultProfileID = "k:api_profiles:default_id"
	// KVMigratePrefix:导入迁移统计(可选保留最近一次,便于前端回显)。
	KVMigratePrefix = "k:migrate:last_stats"
)

// KV:键值行。表:kvs。
type KV struct {
	// Key:键名(主键,见上方常量)。
	Key string `json:"key" gorm:"primaryKey"`
	// Value:JSON 编码的值(字符串直接存原文;对象存 JSON 文本)。
	Value string `json:"value" gorm:"type:text"`
	// UpdatedAt:最近写入时间(调试查看)。
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName:表名(kvs)。
func (KV) TableName() string { return "kvs" }
