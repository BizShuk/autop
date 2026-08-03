#!/usr/bin/env bash

set -euo pipefail

# Print the original autop command and the driver-mapped command line for every
# -c / -t / --model / --bypass-permission combination. Every run uses --dry-run,
# so no child CLI is ever started and no credential is ever resolved.
#
# Every combination is still printed when it fails; the failing autop error is
# shown in place of the command line and the script exits 1 with a final count.
#
# Overrides (space separated lists):
#   AUTOP_BIN                 prebuilt autop binary, default: build from source
#   AUTOP_MATRIX_WORKDIR      working directory for each run, default: repo root
#   AUTOP_MATRIX_CLIENTS      default: every enabled client in the merged config
#   AUTOP_MATRIX_TEMPLATES    default: "-" (no template) plus every template
#   AUTOP_MATRIX_MODELS       default: every model the current client supports
#   AUTOP_MATRIX_BYPASS       default: "true false"
#   AUTOP_MATRIX_PROMPT       default: "summarize current repository"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$script_dir/.." && pwd)"
work_dir="${AUTOP_MATRIX_WORKDIR:-$project_root}"

if [[ ! -d "$work_dir" ]]; then
    printf 'AUTOP_MATRIX_WORKDIR is not a directory: %s\n' "$work_dir" >&2
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

prompt="${AUTOP_MATRIX_PROMPT:-summarize current repository}"
no_template="-"

# quote renders one argument the way autop renders it in a command preview:
# bare when it is shell safe, single quoted with escaped quotes otherwise.
quote() {
    local value="$1"
    if [[ -n "$value" && "$value" =~ ^[A-Za-z0-9_@%+=:,./-]+$ ]]; then
        printf '%s' "$value"
        return
    fi
    printf "'%s'" "${value//\'/\'\"\'\"\'}"
}

config_table="$("$autop_bin" config)"

# config_values prints the raw value column for one merged config key.
config_values() {
    awk -v key="$1" '$1 == key { sub(/^[^ ]+[ \t]+/, ""); print }' <<<"$config_table"
}

read_list() {
    local -n target="$1"
    target=()
    local item
    while IFS= read -r item; do
        [[ -n "$item" ]] && target+=("$item")
    done
}

clients=()
if [[ -n "${AUTOP_MATRIX_CLIENTS:-}" ]]; then
    read -r -a clients <<<"$AUTOP_MATRIX_CLIENTS"
else
    read_list clients < <(
        awk '$1 ~ /^clients\.[^.]+\.disabled$/ && $2 == "false" {
            split($1, part, "."); print part[2]
        }' <<<"$config_table" | sort -u
    )
fi

templates=()
if [[ -n "${AUTOP_MATRIX_TEMPLATES:-}" ]]; then
    read -r -a templates <<<"$AUTOP_MATRIX_TEMPLATES"
else
    templates=("$no_template")
    discovered_templates=()
    read_list discovered_templates < <(
        awk '$1 ~ /^templates\./ { split($1, part, "."); print part[2] }' <<<"$config_table" | sort -u
    )
    templates+=("${discovered_templates[@]}")
fi

bypass_values=()
read -r -a bypass_values <<<"${AUTOP_MATRIX_BYPASS:-true false}"

# client_models lists the models one client accepts: its default model first,
# then the configured models list, matching autop's own override validation.
client_models() {
    local client="$1"
    local models_json default_model
    default_model="$(config_values "clients.$client.model")"
    models_json="$(config_values "clients.$client.models")"

    {
        [[ -n "$default_model" ]] && printf '%s\n' "$default_model"
        if [[ -n "$models_json" ]]; then
            jq -r '.[]' <<<"$models_json"
        fi
    } | awk 'NF && !seen[$0]++'
}

total=0
failed=0

for client in "${clients[@]}"; do
    models=()
    if [[ -n "${AUTOP_MATRIX_MODELS:-}" ]]; then
        read -r -a models <<<"$AUTOP_MATRIX_MODELS"
    else
        read_list models < <(client_models "$client")
    fi
    if [[ ${#models[@]} -eq 0 ]]; then
        printf '\n### %s\n!! client has no model to iterate\n' "$client"
        failed=$((failed + 1))
        continue
    fi

    for template in "${templates[@]}"; do
        for model in "${models[@]}"; do
            for bypass in "${bypass_values[@]}"; do
                args=(--dry-run -c "$client")
                [[ "$template" != "$no_template" ]] && args+=(-t "$template")
                args+=(--model "$model" "--bypass-permission=$bypass" -- "$prompt")

                origin="autop"
                for arg in "${args[@]}"; do
                    [[ "$arg" == "--dry-run" ]] && continue
                    origin+=" $(quote "$arg")"
                done

                total=$((total + 1))
                if executed="$(cd "$work_dir" && "$autop_bin" "${args[@]}" 2>&1)"; then
                    status="ok"
                else
                    status="error"
                    failed=$((failed + 1))
                fi

                printf '\n### %s | t=%s | model=%s | bypass=%s | %s\n' \
                    "$client" "$template" "$model" "$bypass" "$status"
                printf 'autop: %s\n' "$origin"
                printf 'exec : %s\n' "$executed"
            done
        done
    done
done

printf '\n=== %d combinations, %d failed ===\n' "$total" "$failed"
[[ "$failed" -eq 0 ]]
