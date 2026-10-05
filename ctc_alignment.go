package ctcalgo

import (
	"fmt"
	"math"
	"strings"
)

// ctcAlignment performs CTC forced alignment between emissions and a target sequence.
//
// Parameters:
//   - logProbs: 2D slice of shape [T][VocabSize] containing log probabilities
//     (e.g. log-softmax output from a wav2vec2 model)
//   - targets: 1D slice of token indices representing the target transcript
//   - blank: index of the CTC blank token (typically 0)
//
// Returns:
//   - path:  []int64  — aligned token index for each time step (length T)
//   - scores: []float32 — log-prob of the chosen token at each time step (length T)
func ctcAlignment(logProbs [][]float32, targets []int64, blank int64) ([]int64, []float32, error) {
	const negInf = -math.MaxFloat32

	T := len(logProbs)
	if T == 0 {
		return nil, nil, fmt.Errorf("logProbs must not be empty")
	}
	L := len(targets)
	if L == 0 {
		return nil, nil, fmt.Errorf("targets must not be empty")
	}

	// S is the length of the extended label sequence:
	// blank, t[0], blank, t[1], blank, ..., t[L-1], blank
	S := 2*L + 1

	// Count repeated adjacent labels — they require extra time steps.
	R := 0
	for i := 1; i < L; i++ {
		if targets[i] == targets[i-1] {
			R++
		}
	}

	if T < L+R {
		return nil, nil, fmt.Errorf("targets length is too long for CTC: T=%d, need at least %d", T, L+R)
	}

	// Only positions [start, end) of the extended sequence are reachable at
	// time step t: end grows over the first L+R steps and start over the
	// last L+R, by two where the blank between distinct labels is skipped.
	advance := func(t, start, end int) (int, int) {
		// Advance the start pointer when there aren't enough remaining frames
		if T-t <= L+R {
			if start%2 == 1 {
				nextTarget := start/2 + 1
				if nextTarget < L && targets[start/2] != targets[nextTarget] {
					start++
				}
			}
			start++
		}
		// Advance the end pointer as more of the sequence becomes reachable
		if t <= L+R {
			if end%2 == 0 && end < 2*L {
				if targets[end/2-1] != targets[end/2] {
					end++
				}
			}
			end++
		}
		return start, end
	}

	// At t=0 only the leading blank and the first label are reachable, the
	// blank only if there is a frame to spare for it.
	start := 0
	if T <= L+R {
		start = 1
	}
	end := 1
	if S > 1 {
		end = 2
	}

	// alphas[2*S]: two rolling rows (current and previous time step).
	alphas := make([]float32, 2*S)
	for i := range alphas {
		alphas[i] = negInf
	}

	// Back-pointer storage, compactly encoded as two bit-vectors.
	// We only need T-1 back-pointer rows (no back-pointer at t=0).
	backPtrOffset := make([]uint64, T-1) // starting position in the sequence at each t
	backPtrSeek := make([]uint64, T-1)   // byte offset into the bit vectors at each t

	// Bit 0 of the back-pointer pair: came from i-1
	// Bit 1 of the back-pointer pair: came from i-2
	// Both false: came from i (stayed)
	// One pair is stored per reachable position at each t >= 1: replay the
	// band to size the bit vectors exactly. Estimating it as (S+1)*(T-L)
	// falls short once T drops below ~2L, i.e. under 2 frames per target.
	capacity := 0
	for t, s, e := 1, start, end; t < T; t++ {
		s, e = advance(t, s, e)
		capacity += e - s
	}
	backPtrBit0 := make([]bool, capacity)
	backPtrBit1 := make([]bool, capacity)

	// Initialise t=0
	for i := start; i < end; i++ {
		var labelIdx int64
		if i%2 == 0 {
			labelIdx = blank
		} else {
			labelIdx = targets[i/2]
		}
		alphas[i] = logProbs[0][labelIdx]
	}

	var seek uint64
	for t := 1; t < T; t++ {
		start, end = advance(t, start, end)

		startloop := start
		curOff := t % 2
		prevOff := (t - 1) % 2

		// Clear current row
		for i := curOff * S; i < (curOff+1)*S; i++ {
			alphas[i] = negInf
		}

		backPtrSeek[t-1] = seek
		backPtrOffset[t-1] = uint64(start)

		if start == 0 {
			alphas[curOff*S] = alphas[prevOff*S] + logProbs[t][blank]
			startloop++
			seek++
		}

		for i := startloop; i < end; i++ {
			x0 := alphas[prevOff*S+i]
			x1 := alphas[prevOff*S+i-1]
			x2 := float32(negInf)

			var labelIdx int64
			if i%2 == 0 {
				labelIdx = blank
			} else {
				labelIdx = targets[i/2]
			}

			// Skip-blank transition: only valid when not on a blank position,
			// not at the very first non-blank (i==1), and the current and
			// previous labels differ.
			if i%2 != 0 && i != 1 && targets[i/2] != targets[i/2-1] {
				x2 = alphas[prevOff*S+i-2]
			}

			var result float32
			idx := seek + uint64(i-startloop)
			if x2 > x1 && x2 > x0 {
				result = x2
				backPtrBit1[idx] = true
			} else if x1 > x0 && x1 > x2 {
				result = x1
				backPtrBit0[idx] = true
			} else {
				result = x0
			}

			alphas[curOff*S+i] = result + logProbs[t][labelIdx]
		}
		seek += uint64(end - startloop)
	}

	// Pick the better of the two final positions (last blank or last label).
	lastOff := (T - 1) % 2
	ltrIdx := S - 2
	if alphas[lastOff*S+S-1] > alphas[lastOff*S+S-2] {
		ltrIdx = S - 1
	}

	// Trace back through back-pointers to recover the path.
	path := make([]int64, T)
	for t := T - 1; t >= 0; t-- {
		var lblIdx int64
		if ltrIdx%2 == 0 {
			lblIdx = blank
		} else {
			lblIdx = targets[ltrIdx/2]
		}
		path[t] = lblIdx

		if t == 0 {
			break
		}
		tPrev := t - 1
		bpIdx := backPtrSeek[tPrev] + uint64(ltrIdx) - backPtrOffset[tPrev]
		shift := 0
		if backPtrBit1[bpIdx] {
			shift += 2
		}
		if backPtrBit0[bpIdx] {
			shift += 1
		}
		ltrIdx -= shift
	}

	// Extract per-frame log-prob scores.
	scores := make([]float32, T)
	for t := 0; t < T; t++ {
		scores[t] = logProbs[t][path[t]]
	}

	return path, scores, nil
}

func getAlignments(logProbs [][]float32, tokensStarred []string) (segments []segment, scores []float32, err error) {
	tokenIndices := make([]int64, 0)
	for _, token := range tokensStarred {
		for word := range strings.SplitSeq(token, " ") {
			if val, ok := tk2IdxMap[word]; ok {
				tokenIndices = append(tokenIndices, int64(val))
			}
		}
	}

	path, scores, err := ctcAlignment(logProbs, tokenIndices, BLANK)
	if err != nil {
		return nil, nil, err
	}

	segments = mergeRepeats(path)
	return segments, scores, nil
}
