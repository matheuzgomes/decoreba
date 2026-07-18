package search

import (
	"sort"
	"strings"

	"decoreba/internal/core"
)

// FuzzyMatch reports whether query is a subsequence of target
// (case-insensitive). Consecutive runs and matches at word boundaries raise
// the score; gaps lower it, so very sparse matches can score negative.
func FuzzyMatch(query, target string) (score int, positions []int, ok bool) {
	q := []rune(strings.ToLower(query))
	t := []rune(strings.ToLower(target))
	if len(q) == 0 {
		return 0, nil, true
	}
	ti := 0
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
				positions = append(positions, ti)
				ti++
				break
			}
			consecutive = 0
			score--
		}
		if !found {
			return 0, nil, false
		}
	}
	return score, positions, true
}

// Match scores a command against the query across title, command, context and
// tags. It returns the best field score plus the match positions inside the
// command string itself (nil when the command did not match).
func Match(query string, c core.Command) (score int, cmdPositions []int, ok bool) {
	if query == "" {
		return 0, nil, true
	}
	best := 0
	matched := false
	fields := append([]string{c.Title, c.Command, c.Context}, c.Tags...)
	for _, f := range fields {
		if s, _, found := FuzzyMatch(query, f); found {
			if !matched || s > best {
				best = s
			}
			matched = true
		}
	}
	if !matched {
		return 0, nil, false
	}
	if _, pos, found := FuzzyMatch(query, c.Command); found {
		cmdPositions = pos
	}
	return best, cmdPositions, true
}

// Matches is the desktop-facing contract: sparse subsequences with a negative
// score are rejected, unlike the palette path (Match) which accepts them.
func Matches(query string, c core.Command) (int, bool) {
	score, _, ok := Match(query, c)
	if !ok || score < 0 {
		return 0, false
	}
	return score, true
}

type Scored struct {
	Cmd   core.Command
	Pos   []int
	Score int
}

// Sort filters the pool down to matching commands and orders them by score,
// breaking ties by usage count.
func Sort(pool []core.Command, query string) []Scored {
	var results []Scored
	for _, c := range pool {
		if s, pos, ok := Match(query, c); ok {
			results = append(results, Scored{c, pos, s})
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
