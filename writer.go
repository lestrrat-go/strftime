package strftime

import (
	"io"
	"time"

	"github.com/pkg/errors"
)

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
