# Autop LLM CLI Façade

`autop` 是獨立的 Go CLI，提供單一入口來啟動已設定的本機 LLM CLI。它依
`-c` 選擇 `agy`、Claude family、`codex` 或 `grok` client，依 `-t` 套用 prompt
template，不直接呼叫 provider API。

## 使用方式

```bash
# 從 repository root 安裝 binary
go install .

# 使用 default client，不套 template
printf '%s' 'summarize current workspace' | autop

# 使用 client 與 template
autop -c codex -t system

# 只印出實際要執行的 command line，不啟動 CLI
autop --dry-run -c claudem -t system -- review workspace

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
   映射成本機 `agy`、`claude`、`codex exec` 或 `grok` 的 arguments；agy、Claude
   family 帶入 `--add-dir <cwd>`，Codex 帶入 `-C <cwd>`，Grok 帶入 `--cwd <cwd>`。
3. 使用 Go `text/template` 渲染具名 template；Codex skill 使用 `$` prefix，agy 與
   Claude skill 使用 `/` prefix。
4. Child 啟動前以 `log/slog` 輸出 shell-safe command，再串流 stdout／stderr 並保留
   child exit code。

Credential 不由 `autop` 保存或輸出：OAuth 使用各 CLI 原有 login state；API-key
client 從設定的 environment variable 繼承 secret。

`cmd/settings.example.json` 會在建置時嵌入 binary，`autop config default` 才會寫入
user-level config；既有 `settings.json` 預設保留，需明確使用 `--merge` 或 `--force`。

## Client 與 template

內建 client profile：

- `agy`
- `codex`
- `grok`
- `claude`
- `claudem`
- `claudew`
- `claudep`
- `claudet`（預設停用，待 credential contract 完成）

所有 Claude profile 都直接啟動 `claude` executable，profile 差異由 settings 檔與
credential environment 表示，不再依賴 `~/bin` 下的 wrapper script：

| Profile | Executable | Settings | Credential |
| ------- | ---------- | -------- | ---------- |
| `claude` | `claude` | `~/projects/cc-plugin/config/settings.json` | OAuth |
| `claudem` | `claude` | `~/projects/cc-plugin/config/minimax.json` | `MINIMAX_API_KEY` → `ANTHROPIC_AUTH_TOKEN` |
| `claudew` | `claude` | `~/projects/cc-plugin/config/llmbox.json` | `TIKTOK_API_KEY` → `ANTHROPIC_AUTH_TOKEN` |
| `claudep` | `claude` | `~/projects/cc-plugin/config/proxy.json` | `AGENTSDK_PROXY_API_KEY` → `ANTHROPIC_AUTH_TOKEN` |

例如 `autop -c claudem --bypass-permission=true -- review workspace` 會執行等同於
`ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY" claude --dangerously-skip-permissions
--settings ~/projects/cc-plugin/config/minimax.json ...` 的完整命令。每個 client
與 flag 組合對應的實際 command line 見
[`docs/command-matrix.md`](docs/command-matrix.md)。

啟動 CLI 前 autop 會先做 preflight：profile 若指定 `settings` 檔而該檔不存在（或指到
目錄），命令直接失敗並列出完整路徑，不會啟動 child process。不需要 settings 檔的
profile（`agy`、`codex`、`grok`）則跳過這道檢查。

`--dry-run` 只把解析後的 command line 印到 stdout，不啟動 child process。輸出與
執行前 log 的內容一致：shell-safe quoting、stdin prompt 以 `printf '%s' ... |` 呈現，
credential 以 `TARGET="$SOURCE"` 形式顯示而不解析 secret。因為不會執行任何東西，
dry run 一併跳過 settings 檔與 credential 兩道 preflight，所以尚未設定 API key 或
settings 檔時仍可預覽命令。

內建 template：`system` 與 `auto-evolving`，兩者都以 inline `content` 定義；
template 另可用 `file` 指向外部檔案，預設不含這類 template。完整 client、model、effort、
credential 與 template 範例見
[`cmd/settings.example.json`](cmd/settings.example.json)。

設定中的 `auto_approve` 只作為 wizard 的 bypass 預設值；直接執行時必須明確提供
`--bypass-permission=true`，才會映射 provider dangerous flag：

| Client driver | Permission bypass flag |
| --- | --- |
| `agy` | `--dangerously-skip-permissions` |
| Claude family | `--dangerously-skip-permissions` |
| `codex` | `--dangerously-bypass-approvals-and-sandbox` |
| `grok` | `--always-approve --permission-mode auto` |

## PM2 wizard

`autop wizard` 依序詢問 CLI、template、permission bypass、model、effort、task prompt
與 optional cron schedule，寫入 PM2 task 後依序顯示 shell-safe `Original command
(autop)`、driver mapping 後的 `Execute command (<cli>)`、`ecosystem.config.js` 絕對
路徑及實際寫入的完整 configuration。兩個 command label 以 ANSI color 區分，但 command
value 保持無色；path 與 configuration label 同樣以 ANSI color 標示，而各自的 value
保持無色。`$find-activity` 等 Codex skill prompt 會保持 literal，不會在 preview 中被
shell 展開。Wizard 以 workspace 的 `cmd/` 目錄辨識
project root，task 統一寫入 project root 的 `ecosystem.config.js`；PM2 `cwd` 同樣保存
workspace root 的絕對路徑。由於本機 PM2 會透過 Bash 執行 `script` 與 `args`，wizard
會在 `ecosystem.config.js` 中將完整 prompt 包成 shell-safe 單引號 argument；prompt
內原有的單引號會一併 escape。

Managed task 具備：

- 名稱：`Autop <client> <project-folder>`，其中 `<project-folder>` 是 PM2 `cwd` 的
  project root folder name
- `namespace: "autop"`
- `optional: true`
- `autorestart: false`
- `// autop:begin <task>`／`// autop:end <task>` marker

```bash
pm2 start ecosystem.config.js
```

Wizard 會保留 root `ecosystem.config.js` 中既有 app，並以 managed marker 新增或更新
Autop task。

## 開發

```bash
go test ./cmd/... -count=1
go vet ./...
go build .
```

技術脈絡見 [`CLAUDE.md`](CLAUDE.md)，術語單一定義見
[`docs/terminology.md`](docs/terminology.md)。
