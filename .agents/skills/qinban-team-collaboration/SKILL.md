---
name: qinban-team-collaboration
description: Archive each AI-assisted change checkpoint in the QinBan/亲伴 repository after a member edits or saves work. Use after code or document changes and before reporting, committing, pushing, or preparing a PR to inspect the actual Git diff, enforce the member role boundary, and create an append-only development record without overwriting another member's archive.
---

# 亲伴 AI 变更存档与协同

本 Skill 的首要用途不是在开工前讲一遍规范，而是：**成员完成一轮修改后，AI 自动读取实际 Git 变更，为这次保存点生成一份可追踪的开发档案。**

## 成员怎么使用

成员在一个任务中首次声明 `角色、GitHub 用户名、任务编号` 并调用 `$qinban-team-collaboration`。同一 AI 任务内，AI 后续每完成一轮有意义的修改都沿用这组信息自动存档，不要求成员反复下达“存档”命令。

如果成员先在编辑器中自行修改，再调用本 Skill，AI应立即读取当前 Git diff、总结真实变化并生成档案，不重新规划或擅自改代码。角色、成员或任务编号能从当前任务卡和分支明确判断时直接复用；确实无法判断时只询问缺失项，不能伪造作者。

## 何时必须执行

只要 AI 在本仓库完成了一轮有意义的代码或文档修改，就应在以下动作之前自动执行本 Skill，无需成员再次提醒：

- 向成员汇报“已完成”；
- commit 或 push；
- 创建或更新 PR；
- 切换任务、角色或分支；
- 成员明确说“保存、存档、记录进度”。

一个“保存点”指一轮可以独立描述的修改批次，不是底层工具每写入一行就生成一份档案。成员在 AI 之外手动按编辑器保存键时，Skill 无法自行触发；成员再次调用 AI 或准备提交时，AI必须补做存档。

## 自动存档流程

1. 读取 `git status --short --branch`、`git diff`、暂存区和未跟踪文件，确认本次真实改动。
2. 从当前任务卡、会话和分支复用成员的执行角色、GitHub 用户名和任务编号；只有确实无法判断时才询问缺失项，不能伪造作者。
3. 按 [角色与文件归属](references/roles-and-ownership.md) 检查是否越界。前端改到后端、契约或数据库时必须停止并报告，除非 `backend-lead` 已明确批准。
4. AI 根据 diff 写出：修改目的、功能变化、涉及文件、接口/数据库影响、验证结果、风险和下一位负责人。不得只写“优化代码”“更新页面”等空话。
5. 运行 `scripts/create-change-archive.ps1`，在 `docs/development-archive/YYYY-MM-DD/` 生成一份新的 Markdown 档案。
6. 再次检查档案与代码是否一致。档案和代码进入同一分支、同一 PR；不要单独在主干补日志。
7. 在最终回复中给出档案相对路径。没有实际 diff 时不生成空档案，只说明“本次无可存档变更”。

具体格式见 [存档格式与规则](references/archive-format.md)。

## 存档不可覆盖原则

- 每个保存点生成一个新文件，禁止所有成员共同追加同一个日志文件。
- 文件名包含时间、角色、成员和任务编号，天然避免三名前端与后端互相覆盖。
- 已合并的历史档案只读；发现错误时新增“更正档案”，不重写旧记录。
- 存档只记录本次 diff，不把历史工作、计划功能或未完成占位描述成已经完成。
- 不复制 API Key、Token、Cookie、`.env` 内容、真实用户数据或大段源码到档案。

## 团队决策权

- **总指导决定产品**：方向、优先级、体验目标、验收标准和不做项。
- **`backend-lead` 决定技术**：架构、API、数据库、依赖、安全、测试、冲突处理与是否可合并。
- 三名前端以 AI 氛围编程为主，在技术负责人划定的文件和接口内开发。
- 总指导亲自写前端时也选择一个 `frontend-*` 执行角色，其代码和存档仍交给 `backend-lead` 技术审核。

## 并行与文件边界

后端和纯前端页面可以并行，三个前端也可以并行，但必须满足：分支不同、任务不同、修改路径不重叠、API 已冻结、热点文件没有多人同时占用。

当前热点文件是：

- `frontend/index.html`
- `frontend/js/app.js`
- `frontend/js/store.js`
- `frontend/css/style.css`

这些文件由 `backend-lead` 分配，同一时段一个文件只有一个写入者。每名前端同时最多保留一个编码任务。长期并行前，应先按技术负责人批准的方案完成纯模块拆分。

## Git 与 AI 安全边界

- 主干是 `master`；不得直接在主干开发或 push。
- 不 force push，不 rebase 团队分支，不执行 `git reset --hard`、`git clean -fd`，不动别人的分支和未提交内容。
- 未经 `backend-lead` 批准，前端不得改框架、依赖、构建链、公共状态、API 契约、数据库或 `backend/**`。
- AI 只有在成员明确要求时才 commit、push 或创建 PR；但**生成本地存档属于完成修改的一部分，不需要再次询问**。
- 所有代码 PR 由 `backend-lead` 做最终技术判断；改变产品方向的 PR 同时由总指导验收。

分支和 PR 规则见 [分支、PR 与验收流程](references/workflow.md)。成员可复制 [AI 任务模板](references/ai-task-template.md)。
