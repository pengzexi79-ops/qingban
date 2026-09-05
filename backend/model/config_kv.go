package model

// 键值配置:config_kvs。
// 承载引导状态、数据版本、默认模型配置引用等配置类键值;
// key 明文(唯一索引,用于定位),value **整列加密**——由服务层经本地主密钥
// (utils.SecretBox,AES-256-GCM)加密为 base64 文本后写入,读取经"解密后使用"封装,
// 内存外不出现明文。值列不参与数据库内匹配,涉及值语义的读取一律解密后判断。

import "gorm.io/gorm"

// 键常量(集中声明,避免字符串散落导致键冲突)。
const (
	// KVBootstrapDone:引导完成标记("1"=已初始化;引导接口据此返回 firstRun)。
	KVBootstrapDone = "k:bootstrap:done"
	// KVDataVersion:数据版本(写入 qingban_api_v1;导入前端数据时记录来源版本)。
	KVDataVersion = "k:migrate:data_version"
	// KVDefaultModelConfigID:默认模型配置 id(model_configs.id;删除保护用)。
	KVDefaultModelConfigID = "k:model_configs:default_id"
	// KVMigratePrefix:导入迁移统计(保留最近一次,便于回显)。
	KVMigratePrefix = "k:migrate:last_stats"
)

// ConfigKV:键值配置行。
type ConfigKV struct {
	gorm.Model
	// Key:业务键(唯一;见上方常量)。
	Key string `json:"key" gorm:"size:255;uniqueIndex;not null"`
	// Value:值(整列加密存储:SecretBox 密文 base64;服务层读写)。
	Value string `json:"-" gorm:"type:text;not null"`
}
