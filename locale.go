package strftime

import (
	"fmt"
	"io"
	"time"
)

// MonthNames holds the twelve localized month names, indexed by
// time.Month minus one (January is index 0, December is index 11).
type MonthNames [12]string

// WeekdayNames holds the seven localized weekday names, indexed by
// time.Weekday (Sunday is index 0, Saturday is index 6).
type WeekdayNames [7]string

// Locale carries the locale-dependent strings used by the name-producing
// conversion specifiers (%A, %a, %B, %b, %h, %p). It contains no logic and
// no external data: you populate it with the names for your language and
// pass it via WithLocale.
//
// Month and weekday names are split into a "format" context and a
// "stand-alone" context. In many languages (Russian, Czech, Polish, Greek,
// ...) a month name takes a different grammatical form when it appears next
// to a day number ("2 января") than when it stands alone ("январь"). The
// format fields (Months, ShortMonths) hold the in-date form; the standalone
// fields hold the nominative form.
//
// Only the format fields are currently reachable through the standard
// conversion specifiers. To format the two contexts correctly today, compile
// two Strftime objects with two Locales — one whose Months hold the in-date
// form for patterns like "%-d %B", and one whose Months hold the stand-alone
// form for patterns like "%B %Y". The standalone fields are reserved for a
// future %OB/%Ob specifier and may be left zero if you do not need them.
//
// Any entry left as the empty string falls back to the English default, so a
// partially-filled Locale never produces blank output.
type Locale struct {
	Months      MonthNames // %B — full month name (in-date / format context)
	ShortMonths MonthNames // %b, %h — abbreviated month name (format context)

	StandaloneMonths      MonthNames // reserved for %OB — full, stand-alone context
	StandaloneShortMonths MonthNames // reserved for %Ob — abbreviated, stand-alone

	Weekdays      WeekdayNames // %A — full weekday name
	ShortWeekdays WeekdayNames // %a — abbreviated weekday name

	AMPM [2]string // %p — index 0 for hours before noon, 1 for noon onward
}

// DefaultLocale returns the English locale, derived entirely from the
// standard library's time package. It is a convenient starting point if you
// want to override only some of the names.
func DefaultLocale() Locale {
	var l Locale
	for m := time.January; m <= time.December; m++ {
		full := m.String()
		l.Months[m-1] = full
		l.ShortMonths[m-1] = full[:3]
		l.StandaloneMonths[m-1] = full
		l.StandaloneShortMonths[m-1] = full[:3]
	}
	for d := time.Sunday; d <= time.Saturday; d++ {
		full := d.String()
		l.Weekdays[d] = full
		l.ShortWeekdays[d] = full[:3]
	}
	l.AMPM = [2]string{"AM", "PM"}
	return l
}

// mergeLocale overlays the non-empty entries of loc onto the English default,
// so any name the caller did not provide degrades to English rather than to
// an empty string. The result lives on the heap and is pointed into by the
// appenders created in applyLocale, keeping it alive for their lifetime.
func mergeLocale(loc Locale) *Locale {
	merged := DefaultLocale()
	overlayMonths(&merged.Months, loc.Months)
	overlayMonths(&merged.ShortMonths, loc.ShortMonths)
	overlayMonths(&merged.StandaloneMonths, loc.StandaloneMonths)
	overlayMonths(&merged.StandaloneShortMonths, loc.StandaloneShortMonths)
	overlayWeekdays(&merged.Weekdays, loc.Weekdays)
	overlayWeekdays(&merged.ShortWeekdays, loc.ShortWeekdays)
	for i, v := range loc.AMPM {
		if v != "" {
			merged.AMPM[i] = v
		}
	}
	return &merged
}

func overlayMonths(dst *MonthNames, src MonthNames) {
	for i, v := range src {
		if v != "" {
			dst[i] = v
		}
	}
}

func overlayWeekdays(dst *WeekdayNames, src WeekdayNames) {
	for i, v := range src {
		if v != "" {
			dst[i] = v
		}
	}
}

// applyLocale registers the name-producing appenders that read from loc onto
// the specification set. It is invoked before any explicit WithSpecification
// overrides, so a caller can still replace an individual specifier.
func applyLocale(ds SpecificationSet, loc *Locale) error {
	pairs := []struct {
		b byte
		a Appender
	}{
		{'A', weekdayName{names: &loc.Weekdays}},
		{'a', weekdayName{names: &loc.ShortWeekdays}},
		{'B', monthName{names: &loc.Months}},
		{'b', monthName{names: &loc.ShortMonths}},
		{'h', monthName{names: &loc.ShortMonths}},
		{'p', ampmName{names: &loc.AMPM}},
	}
	for _, p := range pairs {
		if err := ds.Set(p.b, p.a); err != nil {
			return fmt.Errorf("failed to apply locale for %%%c: %w", p.b, err)
		}
	}
	return nil
}

type monthName struct {
	names *MonthNames
}

func (v monthName) Append(b []byte, t time.Time) []byte {
	return append(b, v.names[int(t.Month())-1]...)
}

func (v monthName) dump(out io.Writer) {
	fmt.Fprintf(out, "monthName")
}

type weekdayName struct {
	names *WeekdayNames
}

func (v weekdayName) Append(b []byte, t time.Time) []byte {
	return append(b, v.names[int(t.Weekday())]...)
}

func (v weekdayName) dump(out io.Writer) {
	fmt.Fprintf(out, "weekdayName")
}

type ampmName struct {
	names *[2]string
}

func (v ampmName) Append(b []byte, t time.Time) []byte {
	if t.Hour() < 12 {
		return append(b, v.names[0]...)
	}
	return append(b, v.names[1]...)
}

func (v ampmName) dump(out io.Writer) {
	fmt.Fprintf(out, "ampmName")
}
