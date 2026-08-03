# Autop — 技術脈絡 (Technical Context)

## 專案結構 (Project Structure)

```tree
main.go                       # binary entry point、signal 與 exit code
cmd/
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
│   ├── grok.go               # Grok argv、cwd 與 argument prompt
│   └── driver_test.go        # driver mapping tests
├── runner.go                 # credential preflight 與 exec.CommandContext
├── wizard.go                 # interactive PM2 task wizard
├── install.go                # managed ecosystem.config.js atomic update
├── matrix_test.go            # client × flag → command line golden matrix
└── *_test.go                 # command、settings、template、runner、wizard tests
scripts/
├── generate-wizard-commands.sh    # 對每個 client 跑一次非互動 wizard
└── generate-dry-run-matrix.sh     # -c × -t × --model × --bypass-permission dry run
docs/command-matrix.md        # generated client × flag → command line matrix
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

- `autop` 只啟動已註冊的本機 `agy`、Claude family、`codex` 或 `grok` CLI，不直接呼叫
  provider API。
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
- 所有 Claude profile 的 `command` 一律是 `claude`。Profile 差異由 `settings` 檔與
  `credential` environment mapping 表示，不依賴 `~/bin` 下的 wrapper script
  （`claudem`／`claudew`），因此 autop 只需 `claude` 在 PATH 上即可執行全部 profile。
- `Process.EnvPreview` 是 command 顯示與 log 用的 environment prefix，形式為
  `TARGET="$SOURCE"`；`Process.ExtraEnv` 才帶已解析的 secret，且永不進 log。
- `--dry-run` 走 `driver.BuildProcess` 而非 `driver.Prepare`，把 `formatCommand` 結果
  印到 command stdout 後直接返回：不執行 child、不解析 credential，因此也跳過兩道
  preflight。命令矩陣 golden test 不含此 flag，因為它不改變 driver mapping。
- `driver.Prepare` 在啟動 child 前做兩道 preflight：settings 檔存在性與 credential
  environment。`settings` 為空（agy、codex、grok）時直接跳過檔案檢查。
  `driver.BuildProcess` 不做任何 preflight，wizard 因此能在 settings 檔尚未就緒時
  仍寫出 PM2 task 與 summary。
- agy 與 Claude family 以 `--add-dir <cwd>` 加入目前 workspace；Codex 只以
  `-C <cwd>` 設定 primary workspace。Grok 以 `--cwd <cwd>` 設定工作目錄。
- Prompt template 使用 Go `text/template`。Codex skill 使用 `$` prefix；agy 與 Claude
  skill 使用 `/` prefix。
- Child 啟動前用 `log/slog` 記錄 shell-safe command；credential environment 不進 log。
- Wizard 寫入 PM2 task 後依序顯示 shell-safe original `autop` command、driver mapping
  後的 execute command、root `ecosystem.config.js` 絕對路徑與實際 configuration；輸出
  不解析或顯示 credential environment。Original/execute command 只對 label 套 ANSI
  color，command value 保持無色；ecosystem path 與 configuration 同樣只對 label 上色。
- PM2 task 由 wizard 以 marker 管理、設為 `optional: true`、`autorestart: false`，並以
  atomic rename 更新 `ecosystem.config.js`。
- 本機 PM2 會把 `script` 與 `args` 串接後交給 Bash；wizard 因此只對任意內容的 prompt
  argument 套用 shell-safe 單引號 quoting，包含 `$` preservation 與內部單引號 escape。
- Wizard 以向上找到的 `cmd/` 目錄辨識 workspace root，並將 PM2 設定寫入該 root 的
  `ecosystem.config.js`；task `cwd` 同樣保持 workspace root。找不到 `cmd/` 時則以目前
  working directory 作為 workspace root。

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
| `grok` | `--always-approve --permission-mode auto` |

## 開發與驗證 (Development & Verification)

從 repository root 執行：

```bash
go test ./cmd/... -count=1
go vet ./...
go build .
```

`scripts/generate-dry-run-matrix.sh` 對`目前 merged config`（不是 embedded default）
展開 `-c` × `-t` × `--model` × `--bypass-permission` 全組合，逐筆印出 original autop
command 與 `--dry-run` 映射後的 command line。Client、template 與 model 清單由
`autop config` 解析而來，可用 `AUTOP_MATRIX_CLIENTS`、`AUTOP_MATRIX_TEMPLATES`、
`AUTOP_MATRIX_MODELS`、`AUTOP_MATRIX_BYPASS`、`AUTOP_MATRIX_PROMPT` 縮小範圍。失敗組合
照樣印出並計入結尾統計，script 以 exit 1 結束。與 `docs/command-matrix.md` 的差別是
後者由 golden test 對 embedded default 產生、且不含 model 全展開。

`docs/command-matrix.md` 由 golden test 產生，driver argv 有變動時重新產生：

```bash
go test ./cmd -run CommandMatrix -update-matrix -count=1
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
