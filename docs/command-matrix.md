# Autop command matrix

由 `go test ./cmd/... -run CommandMatrix -update-matrix` 產生，請勿手動編輯。

每組顯示 `autop` 原始命令與 driver mapping 後實際執行的 command line。Credential 以 `TARGET="$SOURCE"` 形式呈現，不含 secret 值。Working directory 固定為 `/workspace`。

## agy

driver `agy` / command `agy` / credential `oauth`

### 預設 model 與 effort

```sh
autop -c agy -- 'summarize current repository'
```

```sh
agy --model=gemini-3.6-flash-high --effort=high --add-dir /workspace -p 'summarize current repository'
```

### --bypass-permission=true

```sh
autop -c agy --bypass-permission=true -- 'summarize current repository'
```

```sh
agy --dangerously-skip-permissions --model=gemini-3.6-flash-high --effort=high --add-dir /workspace -p 'summarize current repository'
```

### --model 覆寫

```sh
autop -c agy --model gemini-3.6-flash-medium -- 'summarize current repository'
```

```sh
agy --model=gemini-3.6-flash-medium --effort=high --add-dir /workspace -p 'summarize current repository'
```

### --effort 覆寫

```sh
autop -c agy --effort low -- 'summarize current repository'
```

```sh
agy --model=gemini-3.6-flash-high --effort=low --add-dir /workspace -p 'summarize current repository'
```

### -t system (template 無 prompt)

```sh
autop -c agy -t system
```

```sh
agy --model=gemini-3.6-flash-high --effort=high --add-dir /workspace -p 'run /system-planner for current workspace'
```

### -t system + prompt

```sh
autop -c agy -t system -- 'focus on dependency boundaries'
```

```sh
agy --model=gemini-3.6-flash-high --effort=high --add-dir /workspace -p 'run /system-planner for current workspace focus on dependency boundaries'
```

### 全 flag 組合

```sh
autop -c agy -t auto-evolving --bypass-permission=true --model gemini-3.6-flash-medium --effort low -- 'focus on dependency boundaries'
```

```sh
agy --dangerously-skip-permissions --model=gemini-3.6-flash-medium --effort=low --add-dir /workspace -p 'run /auto-evolving for current workspace focus on dependency boundaries'
```

### prompt 由 stdin 傳入

```sh
autop -c agy
```

```sh
agy --model=gemini-3.6-flash-high --effort=high --add-dir /workspace -p 'review the staged diff'
```

## claude

driver `claude` / command `claude` / credential `oauth`

### 預設 model 與 effort

```sh
autop -c claude -- 'summarize current repository'
```

```sh
claude --settings ~/projects/cc-plugin/config/settings.json --model opus --effort max --add-dir /workspace --output-format text -p 'summarize current repository'
```

### --bypass-permission=true

```sh
autop -c claude --bypass-permission=true -- 'summarize current repository'
```

```sh
claude --dangerously-skip-permissions --settings ~/projects/cc-plugin/config/settings.json --model opus --effort max --add-dir /workspace --output-format text -p 'summarize current repository'
```

### --model 覆寫

```sh
autop -c claude --model sonnet -- 'summarize current repository'
```

```sh
claude --settings ~/projects/cc-plugin/config/settings.json --model sonnet --effort max --add-dir /workspace --output-format text -p 'summarize current repository'
```

### --effort 覆寫

```sh
autop -c claude --effort low -- 'summarize current repository'
```

```sh
claude --settings ~/projects/cc-plugin/config/settings.json --model opus --effort low --add-dir /workspace --output-format text -p 'summarize current repository'
```

### -t system (template 無 prompt)

```sh
autop -c claude -t system
```

```sh
claude --settings ~/projects/cc-plugin/config/settings.json --model opus --effort max --add-dir /workspace --output-format text -p 'run /system-planner for current workspace'
```

### -t system + prompt

```sh
autop -c claude -t system -- 'focus on dependency boundaries'
```

```sh
claude --settings ~/projects/cc-plugin/config/settings.json --model opus --effort max --add-dir /workspace --output-format text -p 'run /system-planner for current workspace focus on dependency boundaries'
```

### 全 flag 組合

```sh
autop -c claude -t auto-evolving --bypass-permission=true --model sonnet --effort low -- 'focus on dependency boundaries'
```

```sh
claude --dangerously-skip-permissions --settings ~/projects/cc-plugin/config/settings.json --model sonnet --effort low --add-dir /workspace --output-format text -p 'run /auto-evolving for current workspace focus on dependency boundaries'
```

### prompt 由 stdin 傳入

```sh
autop -c claude
```

```sh
claude --settings ~/projects/cc-plugin/config/settings.json --model opus --effort max --add-dir /workspace --output-format text -p 'review the staged diff'
```

## claudem

driver `claude` / command `claude` / credential `MINIMAX_API_KEY → ANTHROPIC_AUTH_TOKEN`

### 預設 model 與 effort

```sh
autop -c claudem -- 'summarize current repository'
```

```sh
ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY" claude --settings ~/projects/cc-plugin/config/minimax.json --model MiniMax-M3 --effort xhigh --add-dir /workspace --output-format text -p 'summarize current repository'
```

### --bypass-permission=true

```sh
autop -c claudem --bypass-permission=true -- 'summarize current repository'
```

```sh
ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY" claude --dangerously-skip-permissions --settings ~/projects/cc-plugin/config/minimax.json --model MiniMax-M3 --effort xhigh --add-dir /workspace --output-format text -p 'summarize current repository'
```

### --model 覆寫

```sh
autop -c claudem --model MiniMax-M2.7 -- 'summarize current repository'
```

```sh
ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY" claude --settings ~/projects/cc-plugin/config/minimax.json --model MiniMax-M2.7 --effort xhigh --add-dir /workspace --output-format text -p 'summarize current repository'
```

### --effort 覆寫

```sh
autop -c claudem --effort low -- 'summarize current repository'
```

```sh
ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY" claude --settings ~/projects/cc-plugin/config/minimax.json --model MiniMax-M3 --effort low --add-dir /workspace --output-format text -p 'summarize current repository'
```

### -t system (template 無 prompt)

```sh
autop -c claudem -t system
```

```sh
ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY" claude --settings ~/projects/cc-plugin/config/minimax.json --model MiniMax-M3 --effort xhigh --add-dir /workspace --output-format text -p 'run /system-planner for current workspace'
```

### -t system + prompt

```sh
autop -c claudem -t system -- 'focus on dependency boundaries'
```

```sh
ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY" claude --settings ~/projects/cc-plugin/config/minimax.json --model MiniMax-M3 --effort xhigh --add-dir /workspace --output-format text -p 'run /system-planner for current workspace focus on dependency boundaries'
```

### 全 flag 組合

```sh
autop -c claudem -t auto-evolving --bypass-permission=true --model MiniMax-M2.7 --effort low -- 'focus on dependency boundaries'
```

```sh
ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY" claude --dangerously-skip-permissions --settings ~/projects/cc-plugin/config/minimax.json --model MiniMax-M2.7 --effort low --add-dir /workspace --output-format text -p 'run /auto-evolving for current workspace focus on dependency boundaries'
```

### prompt 由 stdin 傳入

```sh
autop -c claudem
```

```sh
ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY" claude --settings ~/projects/cc-plugin/config/minimax.json --model MiniMax-M3 --effort xhigh --add-dir /workspace --output-format text -p 'review the staged diff'
```

## claudep

driver `claude` / command `claude` / credential `AGENTSDK_PROXY_API_KEY → ANTHROPIC_AUTH_TOKEN`

### 預設 model 與 effort

```sh
autop -c claudep -- 'summarize current repository'
```

```sh
ANTHROPIC_AUTH_TOKEN="$AGENTSDK_PROXY_API_KEY" claude --settings ~/projects/cc-plugin/config/proxy.json --model gemini-3.5-flash --effort high --add-dir /workspace --output-format text -p 'summarize current repository'
```

### --bypass-permission=true

```sh
autop -c claudep --bypass-permission=true -- 'summarize current repository'
```

```sh
ANTHROPIC_AUTH_TOKEN="$AGENTSDK_PROXY_API_KEY" claude --dangerously-skip-permissions --settings ~/projects/cc-plugin/config/proxy.json --model gemini-3.5-flash --effort high --add-dir /workspace --output-format text -p 'summarize current repository'
```

### --model 覆寫

```sh
autop -c claudep --model gpt-5.6-sol -- 'summarize current repository'
```

```sh
ANTHROPIC_AUTH_TOKEN="$AGENTSDK_PROXY_API_KEY" claude --settings ~/projects/cc-plugin/config/proxy.json --model gpt-5.6-sol --effort high --add-dir /workspace --output-format text -p 'summarize current repository'
```

### --effort 覆寫

```sh
autop -c claudep --effort low -- 'summarize current repository'
```

```sh
ANTHROPIC_AUTH_TOKEN="$AGENTSDK_PROXY_API_KEY" claude --settings ~/projects/cc-plugin/config/proxy.json --model gemini-3.5-flash --effort low --add-dir /workspace --output-format text -p 'summarize current repository'
```

### -t system (template 無 prompt)

```sh
autop -c claudep -t system
```

```sh
ANTHROPIC_AUTH_TOKEN="$AGENTSDK_PROXY_API_KEY" claude --settings ~/projects/cc-plugin/config/proxy.json --model gemini-3.5-flash --effort high --add-dir /workspace --output-format text -p 'run /system-planner for current workspace'
```

### -t system + prompt

```sh
autop -c claudep -t system -- 'focus on dependency boundaries'
```

```sh
ANTHROPIC_AUTH_TOKEN="$AGENTSDK_PROXY_API_KEY" claude --settings ~/projects/cc-plugin/config/proxy.json --model gemini-3.5-flash --effort high --add-dir /workspace --output-format text -p 'run /system-planner for current workspace focus on dependency boundaries'
```

### 全 flag 組合

```sh
autop -c claudep -t auto-evolving --bypass-permission=true --model gpt-5.6-sol --effort low -- 'focus on dependency boundaries'
```

```sh
ANTHROPIC_AUTH_TOKEN="$AGENTSDK_PROXY_API_KEY" claude --dangerously-skip-permissions --settings ~/projects/cc-plugin/config/proxy.json --model gpt-5.6-sol --effort low --add-dir /workspace --output-format text -p 'run /auto-evolving for current workspace focus on dependency boundaries'
```

### prompt 由 stdin 傳入

```sh
autop -c claudep
```

```sh
ANTHROPIC_AUTH_TOKEN="$AGENTSDK_PROXY_API_KEY" claude --settings ~/projects/cc-plugin/config/proxy.json --model gemini-3.5-flash --effort high --add-dir /workspace --output-format text -p 'review the staged diff'
```

## claudew

driver `claude` / command `claude` / credential `TIKTOK_API_KEY → ANTHROPIC_AUTH_TOKEN`

### 預設 model 與 effort

```sh
autop -c claudew -- 'summarize current repository'
```

```sh
ANTHROPIC_AUTH_TOKEN="$TIKTOK_API_KEY" claude --settings ~/projects/cc-plugin/config/llmbox.json --model minimax-m3 --effort xhigh --add-dir /workspace --output-format text -p 'summarize current repository'
```

### --bypass-permission=true

```sh
autop -c claudew --bypass-permission=true -- 'summarize current repository'
```

```sh
ANTHROPIC_AUTH_TOKEN="$TIKTOK_API_KEY" claude --dangerously-skip-permissions --settings ~/projects/cc-plugin/config/llmbox.json --model minimax-m3 --effort xhigh --add-dir /workspace --output-format text -p 'summarize current repository'
```

### --model 覆寫

```sh
autop -c claudew --model glm-5.2 -- 'summarize current repository'
```

```sh
ANTHROPIC_AUTH_TOKEN="$TIKTOK_API_KEY" claude --settings ~/projects/cc-plugin/config/llmbox.json --model glm-5.2 --effort xhigh --add-dir /workspace --output-format text -p 'summarize current repository'
```

### --effort 覆寫

```sh
autop -c claudew --effort low -- 'summarize current repository'
```

```sh
ANTHROPIC_AUTH_TOKEN="$TIKTOK_API_KEY" claude --settings ~/projects/cc-plugin/config/llmbox.json --model minimax-m3 --effort low --add-dir /workspace --output-format text -p 'summarize current repository'
```

### -t system (template 無 prompt)

```sh
autop -c claudew -t system
```

```sh
ANTHROPIC_AUTH_TOKEN="$TIKTOK_API_KEY" claude --settings ~/projects/cc-plugin/config/llmbox.json --model minimax-m3 --effort xhigh --add-dir /workspace --output-format text -p 'run /system-planner for current workspace'
```

### -t system + prompt

```sh
autop -c claudew -t system -- 'focus on dependency boundaries'
```

```sh
ANTHROPIC_AUTH_TOKEN="$TIKTOK_API_KEY" claude --settings ~/projects/cc-plugin/config/llmbox.json --model minimax-m3 --effort xhigh --add-dir /workspace --output-format text -p 'run /system-planner for current workspace focus on dependency boundaries'
```

### 全 flag 組合

```sh
autop -c claudew -t auto-evolving --bypass-permission=true --model glm-5.2 --effort low -- 'focus on dependency boundaries'
```

```sh
ANTHROPIC_AUTH_TOKEN="$TIKTOK_API_KEY" claude --dangerously-skip-permissions --settings ~/projects/cc-plugin/config/llmbox.json --model glm-5.2 --effort low --add-dir /workspace --output-format text -p 'run /auto-evolving for current workspace focus on dependency boundaries'
```

### prompt 由 stdin 傳入

```sh
autop -c claudew
```

```sh
ANTHROPIC_AUTH_TOKEN="$TIKTOK_API_KEY" claude --settings ~/projects/cc-plugin/config/llmbox.json --model minimax-m3 --effort xhigh --add-dir /workspace --output-format text -p 'review the staged diff'
```

## codex

driver `codex` / command `codex` / credential `oauth`

### 預設 model 與 effort

```sh
autop -c codex -- 'summarize current repository'
```

```sh
printf '%s' 'summarize current repository' | codex exec --model gpt-5.6-sol -c 'model_reasoning_effort="xhigh"' -C /workspace -
```

### --bypass-permission=true

```sh
autop -c codex --bypass-permission=true -- 'summarize current repository'
```

```sh
printf '%s' 'summarize current repository' | codex exec --dangerously-bypass-approvals-and-sandbox --model gpt-5.6-sol -c 'model_reasoning_effort="xhigh"' -C /workspace -
```

### --model 覆寫

```sh
autop -c codex --model gpt-5.6-luna -- 'summarize current repository'
```

```sh
printf '%s' 'summarize current repository' | codex exec --model gpt-5.6-luna -c 'model_reasoning_effort="xhigh"' -C /workspace -
```

### --effort 覆寫

```sh
autop -c codex --effort low -- 'summarize current repository'
```

```sh
printf '%s' 'summarize current repository' | codex exec --model gpt-5.6-sol -c 'model_reasoning_effort="low"' -C /workspace -
```

### -t system (template 無 prompt)

```sh
autop -c codex -t system
```

```sh
printf '%s' 'run $system-planner for current workspace' | codex exec --model gpt-5.6-sol -c 'model_reasoning_effort="xhigh"' -C /workspace -
```

### -t system + prompt

```sh
autop -c codex -t system -- 'focus on dependency boundaries'
```

```sh
printf '%s' 'run $system-planner for current workspace focus on dependency boundaries' | codex exec --model gpt-5.6-sol -c 'model_reasoning_effort="xhigh"' -C /workspace -
```

### 全 flag 組合

```sh
autop -c codex -t auto-evolving --bypass-permission=true --model gpt-5.6-luna --effort low -- 'focus on dependency boundaries'
```

```sh
printf '%s' 'run $auto-evolving for current workspace focus on dependency boundaries' | codex exec --dangerously-bypass-approvals-and-sandbox --model gpt-5.6-luna -c 'model_reasoning_effort="low"' -C /workspace -
```

### prompt 由 stdin 傳入

```sh
autop -c codex
```

```sh
printf '%s' 'review the staged diff' | codex exec --model gpt-5.6-sol -c 'model_reasoning_effort="xhigh"' -C /workspace -
```

## grok

driver `grok` / command `grok` / credential `oauth`

### 預設 model 與 effort

```sh
autop -c grok -- 'summarize current repository'
```

```sh
grok --model grok-4.5 --reasoning-effort high --cwd /workspace --single 'summarize current repository'
```

### --bypass-permission=true

```sh
autop -c grok --bypass-permission=true -- 'summarize current repository'
```

```sh
grok --always-approve --permission-mode auto --model grok-4.5 --reasoning-effort high --cwd /workspace --single 'summarize current repository'
```

### --model 覆寫

```sh
autop -c grok --model grok-4.5 -- 'summarize current repository'
```

```sh
grok --model grok-4.5 --reasoning-effort high --cwd /workspace --single 'summarize current repository'
```

### --effort 覆寫

```sh
autop -c grok --effort low -- 'summarize current repository'
```

```sh
grok --model grok-4.5 --reasoning-effort low --cwd /workspace --single 'summarize current repository'
```

### -t system (template 無 prompt)

```sh
autop -c grok -t system
```

```sh
grok --model grok-4.5 --reasoning-effort high --cwd /workspace --single 'run /system-planner for current workspace'
```

### -t system + prompt

```sh
autop -c grok -t system -- 'focus on dependency boundaries'
```

```sh
grok --model grok-4.5 --reasoning-effort high --cwd /workspace --single 'run /system-planner for current workspace focus on dependency boundaries'
```

### 全 flag 組合

```sh
autop -c grok -t auto-evolving --bypass-permission=true --model grok-4.5 --effort low -- 'focus on dependency boundaries'
```

```sh
grok --always-approve --permission-mode auto --model grok-4.5 --reasoning-effort low --cwd /workspace --single 'run /auto-evolving for current workspace focus on dependency boundaries'
```

### prompt 由 stdin 傳入

```sh
autop -c grok
```

```sh
grok --model grok-4.5 --reasoning-effort high --cwd /workspace --single 'review the staged diff'
```
