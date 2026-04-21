package ctcalgo

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

func mergeRepeats(path []int64) (segments []segment) {
	i1, i2 := int64(0), int64(0)
	for i1 < int64(len(path)) {
		for i2 < int64(len(path)) && path[i1] == path[i2] {
			i2++
		}
		segments = append(segments, segment{
			Label: idx2TkMap[path[i1]],
			Start: i1,
			End:   i2 - 1,
		})
		i1 = i2
	}
	return
}

func starTokens(tokens []string, textSplit []string, starFrequency int) ([]string, []string) {
	var tokensStarred []string
	var textStarred []string

	if starFrequency > starFrequencySegments || starFrequency < 0 {
		starFrequency = starFrequencySegments
	}

	switch starFrequency {
	case starFrequencySegments:

		tokensStarred = make([]string, 0, len(tokens)*2)
		for _, token := range tokens {
			tokensStarred = append(tokensStarred, "<star>", token)
		}

		textStarred = make([]string, 0, len(textSplit)*2)
		for _, chunk := range textSplit {
			textStarred = append(textStarred, "<star>", chunk)
		}

	case starFrequencyEdges:
		tokensStarred = make([]string, 0, len(tokens)+2)
		tokensStarred = append(tokensStarred, "<star>")
		tokensStarred = append(tokensStarred, tokens...)
		tokensStarred = append(tokensStarred, "<star>")

		textStarred = make([]string, 0, len(textSplit)+2)
		textStarred = append(textStarred, "<star>")
		textStarred = append(textStarred, textSplit...)
		textStarred = append(textStarred, "<star>")
	}

	return tokensStarred, textStarred
}

func splitText(s string) []string {
	return strings.Fields(s)
}

func getSpans(tokens []string, segments []segment) ([][]segment, error) {
	ltrIdx := 0
	tokensIdx := 0

	type interval struct {
		start int
		end   int
	}

	var intervals []interval
	start, end := 0, 0

	for segIdx, seg := range segments {
		if tokensIdx == len(tokens) {
			if segIdx != len(segments)-1 {
				return nil, errors.New("tokens exhausted before segments")
			}
			if seg.Label != BLANK_TOKEN {
				return nil, errors.New("expected trailing blank segment")
			}
			continue
		}

		curToken := strings.Split(tokens[tokensIdx], " ")
		ltr := curToken[ltrIdx]

		if seg.Label == BLANK_TOKEN {
			continue
		}

		// assuming label is numeric but token is string → convert if needed
		if fmt.Sprint(seg.Label) != ltr {
			return nil, fmt.Errorf("%v != %v", seg.Label, ltr)
		}

		if ltrIdx == 0 {
			start = segIdx
		}

		if ltrIdx == len(curToken)-1 {
			ltrIdx = 0
			tokensIdx++
			intervals = append(intervals, interval{start, segIdx})

			for tokensIdx < len(tokens) && len(tokens[tokensIdx]) == 0 {
				intervals = append(intervals, interval{segIdx, segIdx})
				tokensIdx++
			}
		} else {
			ltrIdx++
		}
	}

	var spans [][]segment

	for idx, iv := range intervals {
		start, end = iv.start, iv.end

		span := make([]segment, end-start+1)
		copy(span, segments[start:end+1])

		// left padding
		if start > 0 {
			prevSeg := segments[start-1]
			if prevSeg.Label == BLANK_TOKEN {
				var padStart int64
				if idx == 0 {
					padStart = prevSeg.Start
				} else {
					padStart = (prevSeg.Start + prevSeg.End) / 2
				}

				pad := segment{
					Label: BLANK_TOKEN,
					Start: padStart,
					End:   span[0].Start,
				}
				span = append([]segment{pad}, span...)
			}
		}

		// right padding
		if end+1 < len(segments) {
			nextSeg := segments[end+1]
			if nextSeg.Label == BLANK_TOKEN {
				var padEnd int64
				if idx == len(intervals)-1 {
					padEnd = nextSeg.End
				} else {
					padEnd = int64(math.Floor(float64(nextSeg.Start+nextSeg.End) / 2))
				}

				pad := segment{
					Label: BLANK_TOKEN,
					Start: span[len(span)-1].End,
					End:   padEnd,
				}
				span = append(span, pad)
			}
		}

		spans = append(spans, span)
	}

	return spans, nil
}

func preprocessText(text string, romanize bool) ([]string, []string) {
	splitted := splitText(text)

	normalizedTokens := make([]string, len(splitted))
	for i := range splitted {
		normalizedTokens[i] = normalizeText(strings.TrimSpace(splitted[i]), true, true, false)
	}

	if romanize {
		normalizedTokens = getRomanizedTokens(normalizedTokens, "")
	}

	return starTokens(
		normalizedTokens, splitted, starFrequencySegments)
}

func postprocessSegments(
	textStarred []string,
	spans [][]segment,
	scores []float32,
) []Token {

	tokens := []Token{}

	for i, t := range textStarred {

		if t == "<star>" {
			continue
		}

		span := spans[i]
		if len(span) == 0 {
			continue
		}

		segStartIdx := span[0].Start
		segEndIdx := span[len(span)-1].End

		// notice that 20ms corresponds to the amount of audio processed
		// in a stride of 320 samples:
		// (320/SR)*1000 = 20ms, with SR = 16000
		audioStartMs := segStartIdx * 20
		audioEndMs := (segEndIdx + 1) * 20

		// check if all labels are <star> or <blank>
		allSpecial := true
		for _, seg := range span {
			if seg.Label != "<star>" && seg.Label != "<blank>" {
				allSpecial = false
				break
			}
		}

		var score float64

		if allSpecial {
			score = math.Inf(-1)
		} else if segStartIdx == segEndIdx {
			score = float64(scores[segStartIdx])
		} else {
			sum := 0.0
			for j := segStartIdx; j < segEndIdx; j++ {
				sum += float64(scores[j])
			}
			score = sum
		}

		sample := Token{
			StartMs: audioStartMs,
			EndMs:   audioEndMs,
			Text:    t,
			Score:   float32(math.Exp(score)),
		}

		tokens = append(tokens, sample)
	}

	// fix overlaps
	for i := 0; i < len(tokens)-1; i++ {
		gap := tokens[i+1].StartMs - tokens[i].EndMs
		if gap < 0 {
			tokens[i+1].StartMs = tokens[i].EndMs
		}
	}

	return tokens
}
