package initapp

// 阶段 ④:打开 SQLite 并迁移建表。
// 全局落点:core.DB(全部业务表唯一读写入口)。
// 注意:_foreign_keys=on 必须开启,否则外键级联约束不生效。

import "qingban/core"

// openDB:NewApp 第 4 步上半——打开连接(驱动在实现阶段引入 gorm.io/driver/sqlite)。
func openDB(cfg *core.Config) error {
	_ = cfg
	// db, err := gorm.Open(sqlite.Open(cfg.DBPath+"?_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=on"))
	// if err != nil { return err }
	// core.DB = db
	// db.Ping 自检:失败→degraded 状态继续(health 返回 dbOk=false,见 server/infra.go)
	return nil // TODO(实现):见函数注释;依赖 gorm.io/driver/sqlite 加入后落地
}

// migrateDB:NewApp 第 4 步下半——AutoMigrate 全实体 + 原生 SQL 补建。
// 实体清单与 docs/model_design.md 一致;原生 SQL 区见本文件尾注(实现时执行)。
func migrateDB() error {
	// 依赖顺序 AutoMigrate:
	//   api_configs/files/users/companions/groups/group_members/conversations/messages/
	//   message_files/message_mentions/memories/usage_records/
	//   config_kvs/member_cursors/chat_short_memories
	// 原生 SQL(逐条执行,均幂等):
	//   ① DROP TABLE IF EXISTS api_profiles;            // 旧实体已删除,不留数据
	//   ② DROP TABLE IF EXISTS rounds; DROP TABLE IF EXISTS round_speakers;
	//      ALTER TABLE groups DROP COLUMN last_round_at; // 轮次易失化:不落库(见 core/cache.go)
	//   ③ FTS5 虚拟表 + 触发器(消息/记忆全文检索):
	//      CREATE VIRTUAL TABLE IF NOT EXISTS fts_messages USING fts5(content, tokenize='unicode61')
	//      (+ 消息插入/删除的触发器同步维护,或用应用层双写)
	//   ④ sqlite-vec 向量表(可选):
	//      memory_vectors(memory_id INTEGER PRIMARY KEY, companion_id INTEGER, vector BLOB,
	//                     model TEXT, updated_at)   // 扩展缺失自动降级关键词检索
	//   ⑤ 补索引:usage_records(created_at)、memories(date 排序)等
	return nil // TODO(实现):见上方清单与原生 SQL 注释
}
