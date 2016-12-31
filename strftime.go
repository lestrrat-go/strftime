package strftime

import (
	"fmt"
	"go/format"
	"io"
	"strconv"
	"strings"
	"time"

	bufferpool "github.com/lestrrat/go-bufferpool"
	"github.com/pkg/errors"
)

var directives = map[byte]writer{
	'A': timefmt("Monday"),
	'a': timefmt("Mon"),
	'B': timefmt("January"),
	'b': timefmt("Jan"),
	'C': &century{},
	'c': timefmt("Mon Jan _2 15:04:05 2006"),
	'D': timefmt("01/02/06"),
	'd': timefmt("02"),
	'e': timefmt("_2"),
	'F': timefmt("2006-01-02"),
	'H': timefmt("15"),
	'h': timefmt("Jan"), // same as 'b'
	'I': timefmt("3"),
	'j': &dayofyear{},
	'k': hourwblank(false),
	'l': hourwblank(true),
	'M': timefmt("04"),
	'm': timefmt("01"),
	'n': verbatim("\n"),
	'p': timefmt("PM"),
	'R': timefmt("15:04"),
	'r': timefmt("3:04:05 PM"),
	'S': timefmt("05"),
	'T': timefmt("15:04:05"),
	't': verbatim("\t"),
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
	'%': verbatim("%"),
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
			appendw(verbatim(p))
			// this is silly, but I don't trust break keywords when there's a
			// possibility of this piece of code being rearranged
			p = p[l:]
			continue
		}

		// we found a '%'. we need the next byte to decide what to do next
		// we already know that i < l - 1
		// everything up to the i is verbatim
		if i > 0 {
			appendw(verbatim(p[:i]))
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

func (f *Strftime) Generate(out io.Writer, name string) error {
	buf := bbpool.Get()
	defer bbpool.Release(buf)

	fmt.Fprintf(buf, "func %s(buf io.Writer, t time.Time) error {", name)
	for _, w := range f.compiled {
		switch w.(type) {
		case *verbatimw:
			fmt.Fprintf(buf, "\nif _, err := io.WriteString(buf, %s); err != nil {", strconv.Quote(w.(*verbatimw).str()))
			fmt.Fprintf(buf, "\nreturn err")
			fmt.Fprintf(buf, "\n}")
		case *timefmtw:
			fmt.Fprintf(buf, "\nif _, err := io.WriteString(buf, t.Format(%s)); err != nil {", strconv.Quote(w.(*timefmtw).str()))
			fmt.Fprintf(buf, "\nreturn err")
			fmt.Fprintf(buf, "\n}")
		}
	}
	fmt.Fprintf(buf, "\nreturn nil")
	fmt.Fprintf(buf, "\n}")

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return errors.Wrap(err, `failed to format generated code`)
	}
	if _, err := out.Write(formatted); err != nil {
		return errors.Wrap(err, `failed to write generated code`)
	}
	return nil
}
