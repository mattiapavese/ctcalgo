package ctcalgo

type segment struct {
	Label string
	Start int64
	End   int64
}

type Token struct {
	Text    string
	StartMs int64
	EndMs   int64
	Score   float32
}
