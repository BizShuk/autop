# Autop — 技術脈絡 (Technical Context)

## 專案結構 (Project Structure)

```tree
main.go                       # binary entry point、signal 與 exit code
cmd/autop/
├── command.go                # RootCmd、root flags、runtime wiring 與 command execution
├── settings.go               # gosdk config.Default、defaults 與 validation
├── defaults.go               # embed settings.example.json 並註冊 SDK default seed
├── settings.example.json     # 完整 client、credential 與 template 範例
├── template.go               # Go text/template renderer
├── driver/
│   ├── driver.go             # ClientConfig、CredentialConfig、Process、Prepare
│   ├── agy.go                # agy argv 與 prompt injection
│   ├── claude.go             # Claude family argv、settings 與 credential env
│   ├── codex.go              # Codex argv、effort 與 stdin prompt
│   └── driver_test.go        # driver mapping tests
├── runner.go                 # credential preflight 與 exec.CommandContext
├── wizard.go                 # interactive PM2 task wizard
├── install.go                # managed ecosystem.config.js atomic update
└── *_test.go                 # command、settings、template、runner、wizard tests
docs/terminology.md           # autop terminology single source
plans/                        # autop design plans
```

## 技術棧 (Tech Stack)

- Language: `Go 1.26.3`
- CLI: `spf13/cobra`
- Configuration: `github.com/bizshuk/gosdk/config` + `spf13/viper`
- Shared config command: `github.com/bizshuk/gosdk/cmd.ConfigCmd`
- Logging: standard `log/slog`，由 `github.com/bizshuk/gosdk/log` 初始化
- Process execution: `os/exec.CommandContext`，不使用 shell

## 關鍵決策 (Key Decisions)

- `autop` 只啟動已註冊的本機 `agy`、Claude family 或 `codex` CLI，不直接呼叫 provider
  API。
- `config.Default(WithAppName("autop"))` 是設定載入唯一入口；runtime config 位於
  `~/.config/autop/`，不依賴 workspace `settings.json`。
- `gosdk/cmd.ConfigCmd` 直接掛在 root Cobra command。`autop config` 顯示合併設定，
  `default`、`--update`、`--add`、`--delete`、`--append` 與 `--remove-from` 由 SDK
  管理寫入檔案。`settings.example.json` 以 `go:embed` 註冊為 `settings.json` 的
  default seed；`autop config default` 才會顯式建立 user-level 設定。
- Cobra command 採 gosdk 慣例：`RootCmd` 與 `WizardCmd` 是 exported package-level
  vars，flags 與 subcommands 由 `init()` 註冊；測試每次執行 singleton command 前重設
  bound values 與 `pflag.Flag.Changed`。
- Profile 的 `auto_approve` 只提供 wizard 預設值。直接執行必須明確提供
  `--bypass-permission=true` 才加入 dangerous permission flag。
- Prompt template 使用 Go `text/template`。Codex skill 使用 `$` prefix；agy 與 Claude
  skill 使用 `/` prefix。
- Child 啟動前用 `log/slog` 記錄 shell-safe command；credential environment 不進 log。
- PM2 task 由 wizard 以 marker 管理、設為 `optional: true`、`autorestart: false`，並以
  atomic rename 更新 `ecosystem.config.js`。
- 在含有 `cmd/autop/` 的 workspace，wizard 將 PM2 設定寫入本目錄，但 task `cwd` 保持
  workspace root；一般 workspace fallback 至根目錄 `ecosystem.config.js`。

## Client contract

`driver.ClientConfig` 至少包含：`driver`、`command`、`model`、`models`、`effort`、
`efforts`、`auto_approve`、`prompt_transport` 與 `credential`。Credential 只描述
OAuth 或 source/target environment mapping，不保存 secret literal。

Provider permission mapping：

| Driver | Bypass flag |
| --- | --- |
| `agy` | `--dangerously-skip-permissions` |
| `claude` | `--dangerously-skip-permissions` |
| `codex` | `--dangerously-bypass-approvals-and-sandbox` |

## 開發與驗證 (Development & Verification)

從 repository root 執行：

```bash
go test ./cmd/autop/... -count=1
go vet ./...
go build .
```

完整 repository 驗證：

```bash
go test ./... -count=1
go vet ./...
go build ./...
go mod tidy -diff
```

命令與術語細節見 [`README.md`](README.md) 與
[`docs/terminology.md`](docs/terminology.md)。
