package test

import (
	"reflect"
	"testing"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/search"
)

func TestFuzzyMatchEmptyQuery(t *testing.T) {
	score, pos, ok := search.FuzzyMatch("", "anything")
	if !ok || score != 0 || pos != nil {
		t.Fatalf("empty query: got score=%d pos=%v ok=%v", score, pos, ok)
	}
}

func TestFuzzyMatchSingleChar(t *testing.T) {
	score, pos, ok := search.FuzzyMatch("a", "a")
	if !ok {
		t.Fatal("expected match")
	}
	if score != 11 {
		t.Fatalf("score = %d, want 11", score)
	}
	if !reflect.DeepEqual(pos, []int{0}) {
		t.Fatalf("pos = %v, want [0]", pos)
	}
}

func TestFuzzyMatchPositions(t *testing.T) {
	target := "docker container prune"
	_, pos, ok := search.FuzzyMatch("prune", target)
	if !ok {
		t.Fatal("expected match")
	}
	want := []int{17, 18, 19, 20, 21}
	if !reflect.DeepEqual(pos, want) {
		t.Fatalf("pos = %v, want %v", pos, want)
	}
	for i, p := range pos {
		if rune(target[p]) != rune("prune"[i]) {
			t.Fatalf("pos[%d]=%d points at %q, want %q", i, p, target[p], "prune"[i])
		}
	}
}

func TestFuzzyMatchNoMatch(t *testing.T) {
	if _, _, ok := search.FuzzyMatch("xyz", "docker"); ok {
		t.Fatal("expected no match")
	}
}

func TestFuzzyMatchCaseInsensitive(t *testing.T) {
	_, pos, ok := search.FuzzyMatch("PRU", "docker prune")
	if !ok {
		t.Fatal("expected case-insensitive match")
	}
	if !reflect.DeepEqual(pos, []int{7, 8, 9}) {
		t.Fatalf("pos = %v, want [7 8 9]", pos)
	}
}

func TestFuzzyMatchUnicode(t *testing.T) {
	_, pos, ok := search.FuzzyMatch("çã", "maçã")
	if !ok {
		t.Fatal("expected unicode match")
	}
	// positions are rune indexes: m=0 a=1 ç=2 ã=3
	if !reflect.DeepEqual(pos, []int{2, 3}) {
		t.Fatalf("pos = %v, want [2 3]", pos)
	}
}

func TestFuzzyMatchConsecutiveBeatsScattered(t *testing.T) {
	scoreConsec, _, _ := search.FuzzyMatch("abc", "abc---")
	scoreScatter, _, _ := search.FuzzyMatch("abc", "axbxc")
	if scoreConsec <= scoreScatter {
		t.Fatalf("consecutive %d should beat scattered %d", scoreConsec, scoreScatter)
	}
}

func TestFuzzyMatchBoundaryBonus(t *testing.T) {
	scoreBoundary, _, _ := search.FuzzyMatch("b", "a-b")
	scoreMid, _, _ := search.FuzzyMatch("b", "axb")
	if scoreBoundary <= scoreMid {
		t.Fatalf("boundary %d should beat mid-word %d", scoreBoundary, scoreMid)
	}
}

func TestMatchFields(t *testing.T) {
	c := core.Command{
		Context: "git",
		Title:   "Undo last commit",
		Command: "git reset --soft HEAD~1",
		Tags:    []string{"undo"},
	}

	// matches the command itself: positions are returned
	score, pos, ok := search.Match("reset", c)
	if !ok || score <= 0 {
		t.Fatal("expected match in command")
	}
	if pos == nil {
		t.Fatal("expected positions for command match")
	}

	// matches only the title: no command positions
	_, pos, ok = search.Match("commit", c)
	if !ok {
		t.Fatal("expected match via title")
	}
	if pos != nil {
		t.Fatalf("expected nil positions, got %v", pos)
	}

	if _, _, ok = search.Match("kubernetes", c); ok {
		t.Fatal("expected no match")
	}

	// subsequence match with negative score (long gaps) still matches:
	// "p" sits far into "docker container prune" after many misses
	sparse := core.Command{Context: "docker", Title: "Remove stopped containers", Command: "docker container prune"}
	if _, _, ok := search.Match("p", sparse); !ok {
		t.Fatal("single char with long gaps should still match")
	}
	if score, _, ok := search.FuzzyMatch("p", "docker container prune"); !ok {
		t.Fatalf("FuzzyMatch should report the match (score=%d)", score)
	}

	// empty query matches everything with no positions
	if _, pos, ok = search.Match("", c); !ok || pos != nil {
		t.Fatal("empty query should match with nil positions")
	}
}

func TestMatchesDesktopContract(t *testing.T) {
	c := core.Command{Context: "git", Title: "Undo", Command: "git reset"}
	if _, ok := search.Matches("git", c); !ok {
		t.Fatal("expected match")
	}
	if _, ok := search.Matches("zzz", c); ok {
		t.Fatal("expected no match")
	}

	// Desktop contract: sparse subsequence with a negative score is a miss,
	// even though FuzzyMatch/Match report it as a hit for the palette.
	sparse := core.Command{Context: "docker", Title: "Remove stopped containers", Command: "docker container prune"}
	if score, _, ok := search.FuzzyMatch("p", sparse.Command); !ok || score >= 0 {
		t.Fatalf("precondition: FuzzyMatch(%q) should be a negative-score hit, got score=%d ok=%v", "p", score, ok)
	}
	if _, ok := search.Matches("p", sparse); ok {
		t.Fatal("Matches must reject negative-score hits (desktop parity)")
	}
	if _, _, ok := search.Match("p", sparse); !ok {
		t.Fatal("Match should accept sparse subsequence for the palette")
	}
}

func TestSortOrdering(t *testing.T) {
	pool := []core.Command{
		{ID: "scattered", Context: "git", Title: "Stash pop", Command: "git stash pop"},
		{ID: "exact", Context: "git", Title: "Apply stash", Command: "git stash apply"},
		{ID: "popular", Context: "git", Title: "Show stash", Command: "git stash show", UsageCount: 10},
	}
	results := search.Sort(pool, "git stash")
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	for i := 1; i < len(results); i++ {
		prev, cur := results[i-1], results[i]
		if prev.Score < cur.Score {
			t.Fatalf("results not sorted by score: %d before %d", prev.Score, cur.Score)
		}
		if prev.Score == cur.Score && prev.Cmd.UsageCount < cur.Cmd.UsageCount {
			t.Fatal("ties should be broken by usage count")
		}
	}
}

func TestSortEmpty(t *testing.T) {
	if got := search.Sort(nil, "x"); len(got) != 0 {
		t.Fatalf("got %d results, want 0", len(got))
	}
}

func TestFuzzyMatchAccentInsensitive(t *testing.T) {
	// Without accents: exact match.
	_, pos, ok := search.FuzzyMatch("maça", "maçã")
	if !ok {
		t.Fatal("expected accent-insensitive match")
	}
	if !reflect.DeepEqual(pos, []int{0, 1, 2, 3}) {
		t.Fatalf("pos = %v, want [0 1 2 3]", pos)
	}

	// Query without accent, target with accent.
	_, pos, ok = search.FuzzyMatch("proximo", "próximo")
	if !ok {
		t.Fatal("expected accent-insensitive match (query without accent)")
	}
	if !reflect.DeepEqual(pos, []int{0, 1, 2, 3, 4, 5, 6}) {
		t.Fatalf("pos = %v, want [0..6]", pos)
	}

	// Query with accent, target without.
	_, _, ok = search.FuzzyMatch("próximo", "proximo")
	if !ok {
		t.Fatal("expected accent-insensitive match (query with accent)")
	}
}

func TestMatchTypoTolerance(t *testing.T) {
	c := core.Command{
		Context: "docker",
		Title:   "Remove stopped containers",
		Command: "docker container prune",
	}

	// Transposition-like: "dackar" should match "docker" via DL.
	// (FuzzyMatch fails: d-a-c-k-a-r is not a subsequence of "docker container prune")
	score, pos, ok := search.Match("dackar", c)
	if !ok {
		t.Fatal("'dackar' should match 'docker' via typo tolerance")
	}
	if score < 0 {
		t.Fatalf("typo score should be >= 0, got %d", score)
	}
	if pos != nil {
		t.Fatal("typo match should have nil positions")
	}

	// Missing char: "dockr" should match "docker" via DL.
	// (FuzzyMatch finds d-o-c-k-r as subsequence, so positions will be non-nil.)
	if _, p, ok := search.Match("dockr", c); !ok {
		t.Fatal("'dockr' should match 'docker'")
	} else if p == nil {
		t.Fatal("'dockr' is a valid subsequence, positions should not be nil")
	}

	// Extra char: "dockker" should match.
	if _, _, ok := search.Match("dockker", c); !ok {
		t.Fatal("'dockker' should match 'docker' via typo tolerance")
	}

	// No match at all: very different strings.
	if _, _, ok := search.Match("kubernetes", c); ok {
		t.Fatal("'kubernetes' should not match")
	}
}

func TestMatchTypoShortQueryNoFalsePositive(t *testing.T) {
	c := core.Command{
		Context: "git",
		Title:   "Show stash",
		Command: "git stash show",
	}

	// 2-char query: DL does not fire (too many false positives).
	// "gi" vs "git" is distance 1, but FuzzyMatch finds it as subsequence anyway.
	if _, _, ok := search.Match("gi", c); !ok {
		t.Fatal("'gi' should match via FuzzyMatch subsequence")
	}

	// "gz" vs everything is distance 2+, DL would fire but short queries skip DL.
	// And FuzzyMatch should fail.
	if _, _, ok := search.Match("gz", c); ok {
		t.Fatal("'gz' should not match — 2 chars, no FuzzyMatch, DL disabled")
	}
}

func TestMatchTypoTagsNotChecked(t *testing.T) {
	c := core.Command{
		Context: "docker",
		Title:   "List containers",
		Command: "docker ps",
		Tags:    []string{"cleanup"},
	}

	// "laenup" vs tag "cleanup": FuzzyMatch fails (l-a-e: wrong order after a),
	// but DL distance = 1 (l→c). Tags are excluded from DL, so no match.
	if _, _, ok := search.Match("laenup", c); ok {
		t.Fatal("'laenup' should not match — DL does not check tags")
	}

	// But FuzzyMatch subsequence on tags still works.
	if _, _, ok := search.Match("clean", c); !ok {
		t.Fatal("'clean' should match tag 'cleanup' via FuzzyMatch")
	}
}

func TestRecencyBonus(t *testing.T) {
	now := time.Now()

	cRecent := core.Command{
		Context:    "docker",
		Title:      "prune",
		Command:    "docker container prune",
		LastUsedAt: now,
	}
	cOld := core.Command{
		Context:    "docker",
		Title:      "ps",
		Command:    "docker ps",
		UsageCount: 100,
		// LastUsedAt is zero — no recency bonus.
	}

	// Both match empty query. Recent one should score higher despite lower UsageCount.
	scoreRecent, _, _ := search.Match("", cRecent)
	scoreOld, _, _ := search.Match("", cOld)

	if scoreRecent <= scoreOld {
		t.Fatalf("recent (%d) should beat old (%d)", scoreRecent, scoreOld)
	}
	if scoreRecent != 10 {
		t.Fatalf("recent score = %d, want 10 (baseRecency)", scoreRecent)
	}
	if scoreOld != 0 {
		t.Fatalf("old score = %d, want 0", scoreOld)
	}

	// Old command with far past LastUsedAt should have low bonus.
	cOld.LastUsedAt = now.Add(-720 * time.Hour) // 30 days
	scoreOld, _, _ = search.Match("", cOld)
	if scoreOld > 1 {
		t.Fatalf("30-day old recency = %d, want near 0", scoreOld)
	}
}

func TestRecencyTiebreaker(t *testing.T) {
	now := time.Now()

	pool := []core.Command{
		{
			ID:         "recent",
			Context:    "docker",
			Title:      "ps",
			Command:    "docker ps",
			LastUsedAt: now,
		},
		{
			ID:         "old",
			Context:    "docker",
			Title:      "container prune",
			Command:    "docker container prune",
			UsageCount: 50,
			// LastUsedAt zero — no recency.
		},
	}

	results := search.Sort(pool, "docker")
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Cmd.ID != "recent" {
		t.Fatalf("recent command should rank first, got ID=%s", results[0].Cmd.ID)
	}
}

func TestMatchesWithRecency(t *testing.T) {
	c := core.Command{
		Context:    "git",
		Title:      "Undo",
		Command:    "git reset",
		LastUsedAt: time.Now(),
	}

	score, ok := search.Matches("git", c)
	if !ok {
		t.Fatal("expected match")
	}
	// Score should be positive (FuzzyMatch + recency bonus).
	if score <= 0 {
		t.Fatalf("score = %d, want > 0", score)
	}
}
