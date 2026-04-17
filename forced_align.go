package ctcalgo

import (
	"fmt"
	"math"
)

func logSoftmax(emission [][]float32) [][]float32 {
	out := make([][]float32, len(emission))

	for i, row := range emission {
		if len(row) == 0 {
			continue
		}

		maxVal := row[0]
		for _, v := range row {
			if v > maxVal {
				maxVal = v
			}
		}

		var sum float64
		for _, v := range row {
			sum += math.Exp(float64(v - maxVal))
		}

		logSumExp := float64(maxVal) + math.Log(sum)

		outRow := make([]float32, len(row))
		for j, v := range row {
			outRow[j] = float32(float64(v) - logSumExp)
		}

		out[i] = outRow
	}

	return out
}

func addStarTokenDim(logProbs [][]float32) [][]float32 {
	out := make([][]float32, len(logProbs))

	for i, row := range logProbs {
		newRow := make([]float32, len(row)+1)
		copy(newRow, row)
		newRow[len(newRow)-1] = 0.0
		out[i] = newRow
	}

	return out
}

func ForcedAlignmentFromEmissions(emissions [][]float32, text string, languageIso string, offsetSamples int64) ([]Token, error) {

	logProbs := addStarTokenDim(logSoftmax(emissions))

	tokensStarred, textStarred := preprocessText(text, true)

	tokens, scores, err := getAlignments(logProbs, tokensStarred)
	if err != nil {
		return nil, err
	}

	spans, err := getSpans(tokensStarred, tokens)
	if err != nil {
		return nil, err
	}

	alignedTokens := postprocessSegments(textStarred, spans, scores, offsetSamples)
	return alignedTokens, nil
}

func generateEmissions(audio []float32) ([][]float32, error) {
	// TODO
	return nil, fmt.Errorf("not implemented")
}

func ForcedAlignmentFromAudio(audio []float32, text string, languageIso string, offsetSamples int64) ([]Token, error) {
	emssions, err := generateEmissions(audio)
	if err != nil {
		return nil, err
	}
	return ForcedAlignmentFromEmissions(emssions, text, languageIso, offsetSamples)
}
