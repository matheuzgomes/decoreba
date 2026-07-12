package core

import (
	"fmt"
	"sort"
	"strings"
)

func fuzzyScore(query, target string) int {
	q := strings.ToLower(query)
	t := strings.ToLower(target)
	if q == "" {
		return 0
	}
	ti := 0
	score := 0
	consecutive := 0
	for qi := 0; qi < len(q); qi++ {
		found := false
		for ; ti < len(t); ti++ {
			if t[ti] == q[qi] {
				found = true
				consecutive++
				score += consecutive * 3
				if ti == 0 || t[ti-1] == ' ' || t[ti-1] == '-' || t[ti-1] == '_' {
					score += 8
				}
				ti++
				break
			}
			consecutive = 0
			score--
		}
		if !found {
			return -1
		}
	}
	return score
}

func MatchesCommand(query string, c Command) (int, bool) {
	if query == "" {
		return 0, true
	}
	best := -1
	fields := append([]string{c.Title, c.Command, c.Context}, c.Tags...)
	for _, f := range fields {
		if s := fuzzyScore(query, f); s > best {
			best = s
		}
	}
	if best < 0 {
		return 0, false
	}
	return best, true
}

func printContexts(store *Store) {
	counts := map[string]int{}
	for _, c := range store.Commands {
		counts[c.Context]++
	}
	if len(counts) == 0 {
		fmt.Println("No commands saved yet. Use 'decoreba add' to start.")
		return
	}
	var names []string
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Println("Available contexts:")
	for _, name := range names {
		fmt.Printf("  - %s (%d)\n", name, counts[name])
	}
}

func runSearch(context, query string) {
	store, err := Load()
	check(err)

	var pool []Command
	if context != "" {
		for _, c := range store.Commands {
			if strings.EqualFold(c.Context, context) {
				pool = append(pool, c)
			}
		}
		if len(pool) == 0 {
			fmt.Printf("No commands in context %q yet.\n\n", context)
			printContexts(store)
			return
		}
	} else {
		pool = store.Commands
	}

	if query == "" {
		label := "Search> "
		if context != "" {
			label = fmt.Sprintf("Search in %s> ", context)
		}
		query = promptLine(label)
	}

	type scored struct {
		cmd   Command
		score int
	}
	var results []scored
	for _, c := range pool {
		if score, ok := MatchesCommand(query, c); ok {
			results = append(results, scored{c, score})
		}
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		return results[i].cmd.UsageCount > results[j].cmd.UsageCount
	})

	if len(results) == 0 {
		fmt.Println("No commands found.")
		return
	}

	cmdList := make([]Command, len(results))
	for i, r := range results {
		cmdList[i] = r.cmd
	}

	chosen, err := runSelector(cmdList)
	if err != nil || chosen == nil {
		return
	}
	if err := CopyToClipboard(chosen.Command); err != nil {
		fmt.Printf("Could not copy automatically (%v).\nCommand: %s\n", err, chosen.Command)
	} else {
		fmt.Printf("✓ Copied: %s\n", chosen.Command)
	}

	for i := range store.Commands {
		if store.Commands[i].ID == chosen.ID {
			store.Commands[i].UsageCount++
			break
		}
	}
	_ = Save(store)
}
