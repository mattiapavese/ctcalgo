package ctcalgo

type Segment struct {
	Label string
	Start int64
	End   int64
}

type Token struct {
	Text  string
	Start int64
	End   int64
	Score float32
}
