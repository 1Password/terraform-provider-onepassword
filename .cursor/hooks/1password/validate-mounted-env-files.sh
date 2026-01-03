#!/bin/bash

set -euo pipefail

# ============================================================================
# TABLE OF CONTENTS
# ============================================================================
#
# - Global Variables
# - Core Utility Functions (logging, JSON escaping)
# - System & Path Utility Functions (OS detection, path normalization)
# - 1Password Database Functions (finding and querying database)
# - Mount Parsing & Validation Functions (parsing mount data, validation)
# - TOML Parsing Functions
# - Main Execution Logic
# - Permission Decision Logic
#
# ============================================================================

# ============================================================================
# GLOBAL VARIABLES
# ============================================================================

# Array of "mount_path|environment_name"
# Local .env files that are created but not enabled in 1Password.
disabled_mounts=()

# Array of "mount_path|environment_name"
# Local .env files that are created but not valid (file is not present or not a FIFO).
invalid_mounts=()

# Array of mount paths
# Local .env files that are required by TOML but missing or invalid.
required_mounts=()

# The final permission decision to return to Cursor.
permission="allow"
# The message for the agent to interpret if the permission is denied.
agent_message=""

# ============================================================================
# CORE UTILITY FUNCTIONS
# ============================================================================

# Log function for debugging
log() {
    local timestamp
    timestamp=$(date +"%Y-%m-%d %H:%M:%S" 2>/dev/null || echo "$(date +%s)")
    local log_message="[${timestamp}] [validate-mounted-env-files] $*"

    if [[ "${DEBUG:-}" == "1" ]]; then
        # If DEBUG=1, echo directly to terminal (stderr for logs)
        echo "$log_message" >&2
    else
        # Otherwise, send to log file
        local log_file="/tmp/1password-cursor-hooks.log"
        # Ensure log file is writable (create if needed, ignore errors if we can't write)
        echo "$log_message" >> "$log_file" 2>/dev/null || true
    fi
}

# Escape JSON string value (returns escaped string without quotes)
escape_json_string() {
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

# Parse JSON input from stdin and extract workspace_roots array
# Returns workspace root paths, one per line
parse_json_workspace_roots() {
    local json_input
    json_input=$(cat)

    # Extract workspace_roots array values
    # Find the line(s) containing "workspace_roots" and extract the array
    # Handle both single-line and multi-line arrays
    local in_array=false
    local array_lines=""

    while IFS= read -r line || [[ -n "$line" ]]; do
        # Check if this line starts the workspace_roots array
        if echo "$line" | grep -qE '"workspace_roots"[[:space:]]*:[[:space:]]*\['; then
            in_array=true
            # Extract content after the opening bracket
            array_lines="${line#*\[}"
            # Check if array closes on same line
            if echo "$array_lines" | grep -qE '\]'; then
                array_lines="${array_lines%\]*}"
                break
            fi
        elif [[ "$in_array" == "true" ]]; then
            # Check if this line closes the array
            if echo "$line" | grep -qE '\]'; then
                array_lines="${array_lines} ${line%\]*}"
                break
            else
                array_lines="${array_lines} ${line}"
            fi
        fi
    done <<< "$json_input"

    # Extract quoted strings from the array content
    echo "$array_lines" | grep -oE '"[^"]+"' | \
        sed 's/^"//;s/"$//' | \
        sed '/^$/d'
}

# Output JSON response with permission decision
output_response() {
    log "Decision: $permission"
    if [[ "$permission" == "deny" ]]; then
        log "Agent message: $agent_message"

        agent_msg_json=$(escape_json_string "$agent_message")

        cat << EOF
{
  "permission": "deny",
  "agent_message": "$agent_msg_json"
}
EOF
    else
        cat << EOF
{
  "permission": "allow"
}
EOF
    fi
}

# ============================================================================
# SYSTEM & PATH UTILITY FUNCTIONS
# ============================================================================

# Detect operating system
detect_os() {
    case "$(uname -s)" in
        Darwin*)
            echo "macos"
            ;;
        Linux*)
            echo "unix"
            ;;
        *)
            log "Warning: Unsupported OS: $(uname -s)"
            echo "unknown"
            ;;
    esac
}

# Validate and sanitize path to prevent command injection
# Returns 0 if path is safe, 1 if unsafe
validate_path() {
    local path="$1"

    # Check for empty path
    [[ -z "$path" ]] && return 1

    # Check for command substitution patterns
    # $() command substitution
    if [[ "$path" =~ \$\( ]] || [[ "$path" =~ \$\{ ]]; then
        return 1
    fi

    # Backtick command substitution
    if [[ "$path" =~ \` ]]; then
        return 1
    fi

    # Check for semicolons, pipes, ampersands, and other command separators
    if [[ "$path" =~ [\;\|\&\<\>] ]]; then
        return 1
    fi

    # Check for control characters that could break commands
    # Remove all printable characters; if anything remains, there are control chars
    local non_printable
    non_printable=$(printf '%s' "$path" | tr -d '[:print:]' 2>/dev/null || echo "")
    if [[ -n "$non_printable" ]]; then
        return 1
    fi

    # Path is considered safe if it passes all checks
    return 0
}

# Normalize path for cross-platform compatibility
normalize_path() {
    local path="$1"
    local normalized normalized_dir file_part dir_part

    # Validate path before using it with cd to prevent command injection
    if ! validate_path "$path"; then
        log "Warning: Unsafe path detected, skipping normalization: ${path}"
        echo "$path"
        return 0
    fi

    # Normalize a given path using cd
    # This resolves . and .. components and symlinks for existing paths
    if [[ -d "$path" ]]; then
        # For directories, use cd to resolve
        normalized=$(cd "$path" && pwd 2>/dev/null)
        if [[ -n "$normalized" ]]; then
            echo "$normalized"
            return 0
        fi
    elif [[ -f "$path" ]] || [[ -p "$path" ]]; then
        # For files/FIFOs, resolve the directory part
        dir_part=$(dirname "$path")
        file_part=$(basename "$path")

        # Validate dir_part before using with cd
        if validate_path "$dir_part" && [[ -d "$dir_part" ]]; then
            normalized_dir=$(cd "$dir_part" && pwd 2>/dev/null)
            if [[ -n "$normalized_dir" ]]; then
                echo "${normalized_dir}/${file_part}"
                return 0
            fi
        fi
    else
        # Attempt to normalize non-existent paths (e.g., with .. components)
        dir_part=$(dirname "$path")
        file_part=$(basename "$path")

        if validate_path "$dir_part" && [[ -d "$dir_part" ]]; then
            normalized_dir=$(cd "$dir_part" && pwd 2>/dev/null)
            if [[ -n "$normalized_dir" ]]; then
                echo "${normalized_dir}/${file_part}"
                return 0
            fi
        fi
    fi

    # Last resort: return path as-is
    echo "$path"
}

# ============================================================================
# 1PASSWORD DATABASE FUNCTIONS
# ============================================================================

# Find 1Password database based on operating system
find_1password_db() {
    local os_type="$1"
    local home_path="${HOME}"
    local db_paths=()

    if [[ "$os_type" == "macos" ]]; then
        db_paths=(
            "${home_path}/Library/Group Containers/2BUA8C4S2C.com.1password/Library/Application Support/1Password/Data/1Password.sqlite"
        )
    elif [[ "$os_type" == "unix" ]]; then
        db_paths=(
            "${home_path}/.config/1Password/1Password.sqlite"
            "${home_path}/snap/1password/current/.config/1Password/1Password.sqlite"
            "${home_path}/.var/app/com.onepassword.OnePassword/config/1Password/1Password.sqlite"
        )
    fi

    for db_path in "${db_paths[@]}"; do
        if [[ -f "$db_path" ]]; then
            echo "$db_path"
            return 0
        fi
    done

    return 1
}

# Query 1Password database for mounts
query_mounts() {
    local db_path="$1"

    if ! command -v sqlite3 &> /dev/null; then
        log "Warning: sqlite3 not found, cannot query 1Password database"
        return 1
    fi

    # Check if database is readable
    if [[ ! -r "$db_path" ]]; then
        log "Warning: 1Password database is not readable: ${db_path}"
        return 1
    fi

    # Check if database file exists and is a valid SQLite database
    if ! sqlite3 "$db_path" "SELECT 1;" &>/dev/null; then
        log "Warning: 1Password database appears to be invalid or locked: ${db_path}"
        return 1
    fi

    # Query for mount entries
    # Suppress errors but capture output
    local result
    result=$(sqlite3 "$db_path" "SELECT hex(data) FROM objects_associated WHERE key_name LIKE 'dev-environment-mount/%';" 2>/dev/null)
    local exit_code=$?

    if [[ $exit_code -ne 0 ]]; then
        log "Warning: Failed to query 1Password database (exit code: $exit_code)"
        return 1
    fi

    # Return result even if empty (empty string is valid - means no mounts)
    echo "$result"
    return 0
}

# ============================================================================
# MOUNT PARSING & VALIDATION FUNCTIONS
# ============================================================================

# Check if mount path is within project
is_project_mount() {
    local mount_path="$1"
    local project_path="$2"

    # Normalize paths for comparison
    local normalized_mount normalized_project

    normalized_mount=$(normalize_path "$mount_path")
    normalized_project=$(normalize_path "$project_path")

    # Ensure both paths end with / for consistent comparison
    [[ "$normalized_project" != */ ]] && normalized_project="${normalized_project}/"

    # Check if mount path starts with project path (mount is within project)
    # Also check original paths in case normalization failed
    if [[ "$normalized_mount" == "$normalized_project"* ]] || \
       [[ "$normalized_mount" == "$project_path" ]] || \
       [[ "$mount_path" == "$project_path"* ]] || \
       [[ "$mount_path" == "$project_path" ]]; then
        return 0
    fi

    return 1
}

# Decode hex string to JSON
hex_to_json() {
    local hex="$1"
    # Remove any whitespace/newlines
    hex=$(echo "$hex" | tr -d '[:space:]')

    # Skip if empty
    [[ -z "$hex" ]] && return 1

    # Use printf with escaped hex
    # Convert hex pairs to \x escaped format
    local escaped_hex decoded
    escaped_hex=$(echo "$hex" | sed 's/\(..\)/\\x\1/g')

    decoded=$(printf "%b" "$escaped_hex" 2>/dev/null || echo "")
    if [[ -n "$decoded" ]] && [[ "$decoded" != "$escaped_hex" ]]; then
        echo "$decoded"
        return 0
    fi

    return 1
}

# Parse mount JSON, extract mount path, enabled status, environment name, uuid, and environmentUuid
parse_mount() {
    local hex_data="$1"
    local json_data

    json_data=$(hex_to_json "$hex_data")

    if [[ -z "$json_data" ]]; then
        return 1
    fi

    # Extract mountPath, isEnabled, environmentName, uuid, and environmentUuid from JSON
    # Note: This may not handle all JSON edge cases (escaped quotes, etc.)
    # but should work for typical 1Password mount JSON structures
    local mount_path is_enabled environment_name uuid environment_uuid

    # Extract mountPath - handle both BSD and GNU sed
    mount_path=$(echo "$json_data" | grep -oE '"mountPath"[[:space:]]*:[[:space:]]*"[^"]*"' | sed -E 's/.*"mountPath"[[:space:]]*:[[:space:]]*"([^"]*)".*/\1/' 2>/dev/null || \
                 echo "$json_data" | grep -o '"mountPath"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/.*"mountPath"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null || echo "")

    # Check for isEnabled: true or false
    if echo "$json_data" | grep -qE '"isEnabled"[[:space:]]*:[[:space:]]*true'; then
        is_enabled="true"
    else
        is_enabled="false"
    fi

    # Extract environmentName - handle both BSD and GNU sed
    environment_name=$(echo "$json_data" | grep -oE '"environmentName"[[:space:]]*:[[:space:]]*"[^"]*"' | sed -E 's/.*"environmentName"[[:space:]]*:[[:space:]]*"([^"]*)".*/\1/' 2>/dev/null || \
                      echo "$json_data" | grep -o '"environmentName"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/.*"environmentName"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null || echo "")

    # Extract uuid - handle both BSD and GNU sed
    uuid=$(echo "$json_data" | grep -oE '"uuid"[[:space:]]*:[[:space:]]*"[^"]*"' | sed -E 's/.*"uuid"[[:space:]]*:[[:space:]]*"([^"]*)".*/\1/' 2>/dev/null || \
           echo "$json_data" | grep -o '"uuid"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/.*"uuid"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null || echo "")

    # Extract environmentUuid - handle both BSD and GNU sed
    environment_uuid=$(echo "$json_data" | grep -oE '"environmentUuid"[[:space:]]*:[[:space:]]*"[^"]*"' | sed -E 's/.*"environmentUuid"[[:space:]]*:[[:space:]]*"([^"]*)".*/\1/' 2>/dev/null || \
                      echo "$json_data" | grep -o '"environmentUuid"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/.*"environmentUuid"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null || echo "")

    if [[ -n "$mount_path" ]]; then
        echo "$mount_path|$is_enabled|$environment_name|$uuid|$environment_uuid"
        return 0
    fi

    return 1
}

# ============================================================================
# TOML PARSING FUNCTIONS
# ============================================================================

# Remove comments and trim whitespace from a TOML line
normalize_toml_line() {
    local line="$1"

    # Remove comments (everything after #)
    line="${line%%#*}"
    # Trim leading/trailing whitespace
    line=$(echo "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

    echo "$line"
}

# Check if TOML file has a mount_paths field defined at top level
# Returns 0 if mount_paths field exists, 1 otherwise
has_toml_mount_paths_field() {
    local toml_file="$1"

    # Check for TOML configuration
    if [[ ! -f "$toml_file" ]]; then
        return 1
    fi

    while IFS= read -r raw_line || [[ -n "$raw_line" ]]; do
        local line
        line=$(normalize_toml_line "$raw_line")

        # Skip empty lines
        [[ -z "$line" ]] && continue

        # If we hit a section header, we're no longer at top level
        if [[ "$line" =~ ^\[\[.*\]\] ]] || [[ "$line" =~ ^\[.*\] ]]; then
            break
        fi

        # Detect 'mount_paths' field at top level
        if [[ "$line" =~ ^mount_paths[[:space:]]*= ]]; then
            return 0
        fi
    done < "$toml_file"

    return 1
}

# Parse TOML file and extract mount paths from top-level mount_paths field
# Returns newline-separated list of mount paths
# Returns empty string (but exit code 0) if mount_paths = []
# Returns exit code 1 if mount_paths field doesn't exist
parse_toml_mount_paths() {
    local toml_file="$1"

    if [[ ! -f "$toml_file" ]]; then
        return 1
    fi

    # Pure bash TOML parsing for environments entries
    # Handles formats like:
    #   mount_paths = [".env", "billing.env"]
    #   mount_paths = [
    #     ".env",
    #     "billing.env"
    #   ]
    #   mount_paths = []
    local in_mount_paths_array=false
    local mount_paths=""
    local array_content=""
    local found_mount_paths_field=false

    while IFS= read -r raw_line || [[ -n "$raw_line" ]]; do
        local line
        line=$(normalize_toml_line "$raw_line")

        # Skip empty lines
        [[ -z "$line" ]] && continue

        # If we hit a section header, we're no longer at top level
        if [[ "$line" =~ ^\[\[.*\]\] ]] || [[ "$line" =~ ^\[.*\] ]]; then
            break
        fi

        # Check for mount_paths = [...] on a single line
        if [[ "$line" =~ ^mount_paths[[:space:]]*=[[:space:]]*\[.*\] ]]; then
            found_mount_paths_field=true
            # Extract content between [ and ]
            local array_part="${line#*\[}"
            array_part="${array_part%\]*}"
            array_content="$array_part"
            in_mount_paths_array=false  # Array is complete on one line

            # Extract quoted strings from the array content
            while [[ "$array_content" =~ \"([^\"]+)\" ]]; do
                mount_paths="${mount_paths}${BASH_REMATCH[1]}"$'\n'
                # Remove the matched string and any following comma/whitespace
                array_content="${array_content#*\"${BASH_REMATCH[1]}\"}"
                array_content=$(echo "$array_content" | sed 's/^[[:space:]]*,[[:space:]]*//;s/^[[:space:]]*//')
            done
        # Check for mount_paths = [ (multi-line array start)
        elif [[ "$line" =~ ^mount_paths[[:space:]]*=[[:space:]]*\[ ]]; then
            found_mount_paths_field=true
            in_mount_paths_array=true
            # Extract any content after the opening [
            array_content="${line#*\[}"
            array_content=$(echo "$array_content" | sed 's/^[[:space:]]*//')
            # If array closes on same line, process it
            if [[ "$array_content" =~ \] ]]; then
                array_content="${array_content%\]*}"
                while [[ "$array_content" =~ \"([^\"]+)\" ]]; do
                    mount_paths="${mount_paths}${BASH_REMATCH[1]}"$'\n'
                    array_content="${array_content#*\"${BASH_REMATCH[1]}\"}"
                    array_content=$(echo "$array_content" | sed 's/^[[:space:]]*,[[:space:]]*//;s/^[[:space:]]*//')
                done
                in_mount_paths_array=false
                array_content=""
            fi
        # If we're in a mount_paths array, collect lines until we hit ]
        elif [[ "$in_mount_paths_array" == "true" ]]; then
            # Check if this line closes the array
            if [[ "$line" =~ \] ]]; then
                # Extract content before the closing ]
                local line_content="${line%\]*}"
                array_content="${array_content} ${line_content}"
                # Process the complete array content
                while [[ "$array_content" =~ \"([^\"]+)\" ]]; do
                    mount_paths="${mount_paths}${BASH_REMATCH[1]}"$'\n'
                    array_content="${array_content#*\"${BASH_REMATCH[1]}\"}"
                    array_content=$(echo "$array_content" | sed 's/^[[:space:]]*,[[:space:]]*//;s/^[[:space:]]*//')
                done
                in_mount_paths_array=false
                array_content=""
            else
                # Add this line to array content
                array_content="${array_content} ${line}"
            fi
        fi
    done < "$toml_file"

    # If mount_paths field was found, return success (even if empty)
    if [[ "$found_mount_paths_field" == "true" ]]; then
        # Remove trailing newline and return
        if [[ -n "$mount_paths" ]]; then
            mount_paths=$(echo "$mount_paths" | sed '/^$/d')
            if [[ -n "$mount_paths" ]]; then
                echo "$mount_paths"
            fi
        fi
        return 0
    fi

    return 1
}

# ============================================================================
# MAIN EXECUTION LOGIC
# ============================================================================

# Query 1Password database and check mounts
log "Checking for local .env files mounted by 1Password..."

# Read JSON input from stdin and extract workspace_roots
workspace_roots_input=$(parse_json_workspace_roots)
workspace_roots_array=()

# Build array of workspace roots
while IFS= read -r workspace_root || [[ -n "$workspace_root" ]]; do
    [[ -z "$workspace_root" ]] && continue
    # Normalize the workspace root path
    normalized_root=$(normalize_path "$workspace_root")
    if [[ -n "$normalized_root" ]]; then
        workspace_roots_array+=("$normalized_root")
    fi
done <<< "$workspace_roots_input"

# If no workspace roots found in JSON, log and exit (fail open)
if [[ ${#workspace_roots_array[@]} -eq 0 ]]; then
    log "No workspace_roots found in JSON input, skipping validation"
    output_response
    exit 0
fi

log "Found ${#workspace_roots_array[@]} workspace root(s) to validate"

# Query 1Password database once (shared across all workspace roots)
os_type=$(detect_os)
db_path=""
mount_hex_data=""

if [[ "$os_type" != "unknown" ]]; then
    db_path=$(find_1password_db "$os_type")
    if [[ -n "$db_path" ]]; then
        mount_hex_data=$(query_mounts "$db_path")
    fi
fi

# Process each workspace root
for workspace_root in "${workspace_roots_array[@]}"; do
    log "Processing workspace root: $workspace_root"

    # Check for TOML configuration at this workspace root
    toml_file="${workspace_root}/.1password/environments.toml"
    use_configured_mode=false

    # Check if TOML exists and has mount_paths field
    if [[ -f "$toml_file" ]]; then
        log "Found environments.toml at ${toml_file}, checking for mount_paths field..."

        if has_toml_mount_paths_field "$toml_file"; then
            use_configured_mode=true
            log "environments.toml has mount_paths field defined - validating specified mounts"

            # Parse and validate TOML mount paths
            toml_mounts=$(parse_toml_mount_paths "$toml_file")
            if [[ $? -ne 0 ]]; then
                log "Warning: Failed to parse environments.toml at ${toml_file}, falling back to default mode"
                use_configured_mode=false
            elif [[ -z "$toml_mounts" ]]; then
                log "environments.toml specifies mount_paths = [] - no local .env files to validate for this workspace"
                continue
            fi
        else
            log "environments.toml exists but does not specify a mount_paths field, using default mode (checking all mounts)"
        fi
    else
        log "No environments.toml found at ${workspace_root}/.1password/environments.toml, using default mode (checking all mounts)"
    fi

    # Configured mode: validate only mounts specified in TOML
    if [[ "$use_configured_mode" == "true" ]]; then
        log "Validating local .env files specified in environments.toml for workspace ${workspace_root}..."

        # Build an array of unique normalized paths from TOML
        toml_paths_array=()
        while IFS= read -r mount_path || [[ -n "$mount_path" ]]; do
            [[ -z "$mount_path" ]] && continue

            # Validate path from TOML to prevent command injection
            if ! validate_path "$mount_path"; then
                log "Warning: Unsafe path detected in environments.toml, skipping: ${mount_path}"
                continue
            fi

            # Resolve mount path relative to workspace root
            if [[ "$mount_path" == /* ]]; then
                resolved_path="$mount_path"
            else
                resolved_path="${workspace_root}/${mount_path}"
            fi

            # Normalize the path
            resolved_path=$(normalize_path "$resolved_path")

            # Skip mounts that resolve outside the current workspace root
            if ! is_project_mount "$resolved_path" "$workspace_root"; then
                log "Skipping required mount outside workspace root: \"${resolved_path}\" (workspace: \"${workspace_root}\")"
                continue
            fi

            # Add to array only if not already present
            path_exists=false
            if [[ ${#toml_paths_array[@]} -gt 0 ]]; then
                for existing_path in "${toml_paths_array[@]}"; do
                    if [[ "$existing_path" == "$resolved_path" ]]; then
                        path_exists=true
                        break
                    fi
                done
            fi
            if [[ "$path_exists" == "false" ]]; then
                toml_paths_array+=("$resolved_path")
            fi
        done <<< "$toml_mounts"

        # Check each TOML-specified mount
        if [[ ${#toml_paths_array[@]} -gt 0 ]]; then
            for resolved_path in "${toml_paths_array[@]}"; do
                log "Checking required local .env file from TOML: \"${resolved_path}\""

                # First, check if it's in the database and what its status is
                found_in_db=false
                is_enabled="false"
                environment_name=""

                if [[ -n "$mount_hex_data" ]]; then
                    while IFS= read -r hex_line || [[ -n "$hex_line" ]]; do
                        [[ -z "$hex_line" ]] && continue

                        mount_info=$(parse_mount "$hex_line")
                        if [[ -n "$mount_info" ]]; then
                            mount_path="${mount_info%%|*}"
                            remaining="${mount_info#*|}"
                            mount_is_enabled="${remaining%%|*}"
                            remaining="${remaining#*|}"
                            mount_env_name="${remaining%%|*}"

                            # Normalize DB mount path for comparison
                            normalized_db_path=$(normalize_path "$mount_path")

                            if [[ "$normalized_db_path" == "$resolved_path" ]]; then
                                found_in_db=true
                                is_enabled="$mount_is_enabled"
                                environment_name="$mount_env_name"
                                break
                            fi
                        fi
                    done <<< "$mount_hex_data"
                fi

                # If found in DB and disabled, report as disabled (consistent with default mode)
                if [[ "$found_in_db" == "true" ]] && [[ "$is_enabled" == "false" ]]; then
                    log "Required local .env file is disabled: \"${resolved_path}\""
                    disabled_mounts+=("$resolved_path|$environment_name")
                    continue
                fi

                # Check if path exists and is a FIFO
                if [[ ! -e "$resolved_path" ]] || [[ ! -p "$resolved_path" ]]; then
                    if [[ "$found_in_db" == "true" ]]; then
                        # File is enabled in DB but missing/invalid
                        log "Required local .env file is missing or invalid: \"${resolved_path}\""
                        invalid_mounts+=("$resolved_path|$environment_name")
                    else
                        # Not found in DB, but required by TOML
                        log "Required local .env file is missing or invalid: \"${resolved_path}\""
                        required_mounts+=("$resolved_path")
                    fi
                else
                    # File exists and is a FIFO
                    if [[ "$found_in_db" == "true" ]]; then
                        log "Required local .env file is valid and enabled: \"${resolved_path}\""
                    else
                        # File exists but not found in DB
                        # The file might have been created manually or the DB query failed
                        log "Required local .env file exists but not found in 1Password database: \"${resolved_path}\""
                    fi
                fi
            done
        fi
    else
        # Default mode: Check all local .env files within this workspace from 1Password database
        log "Using default mode: checking all local .env files in workspace ${workspace_root} from 1Password database"

        if [[ -z "$mount_hex_data" ]]; then
            log "No mount data available from 1Password database, skipping workspace ${workspace_root}"
            continue
        fi

        log "Environment mount data found, checking relevant local .env files for workspace ${workspace_root}..."

        # Process each mount entry from database
        while IFS= read -r hex_line || [[ -n "$hex_line" ]]; do
            [[ -z "$hex_line" ]] && continue

            mount_info=$(parse_mount "$hex_line")
            if [[ -n "$mount_info" ]]; then
                # Parse mount_info: mount_path|is_enabled|environment_name|uuid|environment_uuid
                mount_path="${mount_info%%|*}"
                remaining="${mount_info#*|}"
                is_enabled="${remaining%%|*}"
                remaining="${remaining#*|}"
                environment_name="${remaining%%|*}"
                remaining="${remaining#*|}"
                uuid="${remaining%%|*}"
                environment_uuid="${remaining#*|}"

                log "Checking local .env file with id ${uuid} at path \"${mount_path}\" for environment ${environment_uuid} (${environment_name})"

                # Check if this local .env file is relevant to the current workspace
                if ! is_project_mount "$mount_path" "$workspace_root"; then
                    log "Local .env file does not belong to workspace ${workspace_root}, skipping"
                    continue
                fi

                if [[ "$is_enabled" == "true" ]]; then
                    if [[ ! -e "$mount_path" ]] || [[ ! -p "$mount_path" ]]; then
                        log "Local .env file is invalid (file is not present or not a FIFO)"
                        invalid_mounts+=("$mount_path|$environment_name")
                    else
                        log "Local .env file is valid and enabled"
                    fi
                else
                    log "Local .env file is disabled"
                    disabled_mounts+=("$mount_path|$environment_name")
                fi
            fi
        done <<< "$mount_hex_data"
    fi
done

# ============================================================================
# PERMISSION DECISION LOGIC
# ============================================================================

# Consolidate all missing/invalid mounts (from DB and TOML)
all_missing_invalid=()
if [[ ${#invalid_mounts[@]} -gt 0 ]]; then
    for mount_entry in "${invalid_mounts[@]}"; do
        all_missing_invalid+=("${mount_entry%%|*}")
    done
fi
if [[ ${#required_mounts[@]} -gt 0 ]]; then
    for mount_path in "${required_mounts[@]}"; do
        # Avoid duplicates
        is_duplicate=false
        if [[ ${#all_missing_invalid[@]} -gt 0 ]]; then
            for existing_path in "${all_missing_invalid[@]}"; do
                if [[ "$existing_path" == "$mount_path" ]]; then
                    is_duplicate=true
                    break
                fi
            done
        fi
        if [[ "$is_duplicate" == "false" ]]; then
            all_missing_invalid+=("$mount_path")
        fi
    done
fi

# Generate unified error messages
if [[ ${#all_missing_invalid[@]} -gt 0 ]] || [[ ${#disabled_mounts[@]} -gt 0 ]]; then
    permission="deny"

    # Build message for missing/invalid mounts
    if [[ ${#all_missing_invalid[@]} -gt 0 ]]; then
        log "Denying permission due to missing or invalid environment files"

        # Extract environment name from DB mounts if available
        environment_name=""
        if [[ ${#invalid_mounts[@]} -gt 0 ]]; then
            first_invalid="${invalid_mounts[0]}"
            environment_name="${first_invalid#*|}"
        fi

        if [[ ${#all_missing_invalid[@]} -eq 1 ]]; then
            if [[ -n "$environment_name" ]]; then
                agent_message="This project uses 1Password environments. An environment file is expected to be mounted at the specified path. Error: the file is missing or invalid. Environment name: \"${environment_name}\". Path: \"${all_missing_invalid[0]}\". Suggestion: ensure the local .env file is configured and enabled from the environment's destinations tab in the 1Password app."
            else
                agent_message="This project uses 1Password environments. An environment file is required by environments.toml. Error: the file is missing or invalid. Path: \"${all_missing_invalid[0]}\". Suggestion: ensure the local .env file is configured and enabled from the environment's destinations tab in the 1Password app."
            fi
        else
            file_list=$(IFS=','; echo "${all_missing_invalid[*]}" | sed 's/,/, /g')
            if [[ -n "$environment_name" ]]; then
                agent_message="This project uses 1Password environments. Environment files are expected to be mounted at the specified paths. Error: these files are missing or invalid. Environment name: \"${environment_name}\". Paths: \"${file_list}\". Suggestion: ensure the local .env files are configured and enabled from the environment's destinations tab in the 1Password app."
            else
                agent_message="This project uses 1Password environments. Environment files are required by environments.toml. Error: these files are missing or invalid. Paths: \"${file_list}\". Suggestion: ensure the local .env files are configured and enabled from the environment's destinations tab in the 1Password app."
            fi
        fi
    fi

    # Handle disabled mounts (different issue - needs to be enabled, not configured)
    if [[ ${#disabled_mounts[@]} -gt 0 ]]; then
        log "Denying permission due to disabled local .env files"

        # Extract environment name
        first_disabled="${disabled_mounts[0]}"
        environment_name="${first_disabled#*|}"

        # Extract mount paths
        mount_paths=()
        for mount_entry in "${disabled_mounts[@]}"; do
            mount_paths+=("${mount_entry%%|*}")
        done

        if [[ ${#disabled_mounts[@]} -eq 1 ]]; then
            disabled_msg="Error: the file is not mounted. Environment name: \"${environment_name}\". Path: \"${mount_paths[0]}\". Suggestion: enable the local .env file from the environment's destinations tab in the 1Password app."
        else
            file_list=$(IFS=','; echo "${mount_paths[*]}" | sed 's/,/, /g')
            disabled_msg="Error: these files are not mounted. Environment name: \"${environment_name}\". Paths: \"${file_list}\". Suggestion: enable the local .env files from the environment's destinations tab in the 1Password app."
        fi

        # Combine messages if we have both missing/invalid and disabled
        if [[ ${#all_missing_invalid[@]} -gt 0 ]]; then
            agent_message="${agent_message} ${disabled_msg}"
        else
            if [[ ${#disabled_mounts[@]} -eq 1 ]]; then
                agent_message="This project uses 1Password environments. An environment file is expected to be mounted at the specified path. ${disabled_msg}"
            else
                agent_message="This project uses 1Password environments. Environment files are expected to be mounted at the specified paths. ${disabled_msg}"
            fi
        fi
    fi
fi

# Output JSON response with permission decision
output_response
exit 0

