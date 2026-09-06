---
archive_version: 1
created_at: "2026-09-05T14:05:04+08:00"
member: "pengzexi79-ops"
role: "director"
task_id: "QB-COLLAB-002"
branch: "docs/team-collaboration-skill"
base_commit: "b8b2355"
change_fingerprint: "903ad591a011bb29c07dd54933484802ea3b516daa1916b2238bc49343e0918e"
---

# 亲伴开发保存点：QB-COLLAB-002

## 本次目的

把协同 Skill 修正为成员修改完成后，由 AI 根据实际 Git 变更自动生成开发保存点档案。

## Git 检测到的变更

- ` M` `.agents/skills/qinban-team-collaboration/agents/openai.yaml`
- ` M` `.agents/skills/qinban-team-collaboration/references/ai-task-template.md`
- `??` `.agents/skills/qinban-team-collaboration/references/archive-format.md`
- ` M` `.agents/skills/qinban-team-collaboration/references/workflow.md`
- `??` `.agents/skills/qinban-team-collaboration/scripts/create-change-archive.ps1`
- ` M` `.agents/skills/qinban-team-collaboration/SKILL.md`
- ` M` `.github/pull_request_template.md`
- ` M` `AGENTS.md`
- `??` `docs/development-archive/README.md`

## 具体变化

将 Skill 主流程改为 archive-first：成员首次声明角色、成员和任务编号后，同一 AI 任务内每轮有意义修改自动存档；成员手工修改后调用 Skill 时只读取现有 diff 补档。新增 PowerShell 存档脚本、按日期独立文件、内容指纹去重、档案格式说明、最短调用模板，并在 AGENTS 与 PR 模板中增加存档门禁。

## 接口、数据库与兼容影响

不修改前端业务、后端 API 或数据库；只调整团队协作 Skill、AI 规则、PR 模板和开发存档基础设施。

跨角色技术批准：否。

## 验证

PowerShell Parser 语法检查通过；脚本 DryRun 通过；Skill frontmatter、名称、引用文件、默认提示与 short_description 长度手工校验通过；暂存区和未暂存区 git diff --check 通过。官方 quick_validate.py 因当前 Python 环境缺少 PyYAML 未执行成功，未安装额外全局依赖。

git diff --check：通过（已检查暂存区与未暂存区；未跟踪文件不在 git diff --check 范围内）。

## 风险、未完成与交接

交给 backend-lead 审核角色边界、存档字段、指纹去重和合并规则；当前协作 PR 未经技术审核不合并。
