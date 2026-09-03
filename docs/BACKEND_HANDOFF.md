# 亲伴前端 → 后端研发交接

**基线日期：2026-09-03**

**前端数据版本：`qinban_frontend_v4`（兼容旧键 `qingban_frontend_v4/v3/v2`）**
**当前性质：可点击前端原型，不代表服务端能力已实现**

## 1. 交接目标

后端应尽量保持现有页面、字段含义和用户操作路径不变，把浏览器本地模拟逐步替换成真实服务。前端已经提供业务入口和状态展示，后端主要补齐数据持久化、模型调用、任务调度、权限、安全和用量统计。

## 2. 核心实体

### UserProfile

```json
{
  "id": "user-id",
  "nickname": "林林",
  "avatarUrl": "https://...",
  "signature": "把日子过成被记住的样子",
  "persona": "用户希望如何被理解和回应",
  "preferences": {
    "theme": "light",
    "fontSize": "comfortable",
    "bubbleRadius": 18,
    "messageGap": 14
  }
}
```

### Companion

```json
{
  "id": "companion-id",
  "name": "沐沐",
  "initial": "沐",
  "avatarUrl": "https://...",
  "color": "#776ee8",
  "category": "温柔陪伴",
  "tagline": "会记得你小习惯的温柔朋友",
  "apiProfileId": "api-profile-id",
  "persona": {
    "identity": "角色身份",
    "relationship": "与用户的关系",
    "personality": "性格描述",
    "speakingStyle": "表达风格",
    "boundaries": "关系和安全边界",
    "forbiddenTopics": "禁用行为与内容"
  },
  "memorySettings": {
    "enabled": true,
    "mode": "hybrid",
    "summaryMode": "rolling",
    "timeRangeDays": 365,
    "searchThreshold": 0.65,
    "maxItems": 12
  },
  "chatStyle": {
    "markdown": true,
    "streaming": true,
    "typing": true,
    "splitMessages": true,
    "replyDelay": 650,
    "bubbleStyle": "soft"
  },
  "proactive": {
    "enabled": true,
    "start": "09:00",
    "end": "22:30",
    "frequency": "balanced",
    "minMinutes": 45,
    "maxMinutes": 240,
    "dailyLimit": 4,
    "avoidBusyTime": true
  },
  "capabilities": {
    "hearing": false,
    "tts": true,
    "voiceClone": false,
    "vision": true,
    "video": false,
    "imageGeneration": false,
    "webSearch": false
  }
}
```

### Group

```json
{
  "id": "group-id",
  "name": "晚风茶话会",
  "avatarUrl": "https://...",
  "memberIds": ["companion-a", "companion-b"],
  "announcement": "群聊目的和调度说明",
  "strategy": {
    "enabled": true,
    "mode": "random",
    "order": "balanced",
    "cooldownSeconds": 18,
    "maxSpeakers": 2
  }
}
```

### Message

```json
{
  "id": "message-id",
  "conversationType": "companion",
  "conversationId": "companion-or-group-id",
  "role": "assistant",
  "senderId": "companion-id",
  "content": "消息正文",
  "contentType": "text",
  "timestamp": "2026-09-03T10:00:00+08:00",
  "proactive": false,
  "fallback": false,
  "usageId": "usage-record-id"
}
```

群聊中的 AI 消息必须提供 `senderId`；用户消息的 `senderId` 为当前用户。流式返回建议使用 SSE，并在结束事件中返回最终 `messageId`、用量和记忆处理状态。

### MemoryRecord

```json
{
  "id": "memory-id",
  "companionId": "companion-id",
  "type": "preference",
  "title": "喜欢雨天听歌",
  "content": "标准化后的记忆内容",
  "source": "用户确认",
  "sourceMessageId": "message-id",
  "importance": 0.88,
  "embeddingStatus": "indexed",
  "createdAt": "2026-09-03T10:00:00+08:00",
  "updatedAt": "2026-09-03T10:00:00+08:00"
}
```

向量检索建议返回：

```json
{
  "traceId": "trace-id",
  "query": "用户当前消息",
  "topK": 8,
  "threshold": 0.65,
  "hits": [
    {
      "memoryId": "memory-id",
      "score": 0.87,
      "title": "喜欢雨天听歌",
      "content": "..."
    }
  ],
  "summary": "注入模型前的压缩摘要"
}
```

### Moment

```json
{
  "id": "moment-id",
  "authorType": "companion",
  "authorId": "companion-id",
  "content": "动态正文",
  "media": [],
  "visibility": "all",
  "createdAt": "2026-09-03T10:00:00+08:00",
  "liked": false,
  "likeCount": 0,
  "commentCount": 0,
  "saved": false
}
```

朋友圈点赞、评论和收藏应拆成独立资源，不把完整用户列表长期塞入动态记录。

### ApiProfile

```json
{
  "id": "api-profile-id",
  "name": "主对话配置",
  "provider": "自定义服务商",
  "region": "自定义",
  "protocol": "openai-compatible",
  "enabled": false,
  "baseUrl": "https://provider.example/v1",
  "secretConfigured": true,
  "models": {
    "chat": "chat-model-id",
    "vision": "vision-model-id",
    "hearing": "asr-model-id",
    "tts": "tts-model-id",
    "voiceClone": "voice-clone-model-id",
    "video": "video-model-id",
    "image": "image-model-id"
  },
  "temperature": 0.8,
  "status": "ready",
  "lastTestAt": "2026-09-03T10:00:00+08:00"
}
```

服务端返回时不得返回原始密钥，只返回 `secretConfigured`、脱敏后缀或可轮换状态。

## 3. 建议接口分组

现有 `docs/openapi.json` 是早期单聊/记忆草案，可作为命名参考，但尚未覆盖完整原型。后端正式定义建议至少包含：

```text
/me
/companions
/companions/{id}
/conversations
/conversations/{id}/messages
/conversations/{id}/read
/groups
/groups/{id}
/groups/{id}/members
/groups/{id}/rounds
/memories
/memories/{id}
/memories/search
/moments
/moments/{id}/likes
/moments/{id}/comments
/moments/{id}/save
/api-profiles
/api-profiles/{id}/test
/api-profiles/{id}/models
/usage/summary
/usage/records
/data/export
/data/import
/data
```

## 4. 关键后端流程

### 单聊

1. 校验用户与 AI 好友关系。
2. 读取 AI 人设、能力和绑定的 API 配置。
3. 召回长期记忆与最近会话，压缩到上下文预算内。
4. 调用对应协议适配器，流式返回文本或多模态事件。
5. 保存消息、用量、错误、供应商请求 ID 和安全结果。
6. 异步进行摘要、记忆候选提取；需要时等待用户确认后入库。

### 多 AI 群聊

1. 读取群成员和调度策略。
2. 为一轮生成唯一 `roundId`，选择发言成员并限制最大数量。
3. 每位 AI 只读取其允许访问的角色记忆和群聊上下文。
4. 逐条产生带 `senderId` 的消息，遵守冷却时间和取消信号。
5. 防止 AI 间无限自激循环，达到上限立即结束本轮。

### 主动消息与主动朋友圈

1. 后端任务系统读取时间窗、频率、每日上限和免打扰状态。
2. 生成前先检查用户关闭状态、最近互动时间和内容安全。
3. 使用幂等键避免重复推送或重复发动态。
4. 所有主动行为可查看、可关闭、可删除，并保留审计记录。

## 5. 多模态与授权

- `hearing`：音频上传、ASR 转写和权限提示。
- `tts`：文本转语音，返回可播放资源和时长。
- `voiceClone`：单独授权、样本用途说明、撤销和审计，不与普通 TTS 混为一个开关。
- `vision`：图片上传、理解结果和原图保留策略。
- `video`：视频上传、转码、理解任务和异步状态。
- `imageGeneration`：文生图任务、内容审核、计费和结果资产。
- `webSearch`：搜索来源、引用、时效和隐私边界。

前端开关只表达“用户希望允许该能力”，后端仍需验证模型、供应商权限、额度和协议是否真实可执行。

## 6. 用量与错误

每次模型调用建议记录：

```json
{
  "id": "usage-record-id",
  "userId": "user-id",
  "conversationId": "conversation-id",
  "companionId": "companion-id",
  "apiProfileId": "api-profile-id",
  "provider": "provider-name",
  "model": "model-id",
  "capability": "chat",
  "inputTokens": 1200,
  "outputTokens": 260,
  "cachedTokens": 0,
  "estimatedCost": 0.0021,
  "currency": "USD",
  "latencyMs": 1340,
  "status": "success",
  "providerRequestId": "request-id",
  "createdAt": "2026-09-03T10:00:00+08:00"
}
```

错误响应应统一包含 `code`、`message`、`traceId`、`retryable` 和可展示给用户的简短说明，避免把供应商密钥或完整原始响应传到前端。

## 7. 替换 localStorage 的顺序

1. 用户资料、AI 好友、群聊和基础设置。
2. 会话列表、历史消息和已读状态。
3. API 配置的服务端安全保存和模型目录。
4. 长期记忆、向量检索和记忆确认流程。
5. 多 AI 群聊调度、主动消息和主动朋友圈任务。
6. 多模态能力、真实用量和审计后台。

每完成一层后，保留当前前端模拟作为开发环境 fallback，但界面必须明确显示“模拟”或“真实服务”状态，不能混淆。
