package search

import (
	"reflect"
	"testing"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func TestDamerauLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "abcd", 1},
		{"abc", "abd", 1},
		{"abc", "acb", 1},
		{"abc", "xabc", 1},
		{"abc", "axbc", 1},
		{"kitten", "sitting", 3},
		{"docker", "dackar", 2},
		{"docker", "dockr", 1},
		{"docker", "dockker", 1},
	}
	for _, tt := range tests {
		got := damerauLevenshtein([]rune(tt.a), []rune(tt.b))
		if got != tt.want {
			t.Errorf("DL(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestStripAccents(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hello", "hello"},
		{"maçã", "maca"},
		{"próximo", "proximo"},
		{"àáâãäèéêëìíîïòóôõöùúûüçñ", "aaaaaeeeeiiiiooooouuuucn"},
		{"ÀÁÂÃÄÈÉÊËÌÍÎÏÒÓÔÕÖÙÚÛÜÇÑ", "aaaaaeeeeiiiiooooouuuucn"},
		{"", ""},
	}
	for _, tt := range tests {
		got := stripAccents(tt.in)
		if got != tt.want {
			t.Errorf("stripAccents(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTypoMatchShortQuery(t *testing.T) {
	if typoMatch("ab", []string{"cd", "ef"}) {
		t.Fatal("2-char query should not match via typo")
	}
	if typoMatch("a", []string{"bc"}) {
		t.Fatal("1-char query should not match via typo")
	}
}

func TestTypoMatchPasses(t *testing.T) {
	if !typoMatch("dackar", []string{"docker", "git", "undo"}) {
		t.Fatal("dackar should typo-match docker")
	}
	if typoMatch("xxxxxx", []string{"docker"}) {
		t.Fatal("xxxxxx should not typo-match docker")
	}
}

func TestRecencyBonus(t *testing.T) {
	t.Run("zero time is zero", func(t *testing.T) {
		if got := recencyBonus(core.Command{}); got != 0 {
			t.Fatalf("got %d", got)
		}
	})
	t.Run("recent is high", func(t *testing.T) {
		got := recencyBonus(core.Command{LastUsedAt: time.Now()})
		if got != 10 {
			t.Fatalf("got %d, want 10", got)
		}
	})
	t.Run("future time clamped", func(t *testing.T) {
		got := recencyBonus(core.Command{LastUsedAt: time.Now().Add(time.Hour)})
		if got != 0 {
			t.Fatalf("got %d, want 0", got)
		}
	})
	t.Run("old decays", func(t *testing.T) {
		got := recencyBonus(core.Command{LastUsedAt: time.Now().Add(-240 * time.Hour)})
		if got > 0 {
			t.Fatalf("expected 0 for 10-day old, got %d", got)
		}
	})
	t.Run("recent beats old with high usage", func(t *testing.T) {
		now := time.Now()
		cRecent := core.Command{Context: "docker", Title: "prune", Command: "docker container prune", LastUsedAt: now}
		cOld := core.Command{Context: "docker", Title: "ps", Command: "docker ps", UsageCount: 100}

		scoreRecent, _, _ := Match("", cRecent)
		scoreOld, _, _ := Match("", cOld)
		if scoreRecent <= scoreOld {
			t.Fatalf("recent (%d) should beat old with high usage (%d)", scoreRecent, scoreOld)
		}
		if scoreRecent != 10 {
			t.Fatalf("recent score = %d, want 10", scoreRecent)
		}
		if scoreOld != 0 {
			t.Fatalf("old score = %d, want 0", scoreOld)
		}

		cOld.LastUsedAt = now.Add(-720 * time.Hour)
		scoreOld, _, _ = Match("", cOld)
		if scoreOld > 1 {
			t.Fatalf("30-day old recency = %d, want near 0", scoreOld)
		}
	})
}

func TestFuzzyMatch(t *testing.T) {
	t.Run("empty query", func(t *testing.T) {
		score, pos, ok := FuzzyMatch("", "anything")
		if !ok || score != 0 || pos != nil {
			t.Fatalf("got score=%d pos=%v ok=%v", score, pos, ok)
		}
	})
	t.Run("exact", func(t *testing.T) {
		s, pos, ok := FuzzyMatch("prune", "prune")
		if !ok || s <= 0 || len(pos) != 5 {
			t.Fatalf("got score=%d pos=%v", s, pos)
		}
	})
	t.Run("subsequence", func(t *testing.T) {
		s, pos, ok := FuzzyMatch("abc", "aXbYc")
		if !ok || s <= 0 || len(pos) != 3 {
			t.Fatalf("got score=%d pos=%v", s, pos)
		}
	})
	t.Run("no match", func(t *testing.T) {
		if _, _, ok := FuzzyMatch("xyz", "abc"); ok {
			t.Fatal("expected no match")
		}
	})
	t.Run("case insensitive", func(t *testing.T) {
		_, pos, ok := FuzzyMatch("PRU", "docker prune")
		if !ok || len(pos) != 3 {
			t.Fatalf("pos=%v", pos)
		}
	})
	t.Run("accent insensitive", func(t *testing.T) {
		_, pos, ok := FuzzyMatch("proximo", "próximo")
		if !ok || len(pos) != 7 {
			t.Fatalf("pos=%v", pos)
		}
	})
	t.Run("boundary bonus", func(t *testing.T) {
		sBound, _, _ := FuzzyMatch("b", "a-b")
		sMid, _, _ := FuzzyMatch("b", "axb")
		if sBound <= sMid {
			t.Fatalf("boundary %d should beat mid %d", sBound, sMid)
		}
	})
	t.Run("consecutive beats scattered", func(t *testing.T) {
		sConsec, _, _ := FuzzyMatch("abc", "abc---")
		sScatter, _, _ := FuzzyMatch("abc", "axbxc")
		if sConsec <= sScatter {
			t.Fatalf("consecutive %d should beat scattered %d", sConsec, sScatter)
		}
	})
	t.Run("negative score for sparse", func(t *testing.T) {
		s, _, ok := FuzzyMatch("p", "docker container prune")
		if !ok {
			t.Fatal("should match")
		}
		if s >= 0 {
			t.Fatalf("sparse long-gap match should score negative, got %d", s)
		}
	})
	t.Run("unicode runes", func(t *testing.T) {
		_, pos, ok := FuzzyMatch("çã", "maçã")
		if !ok || len(pos) != 2 || pos[0] != 2 || pos[1] != 3 {
			t.Fatalf("pos=%v", pos)
		}
	})
	t.Run("single char score", func(t *testing.T) {
		score, pos, ok := FuzzyMatch("a", "a")
		if !ok {
			t.Fatal("expected match")
		}
		if score != 11 {
			t.Fatalf("score = %d, want 11", score)
		}
		if !reflect.DeepEqual(pos, []int{0}) {
			t.Fatalf("pos = %v, want [0]", pos)
		}
	})
	t.Run("positions are correct", func(t *testing.T) {
		target := "docker container prune"
		_, pos, ok := FuzzyMatch("prune", target)
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
	})
	t.Run("accent insensitive query with target without", func(t *testing.T) {
		_, _, ok := FuzzyMatch("próximo", "proximo")
		if !ok {
			t.Fatal("expected accent-insensitive match (query with accent)")
		}
	})
}

func TestMatch(t *testing.T) {
	c := core.Command{
		Context: "git",
		Title:   "Undo last commit",
		Command: "git reset --soft HEAD~1",
		Tags:    []string{"undo"},
	}

	t.Run("empty query", func(t *testing.T) {
		s, pos, ok := Match("", c)
		if !ok || s < 0 || pos != nil {
			t.Fatalf("score=%d pos=%v", s, pos)
		}
	})
	t.Run("matches command field", func(t *testing.T) {
		s, pos, ok := Match("reset", c)
		if !ok || s <= 0 || pos == nil {
			t.Fatalf("score=%d pos=%v", s, pos)
		}
	})
	t.Run("matches only title", func(t *testing.T) {
		_, pos, ok := Match("commit", c)
		if !ok {
			t.Fatal("expected match via title")
		}
		if pos != nil {
			t.Fatalf("expected nil positions for title-only match, got %v", pos)
		}
	})
	t.Run("no match", func(t *testing.T) {
		if _, _, ok := Match("kubernetes", c); ok {
			t.Fatal("expected no match")
		}
	})
	t.Run("typo tolerance", func(t *testing.T) {
		s, pos, ok := Match("dackar", core.Command{
			Context: "docker",
			Title:   "Prune",
			Command: "docker container prune",
		})
		if !ok {
			t.Fatal("expected typo match")
		}
		if s < 0 {
			t.Fatalf("typo score should be >= 0, got %d", s)
		}
		if pos != nil {
			t.Fatalf("typo match should have nil positions, got %v", pos)
		}
	})
	t.Run("tags matched via fuzzy", func(t *testing.T) {
		_, _, ok := Match("clean", core.Command{
			Context: "docker",
			Title:   "Prune",
			Command: "docker ps",
			Tags:    []string{"cleanup"},
		})
		if !ok {
			t.Fatal("expected fuzzy match on tags")
		}
	})
	t.Run("sparse subsequence still matches", func(t *testing.T) {
		sparse := core.Command{Context: "docker", Title: "Remove stopped containers", Command: "docker container prune"}
		if _, _, ok := Match("p", sparse); !ok {
			t.Fatal("single char with long gaps should still match")
		}
	})
}

func TestMatchTypoToleranceDetail(t *testing.T) {
	c := core.Command{
		Context: "docker",
		Title:   "Remove stopped containers",
		Command: "docker container prune",
	}

	t.Run("dackar typo", func(t *testing.T) {
		score, pos, ok := Match("dackar", c)
		if !ok {
			t.Fatal("'dackar' should match 'docker' via typo tolerance")
		}
		if score < 0 {
			t.Fatalf("typo score should be >= 0, got %d", score)
		}
		if pos != nil {
			t.Fatal("typo match should have nil positions")
		}
	})
	t.Run("missing char subsequence", func(t *testing.T) {
		if _, p, ok := Match("dockr", c); !ok {
			t.Fatal("'dockr' should match 'docker'")
		} else if p == nil {
			t.Fatal("'dockr' is a valid subsequence, positions should not be nil")
		}
	})
	t.Run("extra char", func(t *testing.T) {
		if _, _, ok := Match("dockker", c); !ok {
			t.Fatal("'dockker' should match 'docker' via typo tolerance")
		}
	})
	t.Run("no match", func(t *testing.T) {
		if _, _, ok := Match("kubernetes", c); ok {
			t.Fatal("'kubernetes' should not match")
		}
	})
	t.Run("short query no false positive", func(t *testing.T) {
		c2 := core.Command{Context: "git", Title: "Show stash", Command: "git stash show"}
		if _, _, ok := Match("gi", c2); !ok {
			t.Fatal("'gi' should match via FuzzyMatch subsequence")
		}
		if _, _, ok := Match("gz", c2); ok {
			t.Fatal("'gz' should not match — 2 chars, no FuzzyMatch, DL disabled")
		}
	})
	t.Run("tags not checked by DL", func(t *testing.T) {
		c3 := core.Command{Context: "docker", Title: "List containers", Command: "docker ps", Tags: []string{"cleanup"}}
		if _, _, ok := Match("laenup", c3); ok {
			t.Fatal("'laenup' should not match — DL does not check tags")
		}
		if _, _, ok := Match("clean", c3); !ok {
			t.Fatal("'clean' should match tag 'cleanup' via FuzzyMatch")
		}
	})
}

func TestMatches(t *testing.T) {
	c := core.Command{
		Context: "git",
		Title:   "Undo",
		Command: "git reset",
	}

	t.Run("positive score passes", func(t *testing.T) {
		s, ok := Matches("git", c)
		if !ok || s <= 0 {
			t.Fatalf("score=%d ok=%v", s, ok)
		}
	})
	t.Run("negative score rejected", func(t *testing.T) {
		sparse := core.Command{Context: "docker", Title: "x", Command: "docker container prune"}
		if _, ok := Matches("p", sparse); ok {
			t.Fatal("Matches should reject negative scores")
		}
	})
	t.Run("no match", func(t *testing.T) {
		if _, ok := Matches("zzz", c); ok {
			t.Fatal("expected no match")
		}
	})
	t.Run("desktop contract", func(t *testing.T) {
		if _, ok := Matches("git", c); !ok {
			t.Fatal("expected match")
		}
		if _, ok := Matches("zzz", c); ok {
			t.Fatal("expected no match")
		}
		sparse := core.Command{Context: "docker", Title: "Remove stopped containers", Command: "docker container prune"}
		if score, _, ok := FuzzyMatch("p", sparse.Command); !ok || score >= 0 {
			t.Fatalf("precondition: FuzzyMatch(%q) should be a negative-score hit, got score=%d ok=%v", "p", score, ok)
		}
		if _, ok := Matches("p", sparse); ok {
			t.Fatal("Matches must reject negative-score hits (desktop parity)")
		}
		if _, _, ok := Match("p", sparse); !ok {
			t.Fatal("Match should accept sparse subsequence for the palette")
		}
	})
	t.Run("with recency", func(t *testing.T) {
		cRecent := core.Command{Context: "git", Title: "Undo", Command: "git reset", LastUsedAt: time.Now()}
		score, ok := Matches("git", cRecent)
		if !ok {
			t.Fatal("expected match")
		}
		if score <= 0 {
			t.Fatalf("score = %d, want > 0", score)
		}
	})
}

func TestSort(t *testing.T) {
	t.Run("ordered by score desc", func(t *testing.T) {
		pool := []core.Command{
			{ID: "low", Context: "git", Title: "Stash pop", Command: "git stash pop"},
			{ID: "high", Context: "git", Title: "Apply stash", Command: "git stash apply"},
			{ID: "mid", Context: "git", Title: "Show stash", Command: "git stash show", UsageCount: 10},
		}
		results := Sort(pool, "git stash")
		if len(results) != 3 {
			t.Fatalf("got %d, want 3", len(results))
		}
		for i := 1; i < len(results); i++ {
			if results[i].Score > results[i-1].Score {
				t.Fatalf("not sorted: %d > %d at %d", results[i].Score, results[i-1].Score, i)
			}
		}
	})
	t.Run("pinned first", func(t *testing.T) {
		pool := []core.Command{
			{ID: "a", Context: "git", Title: "Z", Command: "z", Pinned: false, UsageCount: 100},
			{ID: "b", Context: "git", Title: "A", Command: "a", Pinned: true, UsageCount: 1},
		}
		results := Sort(pool, "")
		if len(results) != 2 || results[0].Cmd.ID != "b" {
			t.Fatalf("pinned should be first: %+v", results[0])
		}
	})
	t.Run("ties broken by usage", func(t *testing.T) {
		pool := []core.Command{
			{ID: "scattered", Context: "git", Title: "Stash pop", Command: "git stash pop"},
			{ID: "exact", Context: "git", Title: "Apply stash", Command: "git stash apply"},
			{ID: "popular", Context: "git", Title: "Show stash", Command: "git stash show", UsageCount: 10},
		}
		results := Sort(pool, "git stash")
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
	})
	t.Run("empty pool", func(t *testing.T) {
		if got := Sort(nil, "x"); len(got) != 0 {
			t.Fatalf("got %d", len(got))
		}
	})
	t.Run("no matches", func(t *testing.T) {
		pool := []core.Command{{Context: "git", Title: "x", Command: "y"}}
		if got := Sort(pool, "zzz"); len(got) != 0 {
			t.Fatalf("got %d", len(got))
		}
	})
	t.Run("recency as tiebreaker", func(t *testing.T) {
		now := time.Now()
		pool := []core.Command{
			{ID: "recent", Context: "docker", Title: "ps", Command: "docker ps", LastUsedAt: now},
			{ID: "old", Context: "docker", Title: "container prune", Command: "docker container prune", UsageCount: 50},
		}
		results := Sort(pool, "docker")
		if len(results) != 2 {
			t.Fatalf("got %d results, want 2", len(results))
		}
		if results[0].Cmd.ID != "recent" {
			t.Fatalf("recent command should rank first, got ID=%s", results[0].Cmd.ID)
		}
	})
}

func TestSuggest(t *testing.T) {
	pool := []core.Command{
		{Title: "Remove"},
		{Title: "Removes"},
		{Title: "Follow container logs"},
		{Title: "Show git log"},
	}

	t.Run("short query nil", func(t *testing.T) {
		if got := Suggest(pool, "a", 3); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})
	t.Run("empty pool", func(t *testing.T) {
		if got := Suggest(nil, "test", 3); len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})
	t.Run("finds close title", func(t *testing.T) {
		got := Suggest(pool, "Remve", 2)
		if len(got) != 2 {
			t.Fatalf("got %d results, want 2: %v", len(got), got)
		}
	})
	t.Run("limit n", func(t *testing.T) {
		got := Suggest(pool, "Remve", 1)
		if len(got) != 1 {
			t.Fatalf("got %d, want 1", len(got))
		}
	})
	t.Run("distant title empty", func(t *testing.T) {
		got := Suggest(pool, "zyxwvut", 3)
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})
	t.Run("accent insensitive", func(t *testing.T) {
		got := Suggest([]core.Command{{Title: "próximo"}}, "proximo", 1)
		if len(got) != 1 {
			t.Fatal("suggest should be accent-insensitive")
		}
	})
}
