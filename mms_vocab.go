package ctcalgo

var tk2IdxMap = map[string]int64{
	"<blank>": 0,
	"<pad>":   1,
	"</s>":    2,
	"<unk>":   3,
	"a":       4,
	"i":       5,
	"e":       6,
	"n":       7,
	"o":       8,
	"u":       9,
	"t":       10,
	"s":       11,
	"r":       12,
	"m":       13,
	"k":       14,
	"l":       15,
	"d":       16,
	"g":       17,
	"h":       18,
	"y":       19,
	"b":       20,
	"p":       21,
	"w":       22,
	"c":       23,
	"v":       24,
	"j":       25,
	"z":       26,
	"f":       27,
	"'":       28,
	"q":       29,
	"x":       30,
	"<star>":  31,
}

var idx2TkMap map[int64]string

const BLANK int64 = 0
const BLANK_TOKEN = "<blank>"

func init() {
	idx2TkMap = make(map[int64]string)
	for k, v := range tk2IdxMap {
		idx2TkMap[v] = k
	}
}
