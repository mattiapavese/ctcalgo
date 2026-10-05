package ctcalgo

import (
	"math"
	"math/rand"
	"testing"
)

const testVocabSize = 32 // mms vocab plus <star>

func randomLogProbs(rng *rand.Rand, T int) [][]float32 {
	emissions := make([][]float32, T)
	for t := range emissions {
		emissions[t] = make([]float32, testVocabSize)
		for v := range emissions[t] {
			emissions[t][v] = float32(rng.NormFloat64() * 3)
		}
	}
	return logSoftmax(emissions)
}

// randomTargets returns L non-blank labels; with probability pRepeat each
// label repeats the previous one
func randomTargets(rng *rand.Rand, L int, pRepeat float64) []int64 {
	targets := make([]int64, L)
	for i := range targets {
		if i > 0 && rng.Float64() < pRepeat {
			targets[i] = targets[i-1]
			continue
		}
		targets[i] = 1 + rng.Int63n(testVocabSize-1)
	}
	return targets
}

func minFrames(targets []int64) int {
	n := len(targets)
	for i := 1; i < len(targets); i++ {
		if targets[i] == targets[i-1] {
			n++
		}
	}
	return n
}

// viterbiScore is a plain full-matrix ctc viterbi: the best log-prob of any
// valid path of T frames over targets
func viterbiScore(logProbs [][]float32, targets []int64, blank int64) float64 {
	T, S := len(logProbs), 2*len(targets)+1
	label := func(s int) int64 {
		if s%2 == 0 {
			return blank
		}
		return targets[s/2]
	}
	prev := make([]float64, S)
	cur := make([]float64, S)
	for s := range prev {
		prev[s] = math.Inf(-1)
	}
	prev[0] = float64(logProbs[0][blank])
	if S > 1 {
		prev[1] = float64(logProbs[0][targets[0]])
	}
	for t := 1; t < T; t++ {
		for s := range S {
			best := prev[s]
			if s >= 1 {
				best = max(best, prev[s-1])
			}
			if s >= 2 && s%2 == 1 && targets[s/2] != targets[s/2-1] {
				best = max(best, prev[s-2])
			}
			cur[s] = best + float64(logProbs[t][label(s)])
		}
		prev, cur = cur, prev
	}
	if S == 1 {
		return prev[0]
	}
	return max(prev[S-1], prev[S-2])
}

// collapse merges repeated labels and drops blanks
func collapse(path []int64, blank int64) []int64 {
	out := []int64{}
	for t, l := range path {
		if l != blank && (t == 0 || path[t-1] != l) {
			out = append(out, l)
		}
	}
	return out
}

func checkAlignment(t *testing.T, logProbs [][]float32, targets []int64) {
	t.Helper()
	T := len(logProbs)

	path, scores, err := ctcAlignment(logProbs, targets, BLANK)
	if err != nil {
		t.Fatalf("T=%d L=%d targets=%v: unexpected error: %v", T, len(targets), targets, err)
	}
	if len(path) != T || len(scores) != T {
		t.Fatalf("T=%d L=%d: got path of %d and scores of %d frames", T, len(targets), len(path), len(scores))
	}

	got := collapse(path, BLANK)
	if len(got) != len(targets) {
		t.Fatalf("T=%d L=%d targets=%v: path collapses to %v", T, len(targets), targets, got)
	}
	for i := range got {
		if got[i] != targets[i] {
			t.Fatalf("T=%d L=%d targets=%v: path collapses to %v", T, len(targets), targets, got)
		}
	}

	var score float64
	for _, s := range scores {
		score += float64(s)
	}
	want := viterbiScore(logProbs, targets, BLANK)
	if math.Abs(score-want) > 1e-3*max(1, math.Abs(want)) {
		t.Fatalf("T=%d L=%d targets=%v: path score %f, viterbi %f", T, len(targets), targets, score, want)
	}
}

// fewer than ~2 frames per target overflowed the back-pointer storage:
// "index out of range [851] with length 851" in production
func TestCtcAlignmentTightFrames(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	targets := make([]int64, 24)
	for i := range targets {
		targets[i] = int64(4 + i%26) // no adjacent repeats
	}
	checkAlignment(t, randomLogProbs(rng, 41), targets)
}

func TestCtcAlignmentSweep(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for _, pRepeat := range []float64{0, 0.3, 1} {
		for L := 1; L <= 40; L++ {
			targets := randomTargets(rng, L, pRepeat)
			for T := minFrames(targets); T <= 3*L+5; T++ {
				checkAlignment(t, randomLogProbs(rng, T), targets)
			}
		}
	}
}

func TestCtcAlignmentInfeasible(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	targets := []int64{4, 4, 5, 6}
	if _, _, err := ctcAlignment(randomLogProbs(rng, minFrames(targets)-1), targets, BLANK); err == nil {
		t.Fatalf("expected an error with fewer frames than needed")
	}
	if _, _, err := ctcAlignment(randomLogProbs(rng, 10), nil, BLANK); err == nil {
		t.Fatalf("expected an error with no targets")
	}
	if _, err := ForcedAlignmentFromEmissions(randomLogProbs(rng, 10), "   ", ""); err == nil {
		t.Fatalf("expected an error with blank text")
	}
}
