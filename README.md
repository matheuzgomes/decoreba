<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="cmd/decoreba-desktop/appicon-transparent.png">
    <img alt="decoreba" src="cmd/decoreba-desktop/appicon-transparent.png" width="320">
  </picture>

  # decoreba

  A command vault for the commands you keep looking up.

  [![npm](https://img.shields.io/npm/v/decoreba?style=flat-square&logo=npm&label=npm)](https://www.npmjs.com/package/decoreba)
  [![CI](https://img.shields.io/github/actions/workflow/status/matheuzgomes/decoreba/build-and-release.yaml?style=flat-square&label=CI)](https://github.com/matheuzgomes/decoreba/actions)
  [![MIT](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)

  [Install](#install) · [Quick start](#quick-start) · [Usage](#usage) · [Shell integration](#shell-integration)
</div>

<div align="center">
  <img alt="decoreba search demo" src="assets/gifs/decoreba-search.gif" width="800">
</div>

You know the command. You have used it before. You just cannot remember the
flag that makes it work.

`decoreba` keeps those commands in a small, searchable vault. Save a command
with a title and some context, then find it from the terminal before you have
to open a browser, scroll through shell history, or search your notes.

It is deliberately a vault, not a history searcher: you save commands first.
If you want fuzzy search over everything you have already typed, [fzf](https://github.com/junegunn/fzf)
is the better tool. I use both.

I made decoreba because it is something I needed and wanted to build. It is
free, and it will stay free. If you have an idea, a rough edge, or something
that would make it more useful, open a PR or an issue.

## Install

The simplest path is npm. It downloads a prebuilt binary for your platform:

```bash
npm install -g decoreba
```

Or install from source with Go:

```bash
go install github.com/matheuzgomes/decoreba/cmd/decoreba@latest
```

Supported release targets are Linux (amd64, arm64), macOS (amd64, arm64), and
Windows (amd64).

## Quick start

Add something you keep forgetting:

```bash
decoreba add
```

Then search for it:

```bash
decoreba                    # search all contexts
decoreba docker              # search the docker context
decoreba git undo            # search for "undo" in git
```

The palette opens below your prompt. Type to filter, press `Enter` to copy the
selected command, or press `Ctrl+X` to execute it after confirmation.

Once shell integration is enabled, `Ctrl+O` opens this same palette from your
existing command line. Whatever you already typed becomes the search query;
`Enter` puts the selected command back at the prompt.

## Why decoreba?

Shell history answers: “What did I type?”

Decoreba answers: “What is the command I keep needing?”

Each entry has a context, title, command, tags, notes, and usage history. The
context is usually the tool you are working with—`git`, `docker`, `kubectl`—and
is detected from the directory when possible. Search covers the title,
command, context, and tags, including accents and common typos.

The palette stays inline instead of taking over the terminal. It appears,
helps, and goes away. No alternate screen to restore and no background daemon
to keep running.

## Usage

### Add a command

```bash
decoreba add
```

The form asks for a context, title, command, tags, and notes. Commands can use
placeholders when part of the command changes each time:

```text
docker logs --tail {{lines:100}} {{container}}
```

When you copy or execute it, decoreba asks for the values inline.

<img alt="decoreba add command form" src="assets/gifs/decoreba-add.gif" width="600">

### Run a workflow

Turn a command into a sequence of titled steps with `Ctrl+W` in the form.
Run one step at a time with `Enter`, or run the remaining steps with `Ctrl+X`.
Failed steps are shown in place, and `Esc` aborts the workflow.

<img alt="decoreba workflow runner" src="assets/gifs/decoreba-workflow.gif" width="700">

### Keybindings

| Key | Action |
|---|---|
| `↑` / `Ctrl+K` | Move up |
| `↓` / `Ctrl+J` | Move down |
| `Enter` | Copy the selected command |
| `1`–`9` | Select directly when the search is empty |
| `Ctrl+X` | Execute after confirmation |
| `Ctrl+E` | Edit the selected command |
| `Ctrl+S` | Pin the selected command |
| `Esc` / `Ctrl+C` | Cancel |

`Shift+Enter` executes on terminals that support the [kitty keyboard
protocol](https://sw.kovidgoyal.net/kitty/keyboard-protocol/), including kitty,
WezTerm, and Ghostty. In other terminals it behaves like `Enter`.

### Shell integration

Generate completions for bash, zsh, or fish:

```bash
eval "$(decoreba completion bash)"
```

You can also install the widget and completions together:

```bash
decoreba init          # interactive
decoreba init --yes    # non-interactive
```

The widget opens the palette with the text already on your command line:

```bash
source <(decoreba shell bash)   # or zsh
```

Press `Ctrl+O` after typing part of a command. The text becomes the search
query, so you can go from `docker p` to a saved command without starting over.


## Other commands

```bash
decoreba list                 # list contexts and command counts
decoreba list docker          # list commands in a context
decoreba edit <id>            # edit by id or id prefix
decoreba rm <id>              # remove by id or id prefix
decoreba stats                # show vault statistics
decoreba export               # export commands to stdout
decoreba import [file]         # import from stdin or a file
decoreba help                 # show the full command list
```

## Sync across machines

The vault is one JSON file, but if you want decoreba to move changes through a
private GitHub Gist:

```bash
decoreba sync init
decoreba sync push
decoreba sync pull
decoreba sync status
```

Sync needs `DECOREBA_GIST_TOKEN`, a classic GitHub token with the `gist` scope.
Use `--encrypt` to encrypt the uploaded vault with AES-256-GCM.

If copying a config file between machines is all you need, do that instead.
Sync is an optional transport, not part of the core workflow.

## MCP server (optional)

You do not need MCP to use decoreba. It is an optional way to expose the vault
to AI agents over stdin/stdout using the [Model Context
Protocol](https://modelcontextprotocol.io). I added it because it might be
useful to people who already use agents with their terminal—not because
commands need an AI layer.

If you want it:

```bash
decoreba mcp
```

It supports searching, reading, adding, editing, removing, and executing
commands. Write and delete operations require `confirm: true`; dangerous
commands are blocked, and modifications are backed up first. If you do not
use MCP, you can ignore this section.

## Data

Commands are stored in one `commands.json` file:

| OS | Path |
|---|---|
| Linux | `$XDG_CONFIG_HOME/decoreba/commands.json` |
| macOS | `~/Library/Application Support/decoreba/commands.json` |
| Windows | `%AppData%\decoreba\commands.json` |

Set `DECOREBA_CONFIG` to override the directory. `NO_COLOR` and
`--no-color` disable ANSI colors.

## Development

```bash
git clone https://github.com/matheuzgomes/decoreba
cd decoreba
go test ./...
make build
```

## License

[MIT](LICENSE)
