#!/usr/bin/env bash
#
# hubstaff-fzf.sh - Interactive Hubstaff time tracker with fzf
#
# Optimized for speed - starts with projects view (single API call)
# Tasks are loaded on-demand for selected project only
#
# Keybindings:
#   enter    - Select project / Start tracking task
#   ctrl-e   - Stop current tracking
#   ctrl-r   - Refresh list
#   ctrl-t   - Show tasks for active project
#   esc      - Exit / Go back

set -euo pipefail

# Hubstaff CLI path
HUBSTAFF_CLI="/Applications/Hubstaff.app/Contents/MacOS/HubstaffCLI"

# Cache settings
CACHE_DIR="${TMPDIR:-/tmp}/hubstaff-cache"
CACHE_TTL=300  # 5 minutes

# Ensure cache directory exists
mkdir -p "$CACHE_DIR"

# Check dependencies
if [[ ! -x "$HUBSTAFF_CLI" ]]; then
    echo "Error: Hubstaff CLI not found at $HUBSTAFF_CLI" >&2
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed" >&2
    exit 1
fi

# ─────────────────────────────────────────────────────────────
# API Functions with caching
# ─────────────────────────────────────────────────────────────

get_status() {
    $HUBSTAFF_CLI status 2>/dev/null
}

get_projects_cached() {
    local cache_file="$CACHE_DIR/projects.json"
    local now=$(date +%s)

    if [[ -f "$cache_file" ]]; then
        local cache_time=$(stat -f %m "$cache_file" 2>/dev/null || echo 0)
        if (( now - cache_time < CACHE_TTL )); then
            cat "$cache_file"
            return
        fi
    fi

    local data=$($HUBSTAFF_CLI projects 2>/dev/null)
    echo "$data" > "$cache_file"
    echo "$data"
}

get_tasks_cached() {
    local project_id="$1"
    local cache_file="$CACHE_DIR/tasks_${project_id}.json"
    local now=$(date +%s)

    if [[ -f "$cache_file" ]]; then
        local cache_time=$(stat -f %m "$cache_file" 2>/dev/null || echo 0)
        if (( now - cache_time < CACHE_TTL )); then
            cat "$cache_file"
            return
        fi
    fi

    local data=$($HUBSTAFF_CLI tasks "$project_id" 2>/dev/null)
    echo "$data" > "$cache_file"
    echo "$data"
}

clear_cache() {
    rm -f "$CACHE_DIR"/*.json 2>/dev/null
}

# ─────────────────────────────────────────────────────────────
# Display Functions
# ─────────────────────────────────────────────────────────────

build_header() {
    local status_json
    status_json=$(get_status)

    local tracking project_name task_name tracked_today
    tracking=$(echo "$status_json" | jq -r '.tracking // false')
    project_name=$(echo "$status_json" | jq -r '.active_project.name // "None"')
    task_name=$(echo "$status_json" | jq -r '.active_task.name // "None"')
    tracked_today=$(echo "$status_json" | jq -r '.active_project.tracked_today // "0:00:00"')

    local status_icon
    if [[ "$tracking" == "true" ]]; then
        status_icon="●"
    else
        status_icon="○"
    fi

    echo "$status_icon Tracking: $task_name"
    echo "⏱  Today: $tracked_today | 📂 $project_name"
    echo "─────────────────────────────────────────────────"
    echo "enter: select | ctrl-e: stop | ctrl-t: tasks | esc: exit"
}

list_projects() {
    local projects_json
    projects_json=$(get_projects_cached)

    local status_json
    status_json=$(get_status)
    local active_project_id
    active_project_id=$(echo "$status_json" | jq -r '.active_project.id // 0')
    local tracking
    tracking=$(echo "$status_json" | jq -r '.tracking // false')

    echo "$projects_json" | jq -r '.projects[] | "\(.id)|\(.name)"' | while IFS='|' read -r project_id project_name; do
        local indicator=" "
        if [[ "$project_id" == "$active_project_id" && "$tracking" == "true" ]]; then
            indicator="●"
        elif [[ "$project_id" == "$active_project_id" ]]; then
            indicator="○"
        fi
        printf "%s\t%s\t%s\n" "$indicator" "$project_id" "$project_name"
    done
}

list_tasks_for_project() {
    local project_id="$1"
    local project_name="$2"

    local tasks_json
    tasks_json=$(get_tasks_cached "$project_id")

    local status_json
    status_json=$(get_status)
    local active_task_id
    active_task_id=$(echo "$status_json" | jq -r '.active_task.id // 0')
    local tracking
    tracking=$(echo "$status_json" | jq -r '.tracking // false')

    if [[ -n "$tasks_json" && "$tasks_json" != "null" ]]; then
        echo "$tasks_json" | jq -r '.tasks[]? | "\(.id)|\(.summary)"' 2>/dev/null | while IFS='|' read -r task_id task_name; do
            local indicator=" "
            if [[ "$task_id" == "$active_task_id" && "$tracking" == "true" ]]; then
                indicator="●"
            elif [[ "$task_id" == "$active_task_id" ]]; then
                indicator="○"
            fi
            printf "%s\t%s\t%s\t%s\t%s\n" "$indicator" "$task_id" "$task_name" "$project_id" "$project_name"
        done
    fi
}

list_active_project_tasks() {
    local status_json
    status_json=$(get_status)
    local active_project_id
    active_project_id=$(echo "$status_json" | jq -r '.active_project.id // 0')
    local active_project_name
    active_project_name=$(echo "$status_json" | jq -r '.active_project.name // "Unknown"')

    if [[ "$active_project_id" != "0" && "$active_project_id" != "null" ]]; then
        list_tasks_for_project "$active_project_id" "$active_project_name"
    else
        echo "No active project"
    fi
}

# ─────────────────────────────────────────────────────────────
# FZF Theme
# ─────────────────────────────────────────────────────────────

FZF_THEME=(
    --color='bg+:#313244,bg:#1e1e2e,spinner:#f5e0dc,hl:#f38ba8'
    --color='fg:#cdd6f4,header:#f38ba8,info:#cba6f7,pointer:#f5e0dc'
    --color='marker:#f5e0dc,fg+:#cdd6f4,prompt:#cba6f7,hl+:#f38ba8'
    --color='border:#89b4fa,label:#89b4fa'
    --border='rounded'
    --border-label=' 󰔛 Hubstaff '
    --border-label-pos=3
    --pointer='󰄾'
    --marker='󰄲'
)

# ─────────────────────────────────────────────────────────────
# Main Interface
# ─────────────────────────────────────────────────────────────

run_projects_fzf() {
    local header
    header=$(build_header)

    local selected
    selected=$(list_projects | fzf \
        --ansi \
        --header="$header" \
        --delimiter='\t' \
        --with-nth='1,3' \
        --preview='echo -e "\033[1;34m📂 Project\033[0m\n{3}\n\n\033[2mProject ID: {2}\033[0m"' \
        --preview-window='right:40%:wrap:border-left' \
        --bind="ctrl-e:execute-silent($HUBSTAFF_CLI stop 2>/dev/null)+reload($0 --list-projects)" \
        --bind="ctrl-r:reload($0 --list-projects --clear-cache)" \
        --bind="ctrl-t:become($0 --tasks)" \
        --prompt=' Projects  ' \
        --expect='enter' \
        "${FZF_THEME[@]}")

    if [[ -n "$selected" ]]; then
        local key=$(echo "$selected" | head -1)
        local line=$(echo "$selected" | tail -1)

        if [[ "$key" == "enter" && -n "$line" ]]; then
            local project_id=$(echo "$line" | awk -F'\t' '{print $2}')
            local project_name=$(echo "$line" | awk -F'\t' '{print $3}')

            # Show tasks for selected project
            run_tasks_fzf "$project_id" "$project_name"
        fi
    fi
}

run_tasks_fzf() {
    local project_id="$1"
    local project_name="$2"

    local header
    header=$(build_header)

    list_tasks_for_project "$project_id" "$project_name" | fzf \
        --ansi \
        --header="$header" \
        --header-label="📂 $project_name" \
        --delimiter='\t' \
        --with-nth='1,3' \
        --preview='echo -e "\033[1;33m📋 Task\033[0m\n{3}\n\n\033[2mTask ID: {2}\nProject: {5}\033[0m"' \
        --preview-window='right:40%:wrap:border-left' \
        --bind="enter:execute-silent($HUBSTAFF_CLI start_task {2} 2>/dev/null)+reload($0 --list-tasks $project_id '$project_name')" \
        --bind="ctrl-e:execute-silent($HUBSTAFF_CLI stop 2>/dev/null)+reload($0 --list-tasks $project_id '$project_name')" \
        --bind="ctrl-r:reload($0 --list-tasks $project_id '$project_name' --clear-cache)" \
        --bind='ctrl-p:become('"$0"')' \
        --prompt=' Tasks  ' \
        "${FZF_THEME[@]}"
}

run_active_tasks_fzf() {
    local status_json
    status_json=$(get_status)
    local project_id
    project_id=$(echo "$status_json" | jq -r '.active_project.id // 0')
    local project_name
    project_name=$(echo "$status_json" | jq -r '.active_project.name // "Unknown"')

    if [[ "$project_id" != "0" && "$project_id" != "null" ]]; then
        run_tasks_fzf "$project_id" "$project_name"
    else
        echo "No active project. Starting with projects view..."
        run_projects_fzf
    fi
}

# ─────────────────────────────────────────────────────────────
# CLI Handler
# ─────────────────────────────────────────────────────────────

# Handle --clear-cache flag
if [[ "${*}" == *"--clear-cache"* ]]; then
    clear_cache
fi

case "${1:-}" in
    --list-projects)
        list_projects
        ;;
    --list-tasks)
        list_tasks_for_project "$2" "$3"
        ;;
    --tasks)
        run_active_tasks_fzf
        ;;
    --status)
        get_status | jq .
        ;;
    --stop)
        $HUBSTAFF_CLI stop 2>/dev/null
        ;;
    --clear-cache)
        clear_cache
        echo "Cache cleared"
        ;;
    *)
        run_projects_fzf
        ;;
esac
