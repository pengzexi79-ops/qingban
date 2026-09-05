# 亲伴仓库 AI 开发规则

在本仓库规划、修改、评审或提交代码前，必须阅读并遵守：

- `.agents/skills/qinban-team-collaboration/SKILL.md`
- 与当前角色对应的 `references/` 文档

项目采用“双负责人”模型：

- 总指导决定产品方向、优先级和验收标准；
- `backend-lead` 决定技术架构、API、数据库、依赖、安全、测试和所有代码是否可合并；
- 三名前端以 AI 氛围编程为主，只在技术负责人批准的边界内开发；
- 总指导亲自写前端时也不能绕过 `backend-lead` 技术审核。

不可跳过的事实：

- 产品名称是“亲伴 / QinBan”，仓库名 `qingban` 与旧兼容键可保留；
- 主干分支是 `master`；
- 当前四个前端主文件是共享热点，同一文件同一时段只允许一个任务分支修改；
- 不直接 push 主干，不 force push，不 rebase 团队分支，不清理或覆盖未知工作；
- 未经成员当前任务明确要求，AI 不 commit、不 push、不创建 PR、不修改 GitHub 设置。

产品问题听总指导；技术问题听 `backend-lead`。若二者冲突，先列出可行方案与代价，由总指导选产品取舍，再由 `backend-lead` 定实现。
