#!/bin/bash

set -euo pipefail

# ============================================================================
# GIT COMMIT SECRET SCANNER HOOK
# Scans staged git files for hardcoded secrets before allowing commits.
# ============================================================================

detected_secrets=()
permission="allow"
agent_message=""

# Ignore patterns
ignore_patterns=("*.example.*" "CHANGELOG.md" "*.md")

# Logging
log() {
    local msg="[$(date +"%Y-%m-%d %H:%M:%S")] [scan-secrets-on-commit] $*"
    if [[ "${DEBUG:-}" == "1" ]]; then
        echo "$msg" >&2
    else
        echo "$msg" >> "/tmp/1password-cursor-hooks.log" 2>/dev/null || true
    fi
}

# JSON escaping
escape_json() {
    local str="$1"
    # JSON string escaping (handles most common cases)
    # Escape backslashes, quotes, and control characters
    str=$(echo "$str" | sed 's/\\/\\\\/g')
    str=$(echo "$str" | sed 's/"/\\"/g')
    str=$(echo "$str" | sed 's/\n/\\n/g')
    str=$(echo "$str" | sed 's/\r/\\r/g')
    str=$(echo "$str" | sed 's/\t/\\t/g')
    echo "$str"
}

# Check if file should be ignored
should_ignore() {
    local file="$1"
    local basename=$(basename "$file")
    for pattern in "${ignore_patterns[@]}"; do
        [[ "$basename" == $pattern ]] && return 0
    done
    return 1
}

# Check if text file
is_text_file() {
    local file="$1"
    if command -v file &> /dev/null; then
        local mime=$(file -b --mime-type "$file" 2>/dev/null || echo "")
        [[ "$mime" == text/* ]] || [[ "$mime" == application/json ]] && return 0
    fi
    ! grep -q $'\0' "$file" 2>/dev/null
}

# Scan file for secrets
scan_file() {
    local file="$1"
    local line_num=0

    should_ignore "$file" && return 0
    is_text_file "$file" || return 0

    log "Scanning: $file"

    while IFS= read -r line || [[ -n "$line" ]]; do
        ((line_num++))
        [[ -z "${line// /}" ]] && continue
        [[ "$line" =~ ^[[:space:]]*# ]] && continue
        [[ "$line" =~ ^[[:space:]]*// ]] && continue

        # Detect secrets
        if echo "$line" | grep -qiE "(api[_-]?key|apikey)[[:space:]]*[=:][[:space:]]*['\"]?[a-zA-Z0-9]{20,}"; then
            local match=$(echo "$line" | grep -oiE "(api[_-]?key|apikey)[[:space:]]*[=:][[:space:]]*['\"]?[a-zA-Z0-9]{20,}" | head -1)
            detected_secrets+=("$file|$line_num|API_KEY|$match")
        elif echo "$line" | grep -qiE "(password|passwd|pwd)[[:space:]]*[=:][[:space:]]*['\"]?[^'\"]{8,}"; then
            local match=$(echo "$line" | grep -oiE "(password|passwd|pwd)[[:space:]]*[=:][[:space:]]*['\"]?[^'\"]{8,}" | head -1 | cut -c1-50)
            detected_secrets+=("$file|$line_num|PASSWORD|$match")
        elif echo "$line" | grep -qiE "(token|access_token|refresh_token)[[:space:]]*[=:][[:space:]]*['\"]?[a-zA-Z0-9]{20,}"; then
            local match=$(echo "$line" | grep -oiE "(token|access_token|refresh_token)[[:space:]]*[=:][[:space:]]*['\"]?[a-zA-Z0-9]{20,}" | head -1)
            detected_secrets+=("$file|$line_num|TOKEN|$match")
        elif echo "$line" | grep -qiE "(aws[_-]?access[_-]?key|aws[_-]?secret[_-]?key|secret[_-]?key|private[_-]?key)[[:space:]]*[=:][[:space:]]*['\"]?[a-zA-Z0-9]{20,}"; then
            local match=$(echo "$line" | grep -oiE "(aws[_-]?access[_-]?key|aws[_-]?secret[_-]?key|secret[_-]?key|private[_-]?key)[[:space:]]*[=:][[:space:]]*['\"]?[a-zA-Z0-9]{20,}" | head -1)
            detected_secrets+=("$file|$line_num|SECRET|$match")
        fi
    done < "$file"
}

# Main execution
log "Starting secret scan..."

json_input=$(cat)
if ! echo "$json_input" | grep -qiE "git[[:space:]]+commit"; then
    log "Not a git commit, allowing"
    echo '{"permission": "allow"}'
    exit 0
fi

log "Git commit detected"

# Get staged files
if ! command -v git &> /dev/null || ! git rev-parse --git-dir &> /dev/null; then
    log "Not a git repo, allowing"
    echo '{"permission": "allow"}'
    exit 0
fi

staged_files=$(git diff --cached --name-only --diff-filter=ACM 2>/dev/null || true)
if [[ -z "$staged_files" ]]; then
    log "No staged files, allowing"
    echo '{"permission": "allow"}'
    exit 0
fi

log "Found $(echo "$staged_files" | wc -l | tr -d ' ') staged file(s)"

# Scan files
while IFS= read -r file || [[ -n "$file" ]]; do
    [[ -z "$file" ]] && continue
    [[ "$file" != /* ]] && file="${PWD}/${file}"
    [[ -f "$file" ]] && scan_file "$file"
done <<< "$staged_files"

# Build response
if [[ ${#detected_secrets[@]} -gt 0 ]]; then
    permission="deny"
    log "Denying: ${#detected_secrets[@]} secret(s) detected"

    msg="Potential secrets detected in staged files. Please remove hardcoded credentials and use 1Password instead.\n\nDetected secrets:\n"

    for secret in "${detected_secrets[@]}"; do
        IFS='|' read -r file line type match <<< "$secret"
        msg="${msg}  - $file:$line - $type ($match)\n"
    done

    msg="${msg}\nSuggestion: Store secrets in 1Password and reference them via environment variables instead."
    agent_message="$msg"

    msg_json=$(escape_json "$agent_message")
    echo "{\"permission\": \"deny\", \"agent_message\": \"$msg_json\"}"
else
    log "No secrets detected, allowing"
    echo '{"permission": "allow"}'
fi

exit 0
