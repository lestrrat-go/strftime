# go-strftime

strftime for Go

[![Build Status](https://travis-ci.org/lestrrat/go-strftime.png?branch=master)](https://travis-ci.org/lestrrat/go-strftime)

[![GoDoc](https://godoc.org/github.com/lestrrat/go-strftime?status.svg)](https://godoc.org/github.com/lestrrat/go-strftime)

# SYNOPSIS

```go
f := strftime.New(`.... pattern ...`)
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

# PERFORMANCE / OTHER LIBRARIES

Raw output:

```
// On my OS X 10.11.6, 2.9 GHz Intel Core i5, 16GB memory.
// go version devel +41908a5 Thu Dec 1 02:54:21 2016 +0000 darwin/amd64
hummingbird% go test -tags bench -benchmem -bench .
<snip>
BenchmarkTebeka-4                     300000          4525 ns/op         288 B/op         21 allocs/op
BenchmarkJehiah-4                    1000000          2076 ns/op         256 B/op         17 allocs/op
BenchmarkLestrrat-4                   200000          6021 ns/op        1848 B/op         69 allocs/op
BenchmarkLestrratCachedString-4      3000000           563 ns/op         128 B/op          2 allocs/op
BenchmarkLestrratCachedWriter-4       500000          3059 ns/op         192 B/op          3 allocs/op
PASS
ok      github.com/lestrrat/go-strftime 21.255s
```

This library is much faster than other libraries *IF* you can reuse the format pattern.
Here's the annotated list from the benchmark results. You can clearly see that (re)using a `Strftime` object
and producing a string is the fastest.

| Import Path                     | Commit                                  | Score   | Note                            |
|:--------------------------------|:----------------------------------------|--------:|:--------------------------------|
| github.com/lestrrat/go-strftime | [ac30305](lestrrat/go-strftime@ac30305) | 3000000 | Using `FormatString()` (cached) |
| github.com/jehiah/go-strftime   | [2efbe75](jehiah/go-strftime@2efbe75)   | 1000000 |                                 |
| github.com/lestrrat/go-strftime | [ac30305](lestrrat/go-strftime@ac30305) | 500000  | Using `Format()` (cached)       |
| github.com/tebeka/strftime      | [3f9c776](tebeka/strftime@3f9c776)      | 300000  |                                 |
| github.com/lestrrat/go-strftime | [ac30305](lestrrat/go-strftime@ac30305) | 200000  | Using `Format()` (NOT cached)   |

However, depending on your pattern, this speed may vary. If you find a particular pattern that seems sluggish,
please send in patches or tests.

Please also note that this benchmark only uses the subset of conversion specifications that are supported by *ALL* of the libraries compared.

Somethings to consider when making performance comparisons in the future:

* Can it write to io.Writer?
* Which `%specification` does it handle?
