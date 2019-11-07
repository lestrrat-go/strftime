package strftime

import (
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	fullWeekDayName             = StdlibFormat("Monday")
	abbrvWeekDayName            = StdlibFormat("Mon")
	fullMonthName               = StdlibFormat("January")
	abbrvMonthName              = StdlibFormat("Jan")
	centuryDecimal              = AppendFunc(appendCentury)
	timeAndDate                 = StdlibFormat("Mon Jan _2 15:04:05 2006")
	mdy                         = StdlibFormat("01/02/06")
	dayOfMonthZeroPad           = StdlibFormat("02")
	dayOfMonthSpacePad          = StdlibFormat("_2")
	ymd                         = StdlibFormat("2006-01-02")
	twentyFourHourClockZeroPad  = StdlibFormat("15")
	twelveHourClockZeroPad      = StdlibFormat("3")
	dayOfYear                   = AppendFunc(appendDayOfYear)
	twentyFourHourClockSpacePad = hourwblank(false)
	twelveHourClockSpacePad     = hourwblank(true)
	minutesZeroPad              = StdlibFormat("04")
	monthNumberZeroPad          = StdlibFormat("01")
	newline                     = Verbatim("\n")
	ampm                        = StdlibFormat("PM")
	hm                          = StdlibFormat("15:04")
	imsp                        = StdlibFormat("3:04:05 PM")
	secondsNumberZeroPad        = StdlibFormat("05")
	hms                         = StdlibFormat("15:04:05")
	tab                         = Verbatim("\t")
	weekNumberSundayOrigin      = weeknumberOffset(0) // week number of the year, Sunday first
	weekdayMondayOrigin         = weekday(1)
	// monday as the first day, and 01 as the first value
	weekNumberMondayOriginOneOrigin = AppendFunc(appendWeekNumber)
	eby                             = StdlibFormat("_2-Jan-2006")
	// monday as the first day, and 00 as the first value
	weekNumberMondayOrigin = weeknumberOffset(1) // week number of the year, Monday first
	weekdaySundayOrigin    = weekday(0)
	natReprTime            = StdlibFormat("15:04:05") // national representation of the time XXX is this correct?
	natReprDate            = StdlibFormat("01/02/06") // national representation of the date XXX is this correct?
	year                   = StdlibFormat("2006")     // year with century
	yearNoCentury          = StdlibFormat("06")       // year w/o century
	timezone               = StdlibFormat("MST")      // time zone name
	timezoneOffset         = StdlibFormat("-0700")    // time zone ofset from UTC
	percent                = Verbatim("%")
)

type combiningAppend struct {
	list           appenderList
	prev           Appender
	prevCanCombine bool
}

func (ca *combiningAppend) Append(w Appender) {
	if ca.prevCanCombine {
		if wc, ok := w.(combiner); ok && wc.canCombine() {
			ca.prev = ca.prev.(combiner).combine(wc)
			ca.list[len(ca.list)-1] = ca.prev
			return
		}
	}

	ca.list = append(ca.list, w)
	ca.prev = w
	ca.prevCanCombine = false
	if comb, ok := w.(combiner); ok {
		if comb.canCombine() {
			ca.prevCanCombine = true
		}
	}
}

type compileHandler interface {
	handle(Appender)
}

type appenderListBuilder struct {
	list *combiningAppend
}

func (alb *appenderListBuilder) handle(a Appender) {
	alb.list.Append(a)
}

type appenderExecutor struct {
	t   time.Time
	dst []byte
}

func (ae *appenderExecutor) handle(a Appender) {
	ae.dst = a.Append(ae.dst, ae.t)
}

func compile(handler compileHandler, p string, ds SpecificationSet) error {
	for l := len(p); l > 0; l = len(p) {
		i := strings.IndexByte(p, '%')
		if i < 0 {
			handler.handle(Verbatim(p))
			// this is silly, but I don't trust break keywords when there's a
			// possibility of this piece of code being rearranged
			p = p[l:]
			continue
		}
		if i == l-1 {
			return errors.New(`stray % at the end of pattern`)
		}

		// we found a '%'. we need the next byte to decide what to do next
		// we already know that i < l - 1
		// everything up to the i is verbatim
		if i > 0 {
			handler.handle(Verbatim(p[:i]))
			p = p[i:]
		}

		specification, err := ds.Lookup(p[1])
		if err != nil {
			return errors.Wrap(err, `pattern compilation failed`)
		}

		handler.handle(specification)
		p = p[2:]
	}
	return nil
}

func getSpecificationSetFor(options ...Option) SpecificationSet {
	var ds SpecificationSet = defaultSpecificationSet
	var extraSpecifications []*optSpecificationPair
	for _, option := range options {
		switch option.Name() {
		case optSpecificationSet:
			ds = option.Value().(SpecificationSet)
		case optSpecification:
			extraSpecifications = append(extraSpecifications, option.Value().(*optSpecificationPair))
		}
	}

	if len(extraSpecifications) > 0 {
		// If ds is immutable, we're going to need to create a new
		// one. oh what a waste!
		if raw, ok := ds.(*specificationSet); ok && !raw.mutable {
			ds = NewSpecificationSet()
		}
		for _, v := range extraSpecifications {
			ds.Set(v.name, v.appender)
		}
	}
	return ds
}

// Format takes the format `s` and the time `t` to produce the
// format date/time. Note that this function re-compiles the
// pattern every time it is called.
//
// If you know beforehand that you will be reusing the pattern
// within your application, consider creating a `Strftime` object
// and reusing it.
func Format(p string, t time.Time, options ...Option) (string, error) {
	// TODO: this may be premature optimization
	ds := getSpecificationSetFor(options...)

	var h appenderExecutor
	// TODO: optimize for 64 byte strings
	h.dst = make([]byte, 0, len(p)+10)
	h.t = t
	if err := compile(&h, p, ds); err != nil {
		return "", errors.Wrap(err, `failed to compile format`)
	}

	return string(h.dst), nil
}

// Strftime is the object that represents a compiled strftime pattern
type Strftime struct {
	pattern  string
	compiled appenderList
}

// New creates a new Strftime object. If the compilation fails, then
// an error is returned in the second argument.
func New(p string, options ...Option) (*Strftime, error) {
	// TODO: this may be premature optimization
	ds := getSpecificationSetFor(options...)

	var h appenderListBuilder
	h.list = &combiningAppend{}

	if err := compile(&h, p, ds); err != nil {
		return nil, errors.Wrap(err, `failed to compile format`)
	}

	return &Strftime{
		pattern:  p,
		compiled: h.list.list,
	}, nil
}

// Pattern returns the original pattern string
func (f *Strftime) Pattern() string {
	return f.pattern
}

// Format takes the destination `dst` and time `t`. It formats the date/time
// using the pre-compiled pattern, and outputs the results to `dst`
func (f *Strftime) Format(dst io.Writer, t time.Time) error {
	const bufSize = 64
	var b []byte
	max := len(f.pattern) + 10
	if max < bufSize {
		var buf [bufSize]byte
		b = buf[:0]
	} else {
		b = make([]byte, 0, max)
	}
	if _, err := dst.Write(f.format(b, t)); err != nil {
		return err
	}
	return nil
}

func (f *Strftime) format(b []byte, t time.Time) []byte {
	for _, w := range f.compiled {
		b = w.Append(b, t)
	}
	return b
}

// FormatString takes the time `t` and formats it, returning the
// string containing the formated data.
func (f *Strftime) FormatString(t time.Time) string {
	const bufSize = 64
	var b []byte
	max := len(f.pattern) + 10
	if max < bufSize {
		var buf [bufSize]byte
		b = buf[:0]
	} else {
		b = make([]byte, 0, max)
	}
	return string(f.format(b, t))
}
