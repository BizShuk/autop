# Autop LLM CLI Façade

`autop` 是獨立的 Go CLI，提供單一入口來啟動已設定的本機 LLM CLI。它依
`-c` 選擇 `agy`、Claude family 或 `codex` client，依 `-t` 套用 prompt template，
不直接呼叫 provider API。

## 使用方式

```bash
# 從 cc-plugin workspace root 安裝 binary
go install ./cmd/autop

# 使用 default client，不套 template
printf '%s' 'summarize current workspace' | autop

# 使用 client 與 template
autop -c codex -t system

# 互動建立 PM2 task
autop wizard
autop w

# 檢視合併後設定
autop config
autop config --source

# 將隨 binary 內嵌的範例設定寫入 ~/.config/autop/settings.json
autop config default

# 升級時只補上新欄位，不覆蓋既有值
autop config default --merge

# 更新 user-level settings.local.json
autop config --update clients.codex.model=gpt-5.5
```

未提供 `-c` 時使用 `default_client`；未提供 `-t` 時完全跳過 template。Prompt
可由 positional arguments 或 piped stdin 提供，兩者不可同時使用。

## 執行流程

1. `config.Default(WithAppName("autop"))` 載入
   `~/.config/autop/settings.json` 與 `settings.local.json`。
2. Driver 將 client profile、model、effort、permission bypass、workspace 與 prompt
   映射成 `agy`、`claude` 或 `codex exec` 的 arguments。
3. 使用 Go `text/template` 渲染具名 template；Codex skill 使用 `$` prefix，agy 與
   Claude skill 使用 `/` prefix。
4. Child 啟動前以 `log/slog` 輸出 shell-safe command，再串流 stdout／stderr 並保留
   child exit code。

Credential 不由 `autop` 保存或輸出：OAuth 使用各 CLI 原有 login state；API-key
client 從設定的 environment variable 繼承 secret。

`cmd/autop/settings.example.json` 會在建置時嵌入 binary，`autop config default` 才會寫入
user-level config；既有 `settings.json` 預設保留，需明確使用 `--merge` 或 `--force`。

## Client 與 template

內建 client profile：

- `agy`
- `codex`
- `claude`
- `claudem`
- `claudew`
- `claudep`
- `claudet`（預設停用，待 credential contract 完成）

內建 template：`system`、`auto-evolving` 與 `codex-base`。完整 client、model、effort、
credential 與 template 範例見
[`cmd/autop/settings.example.json`](cmd/autop/settings.example.json)。

設定中的 `auto_approve` 只作為 wizard 的 bypass 預設值；直接執行時必須明確提供
`--bypass-permission=true`，才會映射 provider dangerous flag：

| Client driver | Permission bypass flag |
| --- | --- |
| `agy` | `--dangerously-skip-permissions` |
| Claude family | `--dangerously-skip-permissions` |
| `codex` | `--dangerously-bypass-approvals-and-sandbox` |

## PM2 wizard

`autop wizard` 依序詢問 CLI、template、permission bypass、model、effort、task prompt
與 optional cron schedule。workspace 含有 `cmd/autop/` 時，task 寫入
`cmd/autop/ecosystem.config.js`；PM2 `cwd` 仍保存 workspace root 的絕對路徑。

Managed task 具備：

- 名稱：`AutoP <client> <template>`
- `namespace: "autop"`
- `optional: true`
- `autorestart: false`
- `// autop:begin <task>`／`// autop:end <task>` marker

```bash
pm2 start cmd/autop/ecosystem.config.js
```

Root [`ecosystem.config.js`](../../ecosystem.config.js) 只聚合非 autop 的常駐程序，並
載入本目錄的 PM2 task 定義。

## 開發

```bash
go test ./cmd/autop/... -count=1
go vet ./...
go build ./cmd/autop
```

技術脈絡見 [`CLAUDE.md`](CLAUDE.md)，術語單一定義見
[`docs/terminology.md`](docs/terminology.md)。
