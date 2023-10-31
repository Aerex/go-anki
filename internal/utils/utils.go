package utils

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"html"
	"io"
	"math/rand"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	ALL_CHARACTERS     = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	BASE91_EXTRA_CHARS = "!#$%&()*+,-./:;<=>?@[]^_`{|}~"
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
	d := reMedia.ReplaceAllString(data, " \\1 ")
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

func JoinFields(fields []string) string {
	return strings.Join(fields, "\x1f")
}

func FieldChecksum(data string) string {

	//return int(checksum(stripHTMLMedia(data).encode("utf-8"))[:8], 16)
	// strip html
	data = StripHTMLMedia(data)
	// encode in utf8
	data = UTF8EncodeString(data)
	// generate checksum
	h := sha1.New()
	io.WriteString(h, data)
	return hex.EncodeToString(h.Sum(nil))
}

func CurrentModuleDir() string {
	_, fileName, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(fileName))
}

func Clone(dst interface{}, src interface{}) error {
	out, err := json.Marshal(src)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(out, src); err != nil {
		return err
	}
	return nil

}

func MissingParents(deckName string) bool {
	for _, section := range strings.Split(deckName, "::") {
		if section == "" {
			return true
		}
	}
	return false
}

func base91(num int64) string {
	return base62(num, BASE91_EXTRA_CHARS)
}

func base62(num int64, extra string) string {
	set := ALL_CHARACTERS + BASE91_EXTRA_CHARS
	var rem int64
	l := int64(len(set))
	var buf bytes.Buffer

	for {
		rem = num % l
		buf.WriteString(set[rem : rem+1])
		num = num / l
		if num != 0 {
			break
		}
	}
	return buf.String()
}

func GUID64() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return base91(r.Int63())
}
