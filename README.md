<div align="center">
  <img alt="decoreba" src="cmd/decoreba-desktop/appicon-transparent.png" width="240">

  # decoreba

  A manual command vault for the terminal commands you know you will need again.

  [![npm](https://img.shields.io/npm/v/decoreba?style=flat-square&logo=npm&label=npm)](https://www.npmjs.com/package/decoreba)
  [![CI](https://img.shields.io/github/actions/workflow/status/matheuzgomes/decoreba/build-and-release.yaml?style=flat-square&label=CI)](https://github.com/matheuzgomes/decoreba/actions)
  [![MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
</div>

At work, I kept needing commands with very specific parameters and forgetting to save them anywhere. When I started using tmux, I kept searching Google for commands.

Decoreba keeps those commands in a searchable vault. You add an entry when it is worth remembering, then find it from the terminal when you need it again.

I think of it as a permanent Windows + V for commands. You decide what goes into it, so the vault stays useful instead of collecting every command you ever typed. Shell history and [fzf](https://github.com/junegunn/fzf) still work well for commands you did not save. Decoreba adds another place to look, with saved context, notes, and workflows.

<div align="center">
  <img alt="decoreba search demo" src="assets/gifs/decoreba-search.gif" width="800">
</div>

## Quick start

Install with npm:

```bash
npm install -g decoreba
decoreba version
```

This npm package provides binaries for Linux, macOS, and Windows. You can also install decoreba with Homebrew or build it from source:

```bash
# Homebrew
brew tap matheuzgomes/decoreba
brew install decoreba

# Go 1.25+
go install github.com/matheuzgomes/decoreba/cmd/decoreba@latest
```

Save a command you want to remember:

```bash
decoreba add
```

The form asks for a context, title, command, tags, and notes. For example:

```text
Context: tmux
Title: Rename the current window
Command: tmux rename-window {{name}}
```

Search for it later:

```bash
decoreba tmux rename
```

Select the result and press `Enter` to copy it. Press `Ctrl+X` to execute it after confirmation.

## How it works

Decoreba follows a small, deliberate loop:

1. You choose a command worth saving.
2. Decoreba stores the command and its context in one JSON file.
3. You search by context, title, command, or tags.
4. The inline palette copies the selected command or runs it after confirmation.

The palette opens below your prompt and disappears when you finish. It does not replace your terminal with a full-screen interface or require a background service.

Commands can include placeholders for values that change each time:

```text
docker logs --tail {{lines:100}} {{container}}
```

Decoreba asks for those values when you copy or execute the command.

A saved command can also contain several titled steps. Run one step at a time with `Enter`, or run the remaining steps with `Ctrl+X` after confirmation. Failed steps stay visible in the workflow.

## Use it from your shell

Generate completions for Bash or Zsh:

```bash
source <(decoreba completion bash)
source <(decoreba completion zsh)
```

For Fish:

```fish
decoreba completion fish > ~/.config/fish/completions/decoreba.fish
```

Install the Bash or Zsh widget and completions together:

```bash
decoreba init
decoreba init --yes
```

After loading the shell integration, press `Ctrl+O` while typing a command. Decoreba uses the text already on your command line as the search query. `Enter` puts the selected command back at the prompt. Fish supports completions, but not this widget.

You can load the widget directly when you want to inspect the generated script:

```bash
source <(decoreba shell bash)
source <(decoreba shell zsh)
```

## Other commands

```bash
decoreba                    # search all contexts
decoreba git undo            # search within git
decoreba list                # list contexts and command counts
decoreba list docker         # list commands in docker
decoreba edit <id>           # edit by id or id prefix
decoreba rm <id>             # remove by id or id prefix
decoreba stats               # show vault statistics
decoreba export              # export commands to stdout
decoreba import [file]       # import from stdin or a file
decoreba help                # show the full command list
```

## Why use decoreba?

Use it when you remember that a command exists, but not the exact flags, syntax, or parameters. The saved context and notes help you recognize the right command. Workflows let you keep related commands together when one operation needs several steps.

Use shell history or `fzf` when you want to search everything you already typed. Use decoreba when you want to keep a smaller collection of commands on purpose.

## A deliberate limitation

Decoreba does not save commands automatically. You add entries manually because the useful collection is the one you chose, and because automatic capture raises privacy and security questions. This also means commands you never save will not appear in the vault.

## Optional sync

The vault is one JSON file. You can move it between machines yourself, or sync it through a private GitHub Gist:

```bash
decoreba sync init
decoreba sync push
decoreba sync pull
decoreba sync status
```

Sync needs `DECOREBA_GIST_TOKEN`, a classic GitHub token with the `gist` scope. Use `--encrypt` to encrypt the uploaded vault with AES-256-GCM.

Sync is an optional transport. It is not required for the local workflow.

## Optional MCP server

The MCP server exposes the vault to AI agents over stdin and stdout. You can search, read, add, edit, remove, and execute commands through it.

```bash
decoreba mcp
```

Write and delete operations require `confirm: true`. Dangerous command patterns are blocked by default, and modifications are backed up first. You can ignore this section if you do not use MCP.

## Data and configuration

Commands are stored in one `commands.json` file:

| System | Path |
|---|---|
| Linux | `$XDG_CONFIG_HOME/decoreba/commands.json` |
| macOS | `~/Library/Application Support/decoreba/commands.json` |
| Windows | `%AppData%\decoreba\commands.json` |

Set `DECOREBA_CONFIG` to override the configuration directory. `NO_COLOR` and `--no-color` disable ANSI colors.

## Development

Go 1.25 or newer and `make` are required for the source workflow.

```bash
git clone https://github.com/matheuzgomes/decoreba
cd decoreba
go test ./...
make build
```

## Help and links

- [Issues](https://github.com/matheuzgomes/decoreba/issues)
- [Releases](https://github.com/matheuzgomes/decoreba/releases)
- [MIT license](LICENSE)
