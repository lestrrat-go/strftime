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
