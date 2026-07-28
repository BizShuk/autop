#!/usr/bin/env bash

set -euo pipefail

# Run the real wizard non-interactively for every enabled client. Each run
# accepts the profile defaults and prints both complete command previews.
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$script_dir/.." && pwd)"
work_dir="${AUTOP_WIZARD_WORKDIR:-$project_root}"

if [[ ! -d "$work_dir" ]]; then
    printf 'AUTOP_WIZARD_WORKDIR is not a directory: %s\n' "$work_dir" >&2
    exit 1
fi

temporary_dir=""
autop_bin="${AUTOP_BIN:-}"
if [[ -z "$autop_bin" ]]; then
    temporary_dir="$(mktemp -d)"
    trap 'rm -f -- "$temporary_dir/autop"; rmdir "$temporary_dir"' EXIT
    go build -o "$temporary_dir/autop" "$project_root"
    autop_bin="$temporary_dir/autop"
fi

if [[ ! -x "$autop_bin" ]]; then
    printf 'AUTOP_BIN is not executable: %s\n' "$autop_bin" >&2
    exit 1
fi

clients=(agy claude claudem claudep claudew codex grok)
default_prompt=$'/test \'\' ""'
prompt="${AUTOP_WIZARD_PROMPT:-$default_prompt}"
cron="${AUTOP_WIZARD_CRON:-N}"

for client in "${clients[@]}"; do
    output="$(
        cd "$work_dir"
        printf '%s\n' \
            "$client" \
            "" \
            "" \
            "" \
            "" \
            "$prompt" \
            "$cron" |
            "$autop_bin" wizard
    )"

    printf '\n=== %s ===\n%s\n' "$client" "$output"

    if ! grep -Fq "Original command (autop):" <<<"$output"; then
        printf 'wizard output for %s is missing the full autop command\n' "$client" >&2
        exit 1
    fi
    if ! grep -Fq "Execute command (" <<<"$output"; then
        printf 'wizard output for %s is missing the full execute command\n' "$client" >&2
        exit 1
    fi
done
