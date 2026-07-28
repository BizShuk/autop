# 術語表 (Terminology)

## Autop LLM CLI Façade

| 術語 (Term) | 英文 (English) | 定義 (Definition) | 出處 (Source) |
| ----------- | -------------- | ----------------- | ------------- |
| `Autop` | Autop | App name 與獨立 Go CLI；依設定選擇並啟動本機 LLM CLI，不直接呼叫 provider API | `../main.go`, `../cmd/settings.go` |
| `Client` | Client | `-c` 所選的具名 CLI profile，包含 driver、command、model、effort、auto approval 與 credential contract | `../cmd/driver/driver.go:ClientConfig` |
| `Default client` | Default client | 未提供 `-c` 時使用的 client ID | `../cmd/settings.go:Settings.DefaultClient`, `resolveClient()` |
| `Driver` | CLI driver | 將 agy、Claude family、Codex 或 Grok client profile 與 final prompt 轉換成 executable、arguments、stdin 與 child environment 的獨立 process adapter package | `../cmd/driver/driver.go:Prepare()` |
| `Prompt template` | Prompt template | `-t` 所選的 inline 或 file-backed Go `text/template`；未提供 `-t` 時不參與 prompt | `../cmd/template.go:renderPrompt()` |
| `Prepared process` | Prepared process | 已完成 CLI-specific mapping、可交給 `exec.CommandContext` 執行的 command 資料 | `../cmd/driver/driver.go:Process` |
| `Config command` | Configuration command | 由 `gosdk/cmd.ConfigCmd` 提供的 `autop config` 子命令；顯示合併設定、以 `default` 寫入 embedded `settings.example.json`，或把修改寫入 SDK 管理的 `settings.local.json` | `../cmd/command.go:init()`, `../cmd/defaults.go:embeddedSettings` |
| `Wizard` | Configuration wizard | `autop wizard`（alias `autop w`）依序選擇 CLI、template、permission bypass、model、effort、task prompt 與 optional cron schedule；寫入後依序顯示 shell-safe original Autop command、mapped execute command、ecosystem path 與實際 configuration；task prompt 以 shell-safe quoted positional argument 保存；cron 支援 `N`、`r` 或完整五欄格式 | `../cmd/wizard.go:WizardCmd`, `runWizardCommand()` |
| `Managed PM2 task` | Managed PM2 task | 由 wizard 以 begin/end marker 管理且設為 `optional: true` 的 app block；統一寫入 workspace root 的 `ecosystem.config.js`，名稱為 `Autop <client> <project-folder>`，其中 `<project-folder>` 是 `cwd` 的 project root folder name；`cwd` 為 workspace 絕對路徑 | `../cmd/install.go:installEcosystemTask()`, `../cmd/wizard.go:resolveWizardTarget()` |

## 縮寫 (Abbreviations)

| 縮寫 | 全稱 | 說明 |
| ---- | ---- | ---- |
| `LLM` | Large Language Model | `autop` 所啟動 CLI 的模型執行目標；出處：`../cmd/command.go` |
| `PM2` | Process Manager 2 | 讀取 `ecosystem.config.js` 並執行 `autop` task 的外部 process manager；出處：`../cmd/install.go` |
