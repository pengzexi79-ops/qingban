# 前端 Vue 迁移记录

- 更新时间：2026-09-05 21:34（UTC+08:00）
- 协作分支：`dev-lwtpxr-frontend-design`
- 迁移范围：`frontend/` 前端入口、构建配置、Vue 入口组件和本地数据层

## 本次更新

- 将前端运行入口迁移为 Vue 3 + Vite。
- 新增 `frontend/src/main.js`、`frontend/src/App.vue`、`frontend/src/app.js`、`frontend/src/store.js` 和 `frontend/src/style.css`。
- 新增 `frontend/package.json`、`frontend/pnpm-lock.yaml` 和 `frontend/vite.config.js`。
- 将 `frontend/index.html` 改为 Vite HTML 入口，不再加载 Vue CDN。
- 将本地数据层从全局 `window.QinbanStore` 改为 ES Module 导出。
- 更新根目录 README，说明新的安装、启动和构建方式。
- 保留旧的 `frontend/js`、`frontend/css`、`frontend/vendor` 文件作为历史兼容参考；新的运行入口不引用它们。

## 验证

- `vite build` 通过，生产包正常生成。
- `git diff --check` 通过。
- 已从最新 `origin/master` fast-forward 到当前分支，未发生冲突。
- 本次没有修改 `master`，没有使用 rebase 或 force push。

## 协作说明

本记录对应当前前端设计分支，等待审核人通过 PR 后按团队规范 Squash and merge。审核通过前不直接合并到 `master`。
