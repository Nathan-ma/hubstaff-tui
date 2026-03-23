package main

import (
	"fmt"
	"os"
)

// runCompletion handles the 'completion' subcommand.
// Returns the exit code.
func runCompletion(args []string) int {
	if len(args) == 0 {
		printCompletionHelp()
		return 0
	}
	shell := args[0]
	switch shell {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	case "--help", "-h":
		printCompletionHelp()
	default:
		fmt.Fprintf(os.Stderr, "unknown shell %q: use bash, zsh, or fish\n", shell)
		return 1
	}
	return 0
}

func printCompletionHelp() {
	fmt.Print(`Generate shell completion scripts for hubstaff-tui.

Usage:
  hubstaff-tui completion bash   Output bash completion script
  hubstaff-tui completion zsh    Output zsh completion script
  hubstaff-tui completion fish   Output fish completion script

Installation:
  Bash (add to ~/.bashrc):
    source <(hubstaff-tui completion bash)

  Zsh (add to ~/.zshrc):
    source <(hubstaff-tui completion zsh)

  Fish (save to completions directory):
    hubstaff-tui completion fish > ~/.config/fish/completions/hubstaff-tui.fish
`)
}

const bashCompletion = `# bash completion for hubstaff-tui
_hubstaff_tui_completions() {
    local cur prev words cword
    _init_completion || return

    local subcommands="status setup doctor completion"
    local global_flags="--help --version --config"

    if [[ $cword -eq 1 ]]; then
        COMPREPLY=($(compgen -W "$subcommands $global_flags" -- "$cur"))
        return
    fi

    case "${words[1]}" in
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "$cur"))
            ;;
        --config)
            COMPREPLY=($(compgen -f -- "$cur"))
            ;;
    esac
}

complete -F _hubstaff_tui_completions hubstaff-tui
`

const zshCompletion = `#compdef hubstaff-tui

_hubstaff_tui() {
    local -a subcommands
    subcommands=(
        'status:Print current tracking status for tmux status-right'
        'setup:Configure tmux keybinding and status bar'
        'doctor:Run setup diagnostics'
        'completion:Generate shell completion scripts'
    )

    local -a global_flags
    global_flags=(
        '--help:Show help'
        '--version:Show version'
        '--config:Use a custom config file'
    )

    if (( CURRENT == 2 )); then
        _describe 'subcommands' subcommands
        _describe 'flags' global_flags
        return
    fi

    case "${words[2]}" in
        completion)
            local -a shells
            shells=('bash' 'zsh' 'fish')
            _describe 'shell' shells
            ;;
        --config)
            _files
            ;;
    esac
}

_hubstaff_tui "$@"
`

const fishCompletion = `# fish completion for hubstaff-tui
complete -c hubstaff-tui -f

# Subcommands
complete -c hubstaff-tui -n "__fish_use_subcommand" -a status -d "Print current tracking status for tmux status-right"
complete -c hubstaff-tui -n "__fish_use_subcommand" -a setup -d "Configure tmux keybinding and status bar"
complete -c hubstaff-tui -n "__fish_use_subcommand" -a doctor -d "Run setup diagnostics"
complete -c hubstaff-tui -n "__fish_use_subcommand" -a completion -d "Generate shell completion scripts"

# Global flags
complete -c hubstaff-tui -l help -d "Show help"
complete -c hubstaff-tui -l version -d "Show version"
complete -c hubstaff-tui -l config -d "Use a custom config file" -r

# completion subcommand shells
complete -c hubstaff-tui -n "__fish_seen_subcommand_from completion" -a bash -d "Bash completion script"
complete -c hubstaff-tui -n "__fish_seen_subcommand_from completion" -a zsh -d "Zsh completion script"
complete -c hubstaff-tui -n "__fish_seen_subcommand_from completion" -a fish -d "Fish completion script"
`
