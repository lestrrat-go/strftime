package strftime

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"

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
	if _, ok := w.(timefmt); ok {
		return timefmt{s: v.s + w.str()}
	} else {
		return verbatim{s: v.s + w.str()}
	}
}

func (v verbatim) str() string {
	return v.s
}

type combiner interface {
	canCombine() bool
	combine(combiner) writer
	str() string
}

// does the time.Format thing
type timefmt struct {
	s string
}

func (v timefmt) Write(w io.Writer, t time.Time) error {
	if _, err := io.WriteString(w, t.Format(v.s)); err != nil {
		return errors.Wrap(err, `failed to write timefmt string`)
	}
	return nil
}

func (v timefmt) str() string {
	return v.s
}

func (v timefmt) canCombine() bool {
	return true
}

func (v timefmt) combine(w combiner) writer {
	return timefmt{s: v.s + w.str()}
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

		switch c := p[1]; c {
		case 'A':
			appendw(timefmt{s: "Monday"})
		case 'a':
			appendw(timefmt{s: "Mon"})
		case 'B':
			appendw(timefmt{s: "January"})
		case 'b', 'h':
			appendw(timefmt{s: "Jan"})
		case 'C':
			appendw(century{})
		case 'c':
			appendw(timefmt{s: "Mon Jan _2 15:04:05 2006"})
		case 'D':
			appendw(timefmt{s: "01/02/06"})
		case 'd':
			appendw(timefmt{s: "02"})
		case 'e':
			appendw(timefmt{s: "_2"})
		case 'F':
			appendw(timefmt{s: "2006-01-02"})
		case 'H':
			appendw(timefmt{s: "15"})
		case 'I':
			appendw(timefmt{s: "3"})
		case 'j':
			appendw(dayofyear{})
		case 'k':
			appendw(hourwblank(false))
		case 'l':
			appendw(hourwblank(true))
		case 'M':
			appendw(timefmt{s: "04"})
		case 'm':
			appendw(timefmt{s: "01"})
		case 'n':
			appendw(verbatim{s: "\n"})
		case 'p':
			appendw(timefmt{s: "PM"})
		case 'R':
			appendw(timefmt{s: "15:04"})
		case 'r':
			appendw(timefmt{s: "3:04:05 PM"})
		case 'S':
			appendw(timefmt{s: "05"})
		case 'T':
			appendw(timefmt{s: "15:04:05"})
		case 't':
			appendw(verbatim{s: "\t"})
		case 'U': // week number of the year, Sunday first
			appendw(weeknumberOffset(0))
		case 'u':
			appendw(weekday(1))
		case 'V':
			appendw(weeknumber{})
		case 'v':
			appendw(timefmt{s: "_2-Jan-2006"})
		case 'W': // week number of the year, Monday first
			appendw(weeknumberOffset(1))
		case 'w':
			appendw(weekday(0))
		case 'X': // national representation of the time
			// XXX is this correct?
			appendw(timefmt{s: "15:04:05"})
		case 'x': // national representation of the date
			// XXX is this correct?
			appendw(timefmt{s: "01/02/06"})
		case 'Y': // year with century
			appendw(timefmt{s: "2006"})
		case 'y': // year w/o century
			appendw(timefmt{s: "06"})
		case 'Z': // time zone name
			appendw(timefmt{s: "MST"})
		case 'z': // time zone ofset from UTC
			appendw(timefmt{s: "-0700"})
		default:
			return errors.Errorf(`unknown time format specification '%c'`, c)
		}
		p = p[2:]
	}

	return nil
}

func Format(s string, t time.Time) (string, error) {
	f, err := New(s)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := f.Format(&buf, t); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type Strftime struct {
	compiled writerList
}

func New(f string) (*Strftime, error) {
	var wl writerList
	if err := compile(&wl, f); err != nil {
		return nil, errors.Wrap(err, `failed to compile format`)
	}
	return &Strftime{
		compiled: wl,
	}, nil
}

func (f *Strftime) Format(buf io.Writer, t time.Time) error {
	wl := f.compiled
	for _, w := range wl {
		if err := w.Write(buf, t); err != nil {
			return errors.Wrap(err, `failed to format string`)
		}
	}
	return nil
}
