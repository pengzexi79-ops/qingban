# 亲伴仓库 AI 开发规则

在本仓库规划、修改、评审或提交代码前，必须阅读：

- `.agents/skills/qinban-team-collaboration/SKILL.md`
- 与当前角色对应的 `references/` 文档

最重要的执行规则：**AI完成一轮有意义的代码或文档修改后，必须在汇报、commit、push 或 PR 前，根据实际 Git diff 生成新的开发保存点档案。** 使用 Skill 内的 `scripts/create-change-archive.ps1`，档案保存在 `docs/development-archive/YYYY-MM-DD/`，不得覆盖其他成员或旧记录。

成员在当前 AI 任务首次声明角色、GitHub 用户名和任务编号后，AI 后续自动沿用，不得让成员每轮重复提醒存档。成员先手工改代码再调用 Skill 时，只根据现有实际改动补档，不擅自追加实现。

项目采用“双负责人”模型：

- 总指导决定产品方向、优先级和验收；
- `backend-lead` 决定技术架构、API、数据库、依赖、安全、测试和代码是否可合并；
- 三名前端以 AI 氛围编程为主，只在技术负责人批准的边界内开发；
- 总指导亲自写前端时也声明前端执行角色，不绕过技术审核。

固定事实：

- 产品名称是“亲伴 / QinBan”，仓库名 `qingban` 与旧兼容键可保留；
- 主干是 `master`；不直接 push 主干，不 force push，不 rebase 团队分支；
- 当前四个前端主文件是共享热点，同一文件同一时段只允许一个任务分支修改；
- 不清理或覆盖未知工作；未经成员明确要求，AI 不 commit、不 push、不创建 PR、不修改 GitHub 设置。

产品问题听总指导；技术问题听 `backend-lead`。
