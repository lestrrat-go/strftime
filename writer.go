package strftime

import (
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

// does the time.Format thing
type timefmtw struct {
	s string
}

func timefmt(s string) *timefmtw {
	return &timefmtw{s: s}
}

func (v timefmtw) Write(w io.Writer, t time.Time) error {
	if _, err := io.WriteString(w, t.Format(v.s)); err != nil {
		return errors.Wrap(err, `failed to write timefmtw string`)
	}
	return nil
}

func (v timefmtw) str() string {
	return v.s
}

func (v timefmtw) canCombine() bool {
	return true
}

func (v timefmtw) combine(w combiner) writer {
	return timefmt(v.s + w.str())
}

type verbatimw struct {
	s string
}

func verbatim(s string) *verbatimw {
	return &verbatimw{s: s}
}

func (v verbatimw) Write(w io.Writer, _ time.Time) error {
	if _, err := io.WriteString(w, v.s); err != nil {
		return errors.Wrap(err, `failed to write verbatim string`)
	}
	return nil
}

func (v verbatimw) canCombine() bool {
	return canCombine(v.s)
}

func (v verbatimw) combine(w combiner) writer {
	if _, ok := w.(*timefmtw); ok {
		return timefmt(v.s + w.str())
	}
	return verbatim(v.s + w.str())
}

func (v verbatimw) str() string {
	return v.s
}

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
