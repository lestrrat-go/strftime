package strftime

import (
	"io"
	"strconv"
	"strings"
	"time"

	bufferpool "github.com/lestrrat/go-bufferpool"
	"github.com/pkg/errors"
)

type writer interface {
	Write(io.Writer, time.Time) error
}

type writerList []writer

// These words below, as well as any decimal character
var combineExclusion = []string{
	"Mon",
	"Monday",
	"Jan",
	"January",
	"MST",
}

func canCombine(s string) bool {
	if strings.ContainsAny(s, "0123456789") {
		return false
	}
	for _, word := range combineExclusion {
		if strings.Contains(s, word) {
			return false
		}
	}
	return true
}

type verbatim struct {
	s string
}

func (v verbatim) Write(w io.Writer, _ time.Time) error {
	if _, err := io.WriteString(w, v.s); err != nil {
		return errors.Wrap(err, `failed to write verbatim string`)
	}
	return nil
}

func (v verbatim) canCombine() bool {
	return canCombine(v.s)
}

func (v verbatim) combine(w combiner) writer {
	if _, ok := w.(*timefmtw); ok {
		return timefmt(v.s + w.str())
	}
	return &verbatim{s: v.s + w.str()}
}

func (v verbatim) str() string {
	return v.s
}

type combiner interface {
	canCombine() bool
	combine(combiner) writer
	str() string
}

type century struct{}

func (v century) Write(w io.Writer, t time.Time) error {
	n := t.Year() / 100
	if n < 10 {
		if _, err := io.WriteString(w, "0"); err != nil {
			return errors.Wrap(err, `failed to write century`)
		}
	}
	if _, err := io.WriteString(w, strconv.Itoa(n)); err != nil {
		return errors.Wrap(err, `failed to write century`)
	}
	return nil
}

type weekday int

func (v weekday) Write(w io.Writer, t time.Time) error {
	n := int(t.Weekday())
	if n < int(v) {
		n += 7
	}
	if _, err := w.Write([]byte{byte(n + 48)}); err != nil {
		return errors.Wrap(err, `failed to write weekday`)
	}
	return nil
}

type weeknumberOffset int

func (v weeknumberOffset) Write(w io.Writer, t time.Time) error {
	yd := t.YearDay()
	offset := int(t.Weekday()) - int(v)
	if offset < 0 {
		offset += 7
	}

	if yd < offset {
		if _, err := io.WriteString(w, "00"); err != nil {
			return errors.Wrap(err, `failed to write week number (monday first)`)
		}
	}

	n := ((yd - offset) / 7) + 1
	s := strconv.Itoa(n)
	if n < 10 {
		if _, err := io.WriteString(w, "0"); err != nil {
			return errors.Wrap(err, `failed to write week number`)
		}
	}

	if _, err := io.WriteString(w, s); err != nil {
		return errors.Wrap(err, `failed to write week number`)
	}
	return nil
}

type weeknumber struct{}

func (v weeknumber) Write(w io.Writer, t time.Time) error {
	_, n := t.ISOWeek()
	s := strconv.Itoa(n)
	if n < 10 {
		if _, err := io.WriteString(w, "0"); err != nil {
			return errors.Wrap(err, `failed to write week number`)
		}
	}
	if _, err := io.WriteString(w, s); err != nil {
		return errors.Wrap(err, `failed to write week number`)
	}
	return nil
}

type dayofyear struct{}

func (v dayofyear) Write(w io.Writer, t time.Time) error {
	n := t.YearDay()
	if n < 10 {
		if _, err := io.WriteString(w, "00"); err != nil {
			return errors.Wrap(err, `failed to write week number`)
		}
	} else if n < 100 {
		if _, err := io.WriteString(w, "0"); err != nil {
			return errors.Wrap(err, `failed to write week number`)
		}
	}
	if _, err := io.WriteString(w, strconv.Itoa(n)); err != nil {
		return errors.Wrap(err, `failed to write week number`)
	}
	return nil
}

type hourwblank bool

func (v hourwblank) Write(w io.Writer, t time.Time) error {
	h := t.Hour()
	if bool(v) && h > 12 {
		h = h - 12
	}
	if h < 10 {
		if _, err := io.WriteString(w, " "); err != nil {
			return errors.Wrap(err, `failed to write hour`)
		}
	}
	if _, err := io.WriteString(w, strconv.Itoa(h)); err != nil {
		return errors.Wrap(err, `failed to write hour`)
	}
	return nil
}

var directives = map[byte]writer{
	'A': timefmt("Monday"),
	'a': timefmt("Mon"),
	'B': timefmt("January"),
	'b': timefmt("Jan"),
	'C': &century{},
	'h': timefmt("Jan"), // same as 'b'
	'c': timefmt("Mon Jan _2 15:04:05 2006"),
	'D': timefmt("01/02/06"),
	'd': timefmt("02"),
	'e': timefmt("_2"),
	'F': timefmt("2006-01-02"),
	'H': timefmt("15"),
	'I': timefmt("3"),
	'j': &dayofyear{},
	'k': hourwblank(false),
	'l': hourwblank(true),
	'M': timefmt("04"),
	'm': timefmt("01"),
	'n': verbatim{s: "\n"},
	'p': timefmt("PM"),
	'R': timefmt("15:04"),
	'r': timefmt("3:04:05 PM"),
	'S': timefmt("05"),
	'T': timefmt("15:04:05"),
	't': &verbatim{s: "\t"},
	'U': weeknumberOffset(0), // week number of the year, Sunday first
	'u': weekday(1),
	'V': &weeknumber{},
	'v': timefmt("_2-Jan-2006"),
	'W': weeknumberOffset(1), // week number of the year, Monday first
	'w': weekday(0),
	'X': timefmt("15:04:05"), // national representation of the time XXX is this correct?
	'x': timefmt("01/02/06"), // national representation of the date XXX is this correct?
	'Y': timefmt("2006"),     // year with century
	'y': timefmt("06"),       // year w/o century
	'Z': timefmt("MST"),      // time zone name
	'z': timefmt("-0700"),    // time zone ofset from UTC
}

func compile(wl *writerList, p string) error {
	var prev writer
	var prevCanCombine bool
	var appendw = func(w writer) {
		if prevCanCombine {
			if wc, ok := w.(combiner); ok && wc.canCombine() {
				prev = prev.(combiner).combine(wc)
				(*wl)[len(*wl)-1] = prev
				return
			}
		}

		*wl = append(*wl, w)
		prev = w
		prevCanCombine = false
		if comb, ok := w.(combiner); ok {
			if comb.canCombine() {
				prevCanCombine = true
			}
		}
	}
	for l := len(p); l > 0; l = len(p) {
		i := strings.IndexByte(p, '%')
		if i < 0 || i == l-1 {
			appendw(verbatim{s: p})
			// this is silly, but I don't trust break keywords when there's a
			// possibility of this piece of code being rearranged
			p = p[l:]
			continue
		}

		// we found a '%'. we need the next byte to decide what to do next
		// we already know that i < l - 1
		// everything up to the i is verbatim
		if i > 0 {
			appendw(verbatim{s: p[:i]})
			p = p[i:]
		}

		directive, ok := directives[p[1]]
		if !ok {
			return errors.Errorf(`unknown time format specification '%c'`, p[1])
		}
		appendw(directive)
		p = p[2:]
	}

	return nil
}

var bbpool = bufferpool.New()

// Format takes the format `s` and the time `t` to produce the
// format date/time. Note that this function re-compiles the
// pattern every time it is called.
//
// If you know beforehand that you will be reusing the pattern
// within your application, consider creating a `Strftime` object
// and reusing it.
func Format(s string, t time.Time) (string, error) {
	f, err := New(s)
	if err != nil {
		return "", err
	}

	return f.FormatString(t)
}

// Strftime is the object that represents a compiled strftime pattern
type Strftime struct {
	pattern  string
	compiled writerList
}

// New creates a new Strftime object. If the compilation fails, then
// an error is returned in the second argument.
func New(f string) (*Strftime, error) {
	var wl writerList
	if err := compile(&wl, f); err != nil {
		return nil, errors.Wrap(err, `failed to compile format`)
	}
	return &Strftime{
		pattern:  f,
		compiled: wl,
	}, nil
}

// Pattern returns the original pattern string
func (f *Strftime) Pattern() string {
	return f.pattern
}

// Format takes the destination `buf` and time `t`. It formats the date/time
// using the pre-compiled pattern, and outputs the results to `buf`
func (f *Strftime) Format(buf io.Writer, t time.Time) error {
	wl := f.compiled
	for _, w := range wl {
		if err := w.Write(buf, t); err != nil {
			return errors.Wrap(err, `failed to format string`)
		}
	}
	return nil
}

// FormatString takes the time `t` and formats it, returning the
// string containing the formated data.
func (f *Strftime) FormatString(t time.Time) (string, error) {
	buf := bbpool.Get()
	defer bbpool.Release(buf)

	if err := f.Format(buf, t); err != nil {
		return "", err
	}
	return buf.String(), nil
}
