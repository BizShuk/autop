# 術語表 (Terminology)

## Autop LLM CLI Façade

| 術語 (Term) | 英文 (English) | 定義 (Definition) | 出處 (Source) |
| ----------- | -------------- | ----------------- | ------------- |
| `Autop` | Autop | App name 與獨立 Go CLI；依設定選擇並啟動本機 LLM CLI，不直接呼叫 provider API | `../main.go`, `../cmd/autop/settings.go` |
| `Client` | Client | `-c` 所選的具名 CLI profile，包含 driver、command、model、effort、auto approval 與 credential contract | `../cmd/autop/driver/driver.go:ClientConfig` |
| `Default client` | Default client | 未提供 `-c` 時使用的 client ID | `../cmd/autop/settings.go:Settings.DefaultClient`, `resolveClient()` |
| `Driver` | CLI driver | 將 client profile 與 final prompt 轉換成 executable、arguments、stdin 與 child environment 的獨立 process adapter package | `../cmd/autop/driver/driver.go:Prepare()` |
| `Prompt template` | Prompt template | `-t` 所選的 inline 或 file-backed Go `text/template`；未提供 `-t` 時不參與 prompt | `../cmd/autop/template.go:renderPrompt()` |
| `Prepared process` | Prepared process | 已完成 CLI-specific mapping、可交給 `exec.CommandContext` 執行的 command 資料 | `../cmd/autop/driver/driver.go:Process` |
| `Config command` | Configuration command | 由 `gosdk/cmd.ConfigCmd` 提供的 `autop config` 子命令；顯示合併設定、以 `default` 寫入 embedded `settings.example.json`，或把修改寫入 SDK 管理的 `settings.local.json` | `../cmd/autop/command.go:init()`, `../cmd/autop/defaults.go:embeddedSettings` |
| `Wizard` | Configuration wizard | `autop wizard`（alias `autop w`）依序選擇 CLI、template、permission bypass、model、effort、task prompt 與 optional cron schedule；task prompt 以明文 positional argument 保存；cron 支援 `N`、`r` 或完整五欄格式 | `../cmd/autop/wizard.go:WizardCmd`, `runWizardCommand()` |
| `Managed PM2 task` | Managed PM2 task | 由 wizard 以 begin/end marker 管理且設為 `optional: true` 的 app block；在含有 `cmd/autop/` 的 workspace 寫入 `ecosystem.config.js`，名稱為 `AutoP <client> <template>`，`cwd` 為 workspace 絕對路徑 | `../cmd/autop/install.go:installEcosystemTask()`, `../cmd/autop/wizard.go:resolveWizardTarget()` |

## 縮寫 (Abbreviations)

| 縮寫 | 全稱 | 說明 |
| ---- | ---- | ---- |
| `LLM` | Large Language Model | `autop` 所啟動 CLI 的模型執行目標；出處：`../cmd/autop/command.go` |
| `PM2` | Process Manager 2 | 讀取 `ecosystem.config.js` 並執行 `autop` task 的外部 process manager；出處：`../cmd/autop/install.go` |
