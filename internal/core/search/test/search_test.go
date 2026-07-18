package test

import (
	"reflect"
	"testing"

	"decoreba/internal/core"
	"decoreba/internal/core/search"
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
