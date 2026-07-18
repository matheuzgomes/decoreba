package search

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
)

// accentMap strips diacritics for accent-insensitive matching.
var accentMap = map[rune]rune{
	'á': 'a', 'à': 'a', 'ã': 'a', 'â': 'a', 'ä': 'a',
	'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
	'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
	'ó': 'o', 'ò': 'o', 'õ': 'o', 'ô': 'o', 'ö': 'o',
	'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
	'ç': 'c',
	'ñ': 'n',
	'Á': 'a', 'À': 'a', 'Ã': 'a', 'Â': 'a', 'Ä': 'a',
	'É': 'e', 'È': 'e', 'Ê': 'e', 'Ë': 'e',
	'Í': 'i', 'Ì': 'i', 'Î': 'i', 'Ï': 'i',
	'Ó': 'o', 'Ò': 'o', 'Õ': 'o', 'Ô': 'o', 'Ö': 'o',
	'Ú': 'u', 'Ù': 'u', 'Û': 'u', 'Ü': 'u',
	'Ç': 'c',
	'Ñ': 'n',
}

// stripAccents removes diacritics from s. When there are none, it returns the
// original string unchanged.
func stripAccents(s string) string {
	needs := false
	for _, r := range s {
		if _, ok := accentMap[r]; ok {
			needs = true
			break
		}
	}
	if !needs {
		return s
	}
	runes := []rune(s)
	for i, r := range runes {
		if repl, ok := accentMap[r]; ok {
			runes[i] = repl
		}
	}
	return string(runes)
}

// damerauLevenshtein returns the edit distance between a and b, treating
// insertions, deletions, substitutions and adjacent transpositions each as
// cost 1.
func damerauLevenshtein(a, b []rune) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+cost)
			}
		}
	}
	return d[la][lb]
}

// recencyBonus returns a score bonus that decays exponentially. Parameters:
//
//	baseRecency = 10   — bonus right after a command is used
//	halfLife    = 48 h — time until the bonus drops to half
func recencyBonus(cmd core.Command) int {
	if cmd.LastUsedAt.IsZero() {
		return 0
	}
	hours := time.Since(cmd.LastUsedAt).Hours()
	if hours < 0 {
		return 0
	}
	bonus := 10.0 * math.Exp(-hours/48.0)
	return int(math.Round(bonus))
}

// FuzzyMatch reports whether query is a subsequence of target
// (case-insensitive, accent-insensitive). Consecutive runs and matches at
// word boundaries raise the score; gaps lower it, so very sparse matches can
// score negative.
func FuzzyMatch(query, target string) (score int, positions []int, ok bool) {
	q := []rune(strings.ToLower(stripAccents(query)))
	t := []rune(strings.ToLower(stripAccents(target)))
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

// typoMatch tries Damerau-Levenshtein ≤ 2 against each field. Returns true
// when at least one field passes.
func typoMatch(query string, fields []string) bool {
	q := []rune(strings.ToLower(stripAccents(query)))
	if len(q) == 0 {
		return true
	}
	// When the query is very short (1-2 runes), DL ≤ 2 produces far too many
	// false positives — require an exact FuzzyMatch instead.
	if len(q) <= 2 {
		return false
	}
	for _, f := range fields {
		t := []rune(strings.ToLower(stripAccents(f)))
		if damerauLevenshtein(q, t) <= 2 {
			return true
		}
	}
	return false
}

// Match scores a command against the query across title, command, context
// and tags. It returns the best field score plus the match positions inside
// the command string itself (nil when the command did not match).
//
// When no field passes a strict FuzzyMatch, title, command and context are
// retried with Damerau-Levenshtein distance ≤ 2 as a typo fallback. Those
// results receive a fixed score of 0 plus the recency bonus.
func Match(query string, c core.Command) (score int, cmdPositions []int, ok bool) {
	if query == "" {
		return recencyBonus(c), nil, true
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
		// Second pass: typo-tolerant match on title, command and context only
		// (tags are too short for DL ≤ 2 to be meaningful).
		if typoMatch(query, []string{c.Title, c.Command, c.Context}) {
			matched = true
			best = 0 // fixed score for typo results
		}
	}

	if !matched {
		return 0, nil, false
	}

	score = best + recencyBonus(c)

	// Only compute highlight positions when FuzzyMatch succeeded on the
	// command field itself (typo results get no highlight).
	if _, pos, found := FuzzyMatch(query, c.Command); found {
		cmdPositions = pos
	}
	return score, cmdPositions, true
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
