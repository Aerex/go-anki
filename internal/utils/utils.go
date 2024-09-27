package utils

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"math/rand"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/aerex/go-anki/pkg/models"
	"github.com/google/gapid/core/math/sint"
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
		if num == 0 {
			break
		}
	}
	return buf.String()
}

// GUID64 will generate a uuid in a more compact form
func GUID64() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return base91(r.Int63())
}

func MaxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func MinInt64(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func MaxOfFloat64(a float64, b ...float64) float64 {
	v := a
	for _, x := range b {
		if x > v {
			v = x
		}
	}
	return v
}

func MaxOfFloat32(a float32, b ...float32) float32 {
	v := a
	for _, x := range b {
		if x > v {
			v = x
		}
	}
	return v
}

func MaxOfInt64(a int64, b ...int64) int64 {
	v := a
	for _, x := range b {
		if x > v {
			v = x
		}
	}
	return v
}

func AbsInt64(a int64) int64 {
	if a < 0 {
		return a * -1
	}
	return a
}

func DequeueModelID(lst []models.ID, ele *models.ID) (models.ID, []models.ID) {
	if ele != nil {
		var tmp = []models.ID{}
		for k, v := range lst {
			if v == *ele {
				tmp = append(tmp, lst[0:k]...)
				tmp = append(tmp, lst[k+1:]...)
				return v, tmp
			}
		}
	}
	element := lst[0]
	if len(lst) == 1 {
		var tmp = []models.ID{}
		return element, tmp
	}
	return element, lst[1:]
}

func EnqueueModelID(lst []models.ID, elem models.ID) []models.ID {
	lst = append(lst, elem)
	return lst
}

// FIXME: Generics still not working yet.
// Need to wait until issue is resolved
// https://github.com/golang/go/issues/59224
/*
func Dequeue[T comparable](q []T, ele *T) (ele T, q []T){
  if ele != nil {
   var tmp = []T{}
    for k, v := range q {
      if v == *ele {
        tmp = append(tmp, q[0:k]...)
        tmp = append(tmp, q[k+1:]...)
        return v, tmp
      }
    }
  }
  element := q[0]
  if len(q) == 1 {
    var tmp = []T{}
    return element, tmp
  }
  return element, q[1:]
}

func Enqueue[T comparable](q []T, elem T) []T {
  q = append(q, elem)
  return q
}
*/

func optimalPeriod(ivl int64, point int, unit int) (string, int) {
	var timeUnit string
	if AbsInt64(ivl) < 60 || unit < 1 {
		timeUnit = "seconds"
		point = point - 1
	} else if AbsInt64(ivl) < 3600 || unit < 2 {
		timeUnit = "minutes"
	} else if AbsInt64(ivl) < (60*60*24) || unit < 3 {
		timeUnit = "hours"
	} else if AbsInt64(ivl) < (60*60*24*30) || unit < 4 {
		timeUnit = "days"
	} else if AbsInt64(ivl) < (60*60*24*365) || unit < 5 {
		timeUnit = "months"
		point = point + 1
	} else {
		timeUnit = "years"
		point = point + 1
	}

	return timeUnit, sint.Max(point, 0)
}

func convertSecondsTo(seconds int64, unit string) (int64, error) {
	var convertedTo int64
	if unit == "seconds" {
		convertedTo = seconds
	} else if unit == "minutes" {
		convertedTo = seconds / 60
	} else if unit == "hours" {
		convertedTo = seconds / 3600
	} else if unit == "days" {
		convertedTo = seconds / 86400
	} else if unit == "months" {
		convertedTo = seconds / 2592000
	} else if unit == "years" {
		convertedTo = seconds / 31536000
	} else {
		return 0, fmt.Errorf("could not determine how to convert %s from seconds", unit)
	}

	return convertedTo, nil
}

// TODO: handle translations
func shortTimeFormat(unit string) string {
	return map[string]string{
		"years":   "%dy",
		"months":  "%dmo",
		"days":    "%dd",
		"hours":   "%dh",
		"minutes": "%dm",
		"seconds": "%ds",
	}[unit]
}

func pluralCount(ivl int64, point int) int64 {
	if point > 0 {
		return 2
	}
	return ivl
}

func timeTableFmt(unit string, count int64, inTime bool) string {
	timeFmt := "%d" + string(unit[0])
	if inTime {
		timeFmt = "in " + timeFmt
	}
	if count > 1 {
		timeFmt = timeFmt + "s"
	}

	return timeFmt
}

func FormatTimeSpan(ivl int64, padding int, point int, short bool, inTime bool, unit *int) (string, error) {
	var defaultUnit int = 99
	if unit == nil {
		unit = &defaultUnit
	}
	timeUnit, point := optimalPeriod(ivl, point, *unit)
	var convertErr error
	ivl, convertErr = convertSecondsTo(ivl, timeUnit)
	if convertErr != nil {
		return "", convertErr
	}
	// TODO: handle better once we have translations
	var fmtTime string
	if short {
		fmtTime = shortTimeFormat(timeUnit)
	} else {
		fmtTime = timeTableFmt(timeUnit, pluralCount(ivl, point), inTime)
	}
	return fmt.Sprintf(fmtTime, ivl), nil
}
