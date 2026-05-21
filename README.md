# strftime

Fast strftime for Go

 [![](https://github.com/lestrrat-go/strftime/workflows/CI/badge.svg?branch=master)](https://github.com/lestrrat-go/strftime/actions?query=branch%3Amaster) [![Go Reference](https://pkg.go.dev/badge/github.com/lestrrat-go/strftime.svg)](https://pkg.go.dev/github.com/lestrrat-go/strftime)

# SYNOPSIS

```go
f, err := strftime.New(`.... pattern ...`)
if err := f.Format(buf, time.Now()); err != nil {
    log.Println(err.Error())
}
```

# DESCRIPTION

The goals for this library are

* Optimized for the same pattern being called repeatedly
* Be flexible about destination to write the results out
* Be as complete as possible in terms of conversion specifications

# API

## Format(string, time.Time) (string, error)

Takes the pattern and the time, and formats it. This function is a utility function that recompiles the pattern every time the function is called. If you know beforehand that you will be formatting the same pattern multiple times, consider using `New` to create a `Strftime` object and reuse it.

## New(string) (\*Strftime, error)

Takes the pattern and creates a new `Strftime` object.

## obj.Pattern() string

Returns the pattern string used to create this `Strftime` object

## obj.Format(io.Writer, time.Time) error

Formats the time according to the pre-compiled pattern, and writes the result to the specified `io.Writer`

## obj.FormatString(time.Time) string

Formats the time according to the pre-compiled pattern, and returns the result string.

# SUPPORTED CONVERSION SPECIFICATIONS

| pattern | description |
|:--------|:------------|
| %A      | national representation of the full weekday name |
| %a      | national representation of the abbreviated weekday |
| %B      | national representation of the full month name |
| %b      | national representation of the abbreviated month name |
| %C      | (year / 100) as decimal number; single digits are preceded by a zero |
| %c      | national representation of time and date |
| %D      | equivalent to %m/%d/%y |
| %d      | day of the month as a decimal number (01-31) |
| %e      | the day of the month as a decimal number (1-31); single digits are preceded by a blank |
| %F      | equivalent to %Y-%m-%d |
| %G      | the ISO week year with century as a decimal number with 4 digits |
| %g      | the ISO week year without century as a decimal number (00-99) with 2 digits |
| %H      | the hour (24-hour clock) as a decimal number (00-23) |
| %h      | same as %b |
| %I      | the hour (12-hour clock) as a decimal number (01-12) |
| %j      | the day of the year as a decimal number (001-366) |
| %k      | the hour (24-hour clock) as a decimal number (0-23); single digits are preceded by a blank |
| %l      | the hour (12-hour clock) as a decimal number (1-12); single digits are preceded by a blank |
| %M      | the minute as a decimal number (00-59) |
| %m      | the month as a decimal number (01-12) |
| %n      | a newline |
| %p      | national representation of either "ante meridiem" (a.m.)  or "post meridiem" (p.m.)  as appropriate. |
| %R      | equivalent to %H:%M |
| %r      | equivalent to %I:%M:%S %p |
| %S      | the second as a decimal number (00-60) |
| %T      | equivalent to %H:%M:%S |
| %t      | a tab |
| %U      | the week number of the year (Sunday as the first day of the week) as a decimal number (00-53) |
| %u      | the weekday (Monday as the first day of the week) as a decimal number (1-7) |
| %V      | the week number of the year (Monday as the first day of the week) as a decimal number (01-53) |
| %v      | equivalent to %e-%b-%Y |
| %W      | the week number of the year (Monday as the first day of the week) as a decimal number (00-53) |
| %w      | the weekday (Sunday as the first day of the week) as a decimal number (0-6) |
| %X      | national representation of the time |
| %x      | national representation of the date |
| %Y      | the year with century as a decimal number |
| %y      | the year without century as a decimal number (00-99) |
| %Z      | the time zone name |
| %z      | the time zone offset from UTC |
| %%      | a '%' |

# NO-PADDING FLAG

A `-` (glibc) or `#` (Windows) flag may be placed between the `%` and the
conversion specifier to suppress the leading zero/blank padding on numeric
fields. For example, given `2006-01-02 03:04:05`:

| pattern | result |
|---------|--------|
| %m      | 01     |
| %-m     | 1      |
| %d      | 02     |
| %-d     | 2      |
| %H:%M   | 03:04  |
| %-H:%-M | 3:4    |

The flag has no effect on non-numeric fields (e.g. `%-A` is identical to `%A`).

# LOCALIZATION

By default the name-producing specifiers (`%A`, `%a`, `%B`, `%b`, `%h`, `%p`)
emit English. To localize them, build a `Locale` with `NewLocale` and pass it
via `WithLocale`. The library ships no locale data of its own — you supply the
names for your language:

```go
french := strftime.NewLocale(
  strftime.WithMonths(strftime.MonthNames{
    "janvier", "février", "mars", "avril", "mai", "juin",
    "juillet", "août", "septembre", "octobre", "novembre", "décembre",
  }),
  strftime.WithWeekdays(strftime.WeekdayNames{
    "dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi",
  }),
  // WithShortMonths, WithShortWeekdays, WithMeridiem ... optional
)

s, _ := strftime.New(`%A %d %B %Y`, strftime.WithLocale(french))
// -> "lundi 02 janvier 2006"
```

`MonthNames` is indexed by month minus one (January is index 0); `WeekdayNames`
is indexed by `time.Weekday` (Sunday is index 0). Any name left empty falls
back to the English default, so a partial `Locale` never yields blank output.
Numeric specifiers (`%d`, `%m`, `%Y`, ...) are locale-invariant and unaffected.

`Locale` is an interface, so you can also implement it yourself to back the
names with a map, computed values, or an external dataset. `DefaultLocale()`
returns the English implementation.

## Inflected languages

In some languages (Russian, Czech, Polish, Greek, ...) a month name changes
form depending on whether it stands alone or appears next to a day number —
e.g. Russian "январь" (stand-alone) vs "2 января" (in a date). Because a single
`Locale` carries one form per name, format each context with its own compiled
`Strftime`:

```go
inDate, _   := strftime.New(`%d %B %Y`, strftime.WithLocale(ruInDate))   // января
header, _   := strftime.New(`%B %Y`,    strftime.WithLocale(ruStandalone)) // январь
```

# EXTENSIONS / CUSTOM SPECIFICATIONS

This library in general tries to be POSIX compliant, but sometimes you just need that
extra specification or two that is relatively widely used but is not included in the
POSIX specification.

For example, POSIX does not specify how to print out milliseconds,
but popular implementations allow `%f` or `%L` to achieve this.

For those instances, `strftime.Strftime` can be configured to use a custom set of
specifications:

```
ss := strftime.NewSpecificationSet()
ss.Set('L', ...) // provide implementation for `%L`

// pass this new specification set to the strftime instance
p, err := strftime.New(`%L`, strftime.WithSpecificationSet(ss))
p.Format(..., time.Now())
```

The implementation must implement the `Appender` interface, which is

```
type Appender interface {
  Append([]byte, time.Time) []byte
}
```

For commonly used extensions such as the millisecond example and Unix timestamp, we provide a default
implementation so the user can do one of the following:

```
// (1) Pass a specification byte and the Appender
//     This allows you to pass arbitrary Appenders
p, err := strftime.New(
  `%L`,
  strftime.WithSpecification('L', strftime.Milliseconds),
)

// (2) Pass an option that knows to use strftime.Milliseconds
p, err := strftime.New(
  `%L`,
  strftime.WithMilliseconds('L'),
)
```

Similarly for Unix Timestamp:
```
// (1) Pass a specification byte and the Appender
//     This allows you to pass arbitrary Appenders
p, err := strftime.New(
  `%s`,
  strftime.WithSpecification('s', strftime.UnixSeconds),
)

// (2) Pass an option that knows to use strftime.UnixSeconds
p, err := strftime.New(
  `%s`,
  strftime.WithUnixSeconds('s'),
)
```

If a common specification is missing, please feel free to submit a PR
(but please be sure to be able to defend how "common" it is)

## List of available extensions

- [`Milliseconds`](https://pkg.go.dev/github.com/lestrrat-go/strftime?tab=doc#Milliseconds) (related option: [`WithMilliseconds`](https://pkg.go.dev/github.com/lestrrat-go/strftime?tab=doc#WithMilliseconds));

- [`Microseconds`](https://pkg.go.dev/github.com/lestrrat-go/strftime?tab=doc#Microseconds) (related option: [`WithMicroseconds`](https://pkg.go.dev/github.com/lestrrat-go/strftime?tab=doc#WithMicroseconds));

- [`UnixSeconds`](https://pkg.go.dev/github.com/lestrrat-go/strftime?tab=doc#UnixSeconds) (related option: [`WithUnixSeconds`](https://pkg.go.dev/github.com/lestrrat-go/strftime?tab=doc#WithUnixSeconds)).


# PERFORMANCE / OTHER LIBRARIES

The benchmarks live under `bench/` and compare this library against several others.

```
// AMD Ryzen 9 7900X3D, Linux/amd64
// go version go1.26.1 linux/amd64
% go test -benchmem -bench .
goos: linux
goarch: amd64
pkg: github.com/lestrrat-go/strftime/bench
cpu: AMD Ryzen 9 7900X3D 12-Core Processor
BenchmarkTebeka-24                        	  728451	      1458 ns/op	     260 B/op	      20 allocs/op
BenchmarkJehiah-24                        	 1898193	       622.1 ns/op	     256 B/op	      17 allocs/op
BenchmarkFastly-24                        	 1356129	       881.0 ns/op	     168 B/op	       6 allocs/op
BenchmarkNcruces-24                       	 5115555	       230.7 ns/op	      64 B/op	       1 allocs/op
BenchmarkNcrucesAppend-24                 	 6263023	       199.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkLestrrat-24                      	 5860896	       206.4 ns/op	     128 B/op	       2 allocs/op
BenchmarkLestrratCachedString-24          	 6105082	       189.2 ns/op	     128 B/op	       2 allocs/op
BenchmarkLestrratCachedWriter-24          	 6648992	       168.7 ns/op	      64 B/op	       1 allocs/op
BenchmarkLestrratCachedFormatBuffer-24    	 8669540	       136.7 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/lestrrat-go/strftime/bench	13.281s
```

This library is the fastest of the bunch across every access pattern. The annotated
list below ranks the relevant variants from fastest to slowest:

| Import Path                         | ns/op | allocs | Note                                          |
|:------------------------------------|------:|-------:|:----------------------------------------------|
| github.com/lestrrat-go/strftime     | 136.7 |      0 | `FormatBuffer()` into a reused slice (cached) |
| github.com/lestrrat-go/strftime     | 168.7 |      1 | `Format()` to an `io.Writer` (cached)         |
| github.com/lestrrat-go/strftime     | 189.2 |      2 | `FormatString()` (cached)                     |
| github.com/ncruces/go-strftime      | 199.2 |      0 | `AppendFormat()`                              |
| github.com/lestrrat-go/strftime     | 206.4 |      2 | package-level `Format()` (compiled patterns are cached) |
| github.com/ncruces/go-strftime      | 230.7 |      1 | `Format()`                                    |
| github.com/jehiah/go-strftime       | 622.1 |     17 |                                               |
| github.com/fastly/go-utils/strftime | 881.0 |      6 |                                               |
| github.com/tebeka/strftime          |  1458 |     20 |                                               |

The fastest path is reusing a `Strftime` object and appending into a slice you own
(`FormatBuffer`), which allocates nothing. The package-level `Format()` caches compiled
patterns internally (bounded), so even repeated one-off calls with the same pattern stay fast.

However, depending on your pattern, this speed may vary. If you find a particular pattern that seems sluggish,
please send in patches or tests.

Please also note that this benchmark only uses the subset of conversion specifications that are supported by *ALL* of the libraries compared.

Somethings to consider when making performance comparisons in the future:

* Can it write to io.Writer?
* Which `%specification` does it handle?
