# 青伴（亲伴）· 第一阶段本地 API 清单（冻结 v0.1）

> 版本：2026-09-04 ｜ 目标形态：**本地优先桌面应用（Wails 壳 + Vue 前端 + Go 本地后端）**
> 文档关系：字段全量以 `docs/BACKEND_HANDOFF.md` 为准；架构决策与完整接口面以 `docs/api接口.md`（草案 v0.3）为母版。本文档只做一件事：**冻结第一阶段要实现的一批 API 及其实现顺序**，冲突时以本文档的阶段划分优先。
> 产品名统一为 **青伴**（数据键继续兼容 `qinban_*`）。

## 1. 第一阶段范围一句话

> 让"本地 SQLite + 本地后端 + 单聊/群聊数据面 + AI 记忆 + 个人中心"在本机跑通，前端页面逐步从模拟数据切到 `127.0.0.1` 真实接口；朋友圈自动发布、多模态能力、云端备份、主动消息定时任务**不在第一阶段**。

### 1.1 本阶段做（本地闭环）
- 首次引导（本地空间初始化、前端演示数据一次性迁移导入）
- 我的资料与设置、API 配置（Ollama 默认 + 远程服务商，密钥本机加密）
- AI 通讯录（角色 CRUD）与统一会话列表（单聊 + 群聊聚合）
- 单聊消息：发送（同步 + SSE 流式）、历史、删除、已读、未读角标
- 群聊：CRUD、成员管理、手动触发一轮（调度先做"同步一轮"，异步轮次事件第二阶段再细做）
- 长期记忆：CRUD、语义检索（本地模型不可用时 FTS5 关键词兜底）、重新索引、候选确认
- 搜索：好友 / 消息 / 记忆 / 会话
- 本地文件（头像、聊天图片与附件），消息内嵌引用语法
- 本地实时事件 SSE（`/events`）与状态刷新
- 用量记录（本地真实落库，先记不分析）

### 1.2 本阶段不做（预留 501 占位或明确推迟）
| 推迟项 | 说明 |
|---|---|
| 朋友圈 `/moments*` | 前端页面保留模拟，第二阶段接真实接口 |
| 主动消息定时任务 / 纪念日提醒 | 只预留 `POST /companions/{id}/proactive` 占位（501），任务调度第二阶段 |
| 多模态能力 `/capabilities/*`（语音/视觉/视频/文生图/联网） | 第二阶段 |
| 云端加密备份/恢复 `§5.6` | 独立服务，第三阶段 |
| 系统通知 `notify/register` | 等 Wails 壳并入时一起做 |
| 远程推送、多端同步 | 长期范围外 |

## 2. 通用约定（第一阶段生效）

- **Base URL**：`http://127.0.0.1:8080/api/v1`。端口冲突自动顺延；实际地址由 Wails 注入前端 `window.__QINBAN_API__`（浏览器开发模式可用 `/api/v1` 同源）。
- **鉴权**：本地令牌头 `X-Local-Token`（后端启动时生成并写入数据目录；浏览器调试模式可留空）。无登录注册概念，单用户本地空间。
- **ID**：全部由后端生成，`uuid4` 短横线字符串（兼容现有 `companion-xxx` 这类可读前缀，id 本身不承载业务）。
- **分页**：游标 `?before=<lastId>&limit=`（`limit` 1~100，默认 20），返回 `{ items: [...], nextCursor: string|null }`；`before` 省略 = 最新一页。
- **错误**：统一 `{ code, message, traceId?, details? }`，HTTP 状态码语义化。第一阶段常用 `code`：`NOT_FOUND / VALIDATION_ERROR / RATE_LIMITED / PROACTIVE_DISABLED / COOLDOWN_ACTIVE / PROVIDER_ERROR / FILE_REFERENCED / NOT_IMPLEMENTED`。
- **幂等**：发送类接口支持 `Idempotency-Key` 头（同一 key 只执行一次）。
- **时间**：RFC3339；日期 `YYYY-MM-DD`。
- **图片/附件引用语法**：`![图片](fileId)` / `[文件名](fileId)`，先 `POST /files` 再写入消息 `content`；响应携带解析后的 `refs[]`。
- **文件根目录**：`{dataDir}/files/{fileId}`，与 SQLite 同级，随数据目录整体备份/迁移。

### 2.1 SSE 事件（`GET /events`，本地实时通道）
事件命名（`data:` 内为 JSON，`event:` 前缀如下）：
`new_message`（新消息，含他人/自己/AI）、`typing`、`read`、`presence`、`memory_candidates`（发送 done 后到达）、`round_start / round_message / round_end`（群聊轮次）、`settings_changed`。
第一阶段不推 `proactive_message / moment_published / backup_done`（等对应模块）。

## 3. 冻结 API 清单（按实现批次排序）

> 标注 `[P0]~[P3]`：P0 后端骨架第一批，P1 数据面，P2 AI 单聊链路，P3 群聊与收尾。
> 标注 `501`：本阶段返回 `NOT_IMPLEMENTED` 的占位接口（路由先注册，前端可安全调用）。

### 批次 P0 — 基础设施（先做，5 个）
| 方法 | 路径 | 用途 | 要点 |
|---|---|---|---|
| GET | `/health` | 健康检查 | `{status, apiVersion, dbOk, serverTime}` |
| GET | `/bootstrap` | 首次引导状态 | `{firstRun, userId, hasData, defaultApiProfile, dataVersion}`；`firstRun=false` 后前端直接进主界面 |
| POST | `/bootstrap/init` | 初始化本地空间 | 建库、种子 Ollama 默认 API 配置；`{mode: "empty"\|"import-demo"}`；body 可带 `importPayload` |
| GET | `/events` | 本地事件订阅（SSE） | 见 §2.1 |
| GET | `/refresh` | 状态刷新 | `?include=threads,presence`，返回会话未读/置顶/在线摘要，供下拉刷新 |

### 批次 P1 — 数据面（角色/会话/我的/文件/迁移，23 个）
#### 我的（4）
| 方法 | 路径 | 用途 | 要点 |
|---|---|---|---|
| GET | `/me` | 当前用户资料与设置 | `UserProfile + settings`（见 BACKEND_HANDOFF） |
| PATCH | `/me` | 改昵称/签名/画像/设置 | 分段更新；`settings` 对象整体覆盖子键 |
| POST | `/me/avatar` | 上传头像 | multipart，返回 `{fileId, url}` |
| GET | `/me/stats` | 个人页统计角标 | 收藏数、记忆条数、用量合计（一阶段简单计数即可） |

#### AI 通讯录 / 角色（7）
| 方法 | 路径 | 用途 | 要点 |
|---|---|---|---|
| GET | `/companions` | 角色列表 | `?q=&category=&hasUnread=&sort=`；返回列表含未读数 |
| POST | `/companions` | 创建角色 | 支持快速创建（name/persona/relationship 三字段即可），其余给默认值 |
| GET | `/companions/{companionId}` | 角色详情（编辑表单全量） | 含 persona/memorySettings/chatStyle/proactive/capabilities/apiProfileId |
| PATCH | `/companions/{companionId}` | 更新角色配置 | 各 tab 分段更新 |
| DELETE | `/companions/{companionId}` | 删除角色 | 级联：会话、消息、记忆、文件引用解除 |
| GET | `/companions/{companionId}/memories` | 该角色的长期记忆 | 同 `/memories?companionId=`，兼容保留 |
| GET | `/companions/{companionId}/proactive` `501` | 主动消息立即触发占位 | 先注册，返回 `NOT_IMPLEMENTED` |

#### 会话（统一列表，7）
| 方法 | 路径 | 用途 | 要点 |
|---|---|---|---|
| GET | `/conversations` | 会话列表（单聊+群聊聚合） | `?filter=all\|unread\|companion\|group&q=`；返回 `nextCursor` |
| GET | `/conversations/{conversationId}` | 会话摘要 | 顶栏信息、群成员摘要、公告 |
| PATCH | `/conversations/{conversationId}` | 置顶/免打扰/展示设置 | `{pinned, muted, ...}` |
| DELETE | `/conversations/{conversationId}` | 删除会话与历史 | 本地即删 |
| POST | `/conversations/{conversationId}/read` | 标记已读 | 返回 `{unreadTotal}` |
| GET | `/conversations/{conversationId}/messages` | 历史消息 | 升序返回 + `nextCursor`；消息含 `refs[]` |
| GET | `/search/threads` | 会话页顶栏搜索 | 命中名称/最后消息摘要；`?q=` |

#### 本地文件（4）
| 方法 | 路径 | 用途 | 要点 |
|---|---|---|---|
| POST | `/files` | 上传（图片/附件/头像） | multipart；`kind=image\|file\|voice\|video`、`scope=message\|moment\|avatar`；图片服务端生成缩略图 |
| GET | `/files/{fileId}` | 读取/下载 | 校验归属，`?thumbnail=1` 取缩略图 |
| DELETE | `/files/{fileId}` | 物理删除 | 被引用返回 `409 FILE_REFERENCED + refCount` |
| POST | `/files/purge-orphans` | 清理孤儿文件 | 仅调试/数据页用，返回删除数 |

#### 数据迁移（3）
| 方法 | 路径 | 用途 | 要点 |
|---|---|---|---|
| GET | `/data/export` | 全量导出 | `dataVersion=qingban_api_v1`；apiProfiles 脱敏；文件仅记录 fileId |
| POST | `/data/import` | 导入 | 兼容本格式 + 前端演示 `qinban_frontend_v4/v3/v2`；返回迁移统计 `{companions, conversations, messages, memories, moments?}` |
| DELETE | `/data` | 清空本地业务数据 | `?confirm=true`，不可恢复 |

### 批次 P2 — 消息与 AI 单聊链路（7）
| 方法 | 路径 | 用途 | 要点 |
|---|---|---|---|
| POST | `/conversations/{conversationId}/messages` | 发送消息 | body `MessageCreate{content, contentType?, refs?}`；带 `Accept: text/event-stream` 走 SSE，否则同步返回 `MessageSendResult{userMessage, assistantMessage, memoryCandidates}`（见下） |
| DELETE | `/conversations/{conversationId}/messages/{messageId}` | 删除单条 | 本地删除 + 解除引用 |
| DELETE | `/conversations/{conversationId}/messages` | 清空历史 | `?confirm=true` |
| GET | `/search/messages` | 跨会话消息搜索 | `?q=&conversationId=&before=&limit=`；命中片段 + 可回跳位置（conversationId+messageId） |
| GET | `/usage/summary` | 用量汇总 | 前端 API 损耗页改读本地真实记录；`?from=&to=` |
| GET | `/usage/records` | 用量明细 | `?before=&model=&capability=` |
| GET | `/usage/trend` | 近 7 日趋势 | 替换页面占位图，`?days=` |

**AI 回复 SSE 事件序列（发送带流式头时）**：`typing` → `delta`（增量正文，可含 markdown 片段）→ `message`（完整 assistant 消息）→ `memory_candidates`（可能为空数组）→ `done`。同步模式下一次返回全部。

### 批次 P3 — 群聊 + 长期记忆 + API 配置（16）
#### 群聊（8）
| 方法 | 路径 | 用途 | 要点 |
|---|---|---|---|
| GET | `/groups` | 群列表 | 成员摘要 + strategy |
| POST | `/groups` | 建群 | `{name, avatar?, announcement, memberIds(≥2), strategy{enabled, mode, cooldownSeconds, maxSpeakers, order}}` |
| GET | `/groups/{groupId}` | 群详情 | |
| PATCH | `/groups/{groupId}` | 改名/头像/公告/策略 | |
| DELETE | `/groups/{groupId}` | 解散群 | 消息归档可导出 |
| POST | `/groups/{groupId}/members` | 添加成员 | `{memberIds[]}` |
| DELETE | `/groups/{groupId}/members/{companionId}` | 移出成员 | 少于 2 人禁止 |
| POST | `/groups/{groupId}/rounds` | 触发一轮群聊 | **第一阶段同步执行一轮**：选人→逐个调用→收尾，直接经 `/events` 推 `round_*`；返回 `roundId` |

#### 长期记忆（6）
| 方法 | 路径 | 用途 | 要点 |
|---|---|---|---|
| GET | `/memories` | 记忆数据台列表 | `?companionId=&q=&type=&before=` |
| POST | `/memories` | 手动添加 | `{companionId?, type, title, content, importance?, date?}` |
| PATCH | `/memories/{memoryId}` | 编辑 | 支持 `{status:"confirmed"}` 确认候选 |
| DELETE | `/memories/{memoryId}` | 删除并移除向量/索引 | |
| POST | `/memories/search` | 语义检索 | `{query, companionId?, topK=8, threshold=0.65, include[]}` → `{traceId, hits[{memoryId,score,...}], summary}`；本地无 embedding 时自动降级关键词检索，hits 带 `method:"keyword"` |
| POST | `/memories/{memoryId}/reindex` | 重建 embedding | `pending → indexed` |

#### API 配置（5）
| 方法 | 路径 | 用途 | 要点 |
|---|---|---|---|
| GET | `/api-profiles` | 列表 | 脱敏 `{secretConfigured, maskedKey}` |
| POST | `/api-profiles` | 新建 | `apiKey` 本机加密存储（os keychain/加密文件） |
| PATCH | `/api-profiles/{id}` | 修改 | `apiKey` 缺省不更新 |
| DELETE | `/api-profiles/{id}` | 删除 | 至少保留一套；同步解除角色绑定 |
| POST | `/api-profiles/{id}/test` ｜ `GET /api-profiles/{id}/models` | 连通测试 / 拉模型目录 | 由本地后端代理，浏览器不直连 |

> 群聊轮次记录页 `GET /groups/{groupId}/rounds` 第一阶段返回空数组占位，异步轮次记录第二阶段补。

## 4. 阶段一后端数据表（SQLite，供实现参考）

`users(1行)`、`settings`（并入 users JSON 或独立 kv）、`api_profiles`、`companions`、`groups`、`group_members`、`group_strategy`（可并入 groups）、`conversations`（仅存展示摘要派生缓存，主数据见下）、`messages`、`memories`、`moment 相关表（推迟）`、`files`、`usage_records`、`kv`（幂等键、迁移状态、设置）。
消息与记忆实体字段一律按 `BACKEND_HANDOFF.md`；`conversations` 的未读数/最后一条消息由本地派生，写入缓存表即可（单机无一致性问题）。

## 5. 实现顺序与验收（可直接拆分给 AI 开发）

1. **P0**：健康检查 + bootstrap + SSE 通道 → 后端进程能起、前端能连、能收事件。
2. **P1**：建表迁移 → `/data/import`（迁移演示 JSON）→ `/me`、`/companions`、`/conversations`、`/files` 数据面 CRUD。**验收：现有演示数据导入后，前端会话页、通讯录、我的页展示与模拟数据一致，可增删改。**
3. **P2**：消息发送链路（先同步后 SSE）→ 用量落库。**验收：浏览器 + Ollama（或任意 OpenAI 兼容端点）完成一次"发消息→AI 回复→记忆候选"闭环。**
4. **P3**：记忆检索/API 配置/群聊一轮。**验收：记忆数据台可检索；群聊点"触发一轮"能看到多个角色依次发言。**
5. 前端 `frontend/js/app.js` 中把各模拟方法按接口逐一替换，保留 `fallback` 本地兜底文案。

## 6. 下一步文档动作（本清单通过后执行）

- 按本清单刷新 `docs/openapi.yaml`（合入 §3 全部端点，标注 `501`）。
- 重建 `bruno/` 集合（分 P0~P3 目录 + 示例 JSON body，供后端与前端联调）。
- `backend/` 骨架演进为：`server`（路由注册，501 占位先注册）→ `model`（GORM 实体）→ `core`（本地令牌、SSE hub、幂等）→ `wails`（壳接入时再动）。
