<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="cmd/decoreba-desktop/appicon.png">
    <img alt="decoreba" src="cmd/decoreba-desktop/appicon.png" width="120">
  </picture>

  <h1>decoreba</h1>

  <p>Personal command vault — inline terminal palette. Zero dependencies.
  ~2 ms startup. 3.7 MB binary.</p>

  <p>
    <a href="https://www.npmjs.com/package/decoreba">
      <img src="https://img.shields.io/npm/v/decoreba?label=npm" alt="npm">
    </a>
    <a href="https://github.com/matheuzgomes/decoreba">
      <img src="https://img.shields.io/github/go-mod/go-version/matheuzgomes/decoreba" alt="Go version">
    </a>
    <a href="https://github.com/matheuzgomes/decoreba/actions">
      <img src="https://img.shields.io/github/actions/workflow/status/matheuzgomes/decoreba/build-and-release.yaml?label=CI" alt="CI">
    </a>
    <a href="LICENSE">
      <img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License">
    </a>
  </p>
</div>

---

```text
$ decoreba git

  git › undo
  1  Undo last commit keeping changes     ◂ git reset --soft HEAD~1
  2  Temporarily ignore tracked changes   ◂ git update-index --assume-unchanged
  3  Apply and drop most recent stash     ◂ git stash pop
  4  Show all stashes                     ◂ git stash list
↵ copy   ↑↓ nav   1-9 direct   ^e edit   ^x exec   ^s pin   esc cancel
```

---

## Install

```bash
npm install -g decoreba
```

Downloads a prebuilt static binary for your platform. No Go toolchain needed.

```bash
go install github.com/matheuzgomes/decoreba/cmd/decoreba@latest
```

## Use

```bash
decoreba add                # add a command
decoreba git                # search within "git" context
decoreba git undo           # search with query pre-filled
decoreba                    # auto-detect context from cwd
decoreba ls                 # list contexts and counts
decoreba ls docker          # list commands in "docker"
decoreba rm 0c9             # remove by id prefix
decoreba edit 0c9           # edit by id prefix
decoreba stats              # vault statistics
decoreba export             # dump all commands as JSON
decoreba import < file.json # merge from JSON
decoreba --shell-output     # print result to stdout (pipeable)
```

---

## Inline palette

The overlay appears below your cursor line, not as a fullscreen app. Pure ANSI
escape codes — no libraries, no alternate screen. Dismiss it and the terminal
is exactly as it was before.

```text
$ decoreba docker ps

  docker › ps
  1  List all containers including stopped   ◂ docker ps -a
  2  Clean up unused resources               ◂ docker system prune -af
  3  Show resource usage per container       ◂ docker stats --no-stream
```

Press `Enter` to copy. `Shift+Enter` to execute.

## Fuzzy search + typo tolerance

```text
$ decoreba dackar ps

  docker › ps
  1  List all containers including stopped   ◂ docker ps -a
```

Damerau–Levenshtein distance ≤2 catches typos at the bottom of results when
exact fuzzy match returns nothing. Accents are normalized — `próximo` matches
`proximo` and vice versa.

## Recency ranking

Score = fuzzy match + usage count + exponential time decay (half-life 48 h).
Commands you reach for often stay near the top.

## Add / Edit form

```text
$ decoreba add
╭──────────────────────────────────────────────────────────────────────────────╮
│  ▶ add command                                                                │
│                                                                                │
│  Context    docker                                                             │
│  Title      Prune everything                                                   │
│  Command    docker system prune -af --volumes                                  │
│  Tags       cleanup, disk                                                      │
│  Notes      Careful with --volumes                                             │
│                                                                                │
│  tab next   ⇧tab prev   ^s save   ^w workflow   esc cancel                    │
╰──────────────────────────────────────────────────────────────────────────────╯
```

Five fields: context, title, command, tags, notes. Context autocompletes from
existing entries. Tags render as colored chips. `Ctrl+W` turns command into
a multi-step workflow editor.

`decoreba edit <id>` or `Ctrl+E` from the palette reopens the form with data
pre-filled.

## Workflows

Commands can have multiple steps. Each step is a title + command pair.

```text
$ decoreba deploy

╭──────────────────────────────────────────────────────────────────────────────╮
│  ▶ Running workflow (step 2/3)                                                │
│                                                                                │
│  → 1  Build image          docker build -t web .                              │
│  ✓ 2  Start container      docker run -d -p 80:80 web                         │
│  → 3  Clean up             docker system prune -f                             │
│                                                                                │
│  enter next   ^x run all   esc abort                                          │
╰──────────────────────────────────────────────────────────────────────────────╯
```

`Enter` steps forward. `Ctrl+X` runs all remaining. `Esc` aborts.

## Variables

Add `{{placeholder}}` or `{{name:default}}` to any command. On copy or execute,
decoreba prompts you for each value inline:

```text
$ decoreba deploy

  Container name: [web]
```

## Execute mode

`Shift+Enter` runs the selected command directly. With `--shell-output`,
stdout replaces the clipboard, making the result pipeable:

```bash
decoreba docker ps --shell-output | grep "Up"
```

## Completions

```bash
eval "$(decoreba completion bash)"   # or zsh / fish
```

## Desktop GUI (in development)

Wails-based system tray overlay with a global hotkey (`Alt+Shift+Space`).
Not included in npm or `go install` packages until ready. Build from source
with `wails build` in `cmd/decoreba-desktop/`.

---

## Keybindings

### Palette

| Key | Action |
|---|---|
| `↑` / `Ctrl+K` | Navigate up |
| `↓` / `Ctrl+J` | Navigate down |
| `Enter` | Copy selected |
| `Shift+Enter` | Execute selected |
| `1` – `9` | Direct select (empty search) |
| `Ctrl+E` | Edit selected |
| `Ctrl+S` | Toggle pin |
| `Backspace` (empty) | Remove context chip |
| `Esc` / `Ctrl+C` | Cancel |

### Add / Edit form

| Key | Action |
|---|---|
| `Tab` / `Shift+Tab` | Next / previous field |
| `Ctrl+S` | Save |
| `Ctrl+W` | Toggle workflow mode |
| `Ctrl+N` | Add step (workflow) |
| `Ctrl+D` | Delete step (workflow) |
| `Tab` / `Right` | Accept autocomplete |
| `Esc` × 2 (dirty) | Discard |
| `Esc` / `Ctrl+C` | Cancel |

### Workflow runner

| Key | Action |
|---|---|
| `Enter` | Run next step |
| `Ctrl+X` | Run all remaining |
| `Esc` | Abort |

---

## Data

Single JSON file. Atomic writes (save to `.tmp`, rename over target).

| OS | Path |
|---|---|
| Linux | `$XDG_CONFIG_HOME/decoreba/commands.json` |
| macOS | `~/Library/Application Support/decoreba/commands.json` |
| Windows | `%AppData%\decoreba\commands.json` |

Override: `$DECOREBA_CONFIG`.

---

## Performance

| Metric | Value |
|---|---|
| Binary | 3.7 MB (CLI, stripped) |
| Startup | ~2 ms |
| RSS | ~3 MB |
| Deps | zero (pure Go stdlib) |
| Targets | linux amd64/arm64, macOS amd64/arm64, windows amd64 |

---

## Design

- **Appear, don't replace.** Inline overlay, not alternate screen. Occupies
  space below the prompt, leaves no trace on dismiss.
- **Right answer before you finish typing.** Fuzzy + typo tolerance + recency +
  pinning put the command on screen before the query is complete.
- **Keyboard-first.** Every action has a shortcut. Hint line teaches them
  progressively. No mouse.
- **Context over categories.** Commands live under the tool they belong to.
  Auto-detection from cwd means zero setup.
- **Accessible.** `NO_COLOR` and `--no-color` disable all ANSI escape codes.
