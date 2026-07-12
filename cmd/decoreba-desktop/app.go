package main

import (
	"sort"
	"strings"
	"time"

	"decoreba/internal/core"
)

type App struct {
	store *core.Store
	mtime time.Time
}

func NewApp() *App {
	s, err := core.Load()
	if err != nil {
		s = &core.Store{Version: 1}
	}
	a := &App{store: s}
	a.recordMtime()
	return a
}

func (a *App) recordMtime() {
	path, err := core.ConfigPath()
	if err != nil {
		return
	}
	info, err := core.StatPath(path)
	if err != nil {
		return
	}
	a.mtime = info.ModTime()
}

func (a *App) Refresh() {
	s, err := core.Load()
	if err != nil {
		return
	}
	a.store = s
	a.recordMtime()
}

func (a *App) GetCommands() []core.Command {
	if a.store == nil {
		return nil
	}
	cmds := make([]core.Command, len(a.store.Commands))
	copy(cmds, a.store.Commands)
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].UsageCount > cmds[j].UsageCount
	})
	return cmds
}

type searchResult struct {
	Cmd   core.Command `json:"cmd"`
	Score int          `json:"score"`
}

func (a *App) Search(query, context string) []searchResult {
	if a.store == nil {
		return nil
	}

	pool := a.store.Commands
	if context != "" && !strings.EqualFold(context, "todos") {
		var filtered []core.Command
		for _, c := range pool {
			if strings.EqualFold(c.Context, context) {
				filtered = append(filtered, c)
			}
		}
		pool = filtered
	}

	var results []searchResult
	for _, c := range pool {
		if score, ok := core.MatchesCommand(query, c); ok {
			results = append(results, searchResult{Cmd: c, Score: score})
		}
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Cmd.UsageCount > results[j].Cmd.UsageCount
	})

	return results
}

func (a *App) GetContexts() []string {
	if a.store == nil {
		return nil
	}
	seen := map[string]bool{}
	var list []string
	for _, c := range a.store.Commands {
		ctx := strings.ToLower(c.Context)
		if !seen[ctx] {
			seen[ctx] = true
			list = append(list, ctx)
		}
	}
	sort.Strings(list)
	return list
}

func (a *App) AddCommand(ctx, title, command string) error {
	if ctx == "" || command == "" {
		return nil
	}
	now := time.Now()
	cmd := core.Command{
		ID:        core.GenID(),
		Context:   strings.ToLower(ctx),
		Title:     title,
		Command:   command,
		CreatedAt: now,
		UpdatedAt: now,
	}
	a.store.Commands = append(a.store.Commands, cmd)
	return core.Save(a.store)
}

func (a *App) CopyCommand(id string) error {
	for i := range a.store.Commands {
		if a.store.Commands[i].ID == id {
			cmd := a.store.Commands[i]
			if err := core.CopyToClipboard(cmd.Command); err != nil {
				return err
			}
			a.store.Commands[i].UsageCount++
			return core.Save(a.store)
		}
	}
	return nil
}
