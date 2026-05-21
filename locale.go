package strftime

import (
	"fmt"
	"io"
	"time"
)

// MonthNames holds twelve month names, indexed by time.Month minus one
// (January is index 0, December is index 11). It is used as the argument to
// WithMonths and WithShortMonths.
type MonthNames [12]string

// WeekdayNames holds seven weekday names, indexed by time.Weekday (Sunday is
// index 0, Saturday is index 6). It is used as the argument to WithWeekdays
// and WithShortWeekdays.
type WeekdayNames [7]string

// Locale supplies the locale-dependent strings used by the name-producing
// conversion specifiers (%A, %a, %B, %b, %h, %p). The library ships no locale
// data of its own: build a Locale for your language with NewLocale, or
// implement this interface yourself to back it with any source (a map,
// computed values, an external CLDR dataset, ...).
//
// Months are addressed by time.Month and weekdays by time.Weekday, so an
// implementation never has to worry about index conventions.
//
// Some languages (Russian, Czech, Polish, Greek, ...) inflect a month name
// depending on whether it stands alone or sits next to a day number — e.g.
// Russian "январь" (stand-alone) vs "2 января" (in a date). A Locale returns a
// single form per month, so format each grammatical context with its own
// Strftime object, each compiled with the Locale that holds the matching form.
type Locale interface {
	Month(time.Month) string          // full month name (%B)
	ShortMonth(time.Month) string     // abbreviated month name (%b, %h)
	Weekday(time.Weekday) string      // full weekday name (%A)
	ShortWeekday(time.Weekday) string // abbreviated weekday name (%a)
	Meridiem(hour int) string         // AM/PM marker for the given 0-23 hour (%p)
}

// localeData is the array-backed Locale implementation produced by NewLocale
// and DefaultLocale. Its fields are unexported and never mutated after
// construction, so a Locale is safe to share across goroutines.
type localeData struct {
	months        [12]string
	shortMonths   [12]string
	weekdays      [7]string
	shortWeekdays [7]string
	meridiem      [2]string
}

func (d *localeData) Month(m time.Month) string          { return d.months[int(m)-1] }
func (d *localeData) ShortMonth(m time.Month) string     { return d.shortMonths[int(m)-1] }
func (d *localeData) Weekday(w time.Weekday) string      { return d.weekdays[int(w)] }
func (d *localeData) ShortWeekday(w time.Weekday) string { return d.shortWeekdays[int(w)] }

func (d *localeData) Meridiem(hour int) string {
	if hour < 12 {
		return d.meridiem[0]
	}
	return d.meridiem[1]
}

// englishLocaleData builds the English locale entirely from the standard
// library's time package, so no name data is hard-coded here.
func englishLocaleData() *localeData {
	d := &localeData{meridiem: [2]string{"AM", "PM"}}
	for m := time.January; m <= time.December; m++ {
		full := m.String()
		d.months[m-1] = full
		d.shortMonths[m-1] = full[:3]
	}
	for w := time.Sunday; w <= time.Saturday; w++ {
		full := w.String()
		d.weekdays[w] = full
		d.shortWeekdays[w] = full[:3]
	}
	return d
}

// DefaultLocale returns the English locale. It is the fallback for any name a
// custom Locale leaves unset, and a convenient base for NewLocale.
func DefaultLocale() Locale {
	return englishLocaleData()
}

// LocaleOption configures a Locale built by NewLocale.
type LocaleOption interface {
	configureLocale(*localeData)
}

type localeOptionFunc func(*localeData)

func (f localeOptionFunc) configureLocale(d *localeData) { f(d) }

// NewLocale creates a Locale from the supplied options, starting from the
// English locale. Any name an option leaves as the empty string keeps its
// English default, so a partially-specified Locale never produces blank
// output.
func NewLocale(options ...LocaleOption) Locale {
	d := englishLocaleData()
	for _, o := range options {
		o.configureLocale(d)
	}
	return d
}

// WithMonths sets the full month names (%B), indexed by time.Month minus one.
func WithMonths(names MonthNames) LocaleOption {
	return localeOptionFunc(func(d *localeData) { overlay12(&d.months, names) })
}

// WithShortMonths sets the abbreviated month names (%b, %h).
func WithShortMonths(names MonthNames) LocaleOption {
	return localeOptionFunc(func(d *localeData) { overlay12(&d.shortMonths, names) })
}

// WithWeekdays sets the full weekday names (%A), indexed by time.Weekday.
func WithWeekdays(names WeekdayNames) LocaleOption {
	return localeOptionFunc(func(d *localeData) { overlay7(&d.weekdays, names) })
}

// WithShortWeekdays sets the abbreviated weekday names (%a).
func WithShortWeekdays(names WeekdayNames) LocaleOption {
	return localeOptionFunc(func(d *localeData) { overlay7(&d.shortWeekdays, names) })
}

// WithMeridiem sets the AM/PM markers (%p). An empty string keeps the English
// default for that marker.
func WithMeridiem(am, pm string) LocaleOption {
	return localeOptionFunc(func(d *localeData) {
		if am != "" {
			d.meridiem[0] = am
		}
		if pm != "" {
			d.meridiem[1] = pm
		}
	})
}

func overlay12(dst *[12]string, src MonthNames) {
	for i, v := range src {
		if v != "" {
			dst[i] = v
		}
	}
}

func overlay7(dst *[7]string, src WeekdayNames) {
	for i, v := range src {
		if v != "" {
			dst[i] = v
		}
	}
}

// applyLocale registers the name-producing appenders that read from loc onto
// the specification set. It is invoked before any explicit WithSpecification
// overrides, so a caller can still replace an individual specifier.
func applyLocale(ds SpecificationSet, loc Locale) error {
	pairs := []struct {
		b byte
		a Appender
	}{
		{'A', weekdayNameAppender{loc: loc}},
		{'a', weekdayNameAppender{loc: loc, short: true}},
		{'B', monthNameAppender{loc: loc}},
		{'b', monthNameAppender{loc: loc, short: true}},
		{'h', monthNameAppender{loc: loc, short: true}},
		{'p', meridiemAppender{loc: loc}},
	}
	for _, p := range pairs {
		if err := ds.Set(p.b, p.a); err != nil {
			return fmt.Errorf("failed to apply locale for %%%c: %w", p.b, err)
		}
	}
	return nil
}

type monthNameAppender struct {
	loc   Locale
	short bool
}

func (v monthNameAppender) Append(b []byte, t time.Time) []byte {
	if v.short {
		return append(b, v.loc.ShortMonth(t.Month())...)
	}
	return append(b, v.loc.Month(t.Month())...)
}

func (v monthNameAppender) dump(out io.Writer) {
	fmt.Fprintf(out, "monthName(short=%t)", v.short)
}

type weekdayNameAppender struct {
	loc   Locale
	short bool
}

func (v weekdayNameAppender) Append(b []byte, t time.Time) []byte {
	if v.short {
		return append(b, v.loc.ShortWeekday(t.Weekday())...)
	}
	return append(b, v.loc.Weekday(t.Weekday())...)
}

func (v weekdayNameAppender) dump(out io.Writer) {
	fmt.Fprintf(out, "weekdayName(short=%t)", v.short)
}

type meridiemAppender struct {
	loc Locale
}

func (v meridiemAppender) Append(b []byte, t time.Time) []byte {
	return append(b, v.loc.Meridiem(t.Hour())...)
}

func (v meridiemAppender) dump(out io.Writer) {
	fmt.Fprintf(out, "meridiem")
}
