package ctcalgo

import (
	"bufio"
	"regexp"
	"strings"
)

func loadPuncFromFile() (string, error) {

	f, err := assetsFs.Open("assets/punctuations.lst")
	if err != nil {
		return "", err
	}
	defer f.Close()

	var sb strings.Builder

	// bufio.Scanner handles line-by-line reading; strips \n automatically
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Strip UTF-8 BOM on the very first line (mirrors encoding="utf-8-sig")
		line = strings.TrimPrefix(line, "\xef\xbb\xbf")

		// First tab-separated field is the punctuation character
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 0 || parts[0] == "" {
			continue
		}

		// regexp.QuoteMeta mirrors Python's re.escape()
		sb.WriteString(regexp.QuoteMeta(parts[0]))
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return sb.String(), nil
}

func buildPuncPattern() *regexp.Regexp {
	// Basic punctuation
	basicPunc := `.,?:!{}`

	// Quotation marks
	allPunctQuotes := "\u201c\u201d\u00ab\u00bb\u2039\u203a\u201e\u201a\u201f\u201b\u2019\u2018"

	// Signs
	signs := "\u003c\u003e"

	// Spanish
	spanish := "\u00a1\u00bf"

	// Armenian
	armenianPunc := "\u055a\u055b\u055c\u055d\u055e\u055f\u0589"

	// Arabic
	arabicPunc := "\u060c\u061b\u061f"

	// Chinese
	chinesePunc := "\u3002\uff0c\uff01\uff1f\uff1b\uff1a\uff08\uff09" +
		"\u300c\u300d\u300e\u300f" +
		"\uff41\uff42\uff43\uff44" +
		"\u3008\u3009\u300a\u300b" +
		"\ufe4f\u22ef\u3001\u2027\uff0f\uff5e\u2500\uff3f"

	// Hindi
	hindi := "\u0964"

	// Misc
	misc := "\";"

	// Build character class string
	allChars := basicPunc + allPunctQuotes + signs + spanish + armenianPunc + arabicPunc + chinesePunc + hindi + misc

	// Load additional punctuations from file
	filePunc, _ := loadPuncFromFile()

	// allChars has to be escaped for use inside a character class [],
	// since regexp in Go uses RE2 syntax.
	// filePunc is already re.escape()'d, so append outside QuoteMeta
	pattern := `[` + regexp.QuoteMeta(allChars) + filePunc + `]`
	return regexp.MustCompile(pattern)
}

var punctPattern = buildPuncPattern()

func replacePunctWithSpace(text string) string {
	return punctPattern.ReplaceAllString(text, " ")
}
