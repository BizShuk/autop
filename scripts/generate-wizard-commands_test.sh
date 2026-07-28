#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$script_dir/.." && pwd)"

output="$(
    AUTOP_BIN="$script_dir/testdata/fake-autop" \
    AUTOP_WIZARD_WORKDIR="$project_root" \
        "$script_dir/generate-wizard-commands.sh"
)"

clients=(agy claude claudem claudep claudew codex grok)
for client in "${clients[@]}"; do
    grep -Fq "=== $client ===" <<<"$output"
    grep -Fq "autop -c $client -- '/test '\"'\"''\"'\"' \"\"'" <<<"$output"
done

[[ "$(grep -Fc "Original command (autop):" <<<"$output")" -eq 7 ]]
[[ "$(grep -Fc "Execute command (" <<<"$output")" -eq 7 ]]

printf '%s\n' "generate-wizard-commands test passed"
