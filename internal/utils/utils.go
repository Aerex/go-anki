package utils

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"html"
	"io"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func BoolToInt(val bool) int {
	if val {
		return 1
	}
	return 0
}

func NormalizeString(txt string) (string, error) {
	isMn := func(r rune) bool {
		return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
	}
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	nstr, _, err := transform.String(t, txt)
	return nstr, err
}

func UTF8EncodeString(data string) string {
	in := []byte(data)
	var out bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &out)
	enc.Write(in)
	enc.Close()

	return out.String()
}

func StripHTMLMedia(data string) string {
	// TODO: Move to another method
	// strip html but keep media files
	reMedia, err := regexp.Compile("(?i)<img[^>]+src=[\"']?([^\"'>]+)[\"']?[^>]*>")
	if err != nil {
		return ""
	}
	d := reMedia.ReplaceAllString(" \\1 ", data)
	return stripHTML(d)
}

// EntsToTxt replaces all html entities to friendly text
// TODO: Need to test how this is going to work
func entsToTxt(ht string) string {
	txt := strings.ReplaceAll(ht, "&nbsp;", " ")
	reEnts, _ := regexp.Compile("&#?\\w+;")
	return reEnts.ReplaceAllStringFunc(txt, func(h string) string {
		groups := reEnts.FindStringSubmatch(h)
		t := groups[0]
		return html.UnescapeString(t)
	})
}

func stripHTML(data string) string {
	reComment, _ := regexp.Compile("(?s)<!--.*?-->")
	reStyle, _ := regexp.Compile("(?si)<style.*?>.*?</style>")
	reScript, _ := regexp.Compile("(?si)<script.*?>.*?</script>")
	reTag, _ := regexp.Compile("(?s)<.*?>")
	s := reComment.ReplaceAllString(data, "")
	s = reStyle.ReplaceAllString(s, "")
	s = reScript.ReplaceAllString(s, "")
	s = reTag.ReplaceAllString(s, "")
	s = entsToTxt(s)
	return s
}

func FieldChecksum(data string) []byte {
	// strip html
	data = StripHTMLMedia(data)
	// encode in utf8
	data = UTF8EncodeString(data)
	// generate checksum
	h := sha1.New()
	io.WriteString(h, data)
	return h.Sum(nil)
}

func ArrayStringContains(item string, items []string) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}

func CurrentModuleDir() string {
	_, fileName, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(fileName))
}
