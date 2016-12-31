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

VERY ALPHA CODE. More tests needed

# API

## Format(string, time.Time) (string, error)

Takes the pattern and the time, and formats it. This function is a utility function that recompiles the pattern every time the function is called. If you know beforehand that you will be formatting the same pattern multiple times, consider using `New` to create a `Strftime` object and reuse it.

## New(string) (\*Strftime, error)

Takes the pattern and creates a new `Strftime` object.

## obj.Pattern() string

Returns the pattern string used to create this `Strftime` object

## obj.Format(io.Writer, time.Time) error

Formats the time according to the pre-compiled pattern, and writes the result to the specified `io.Writer`

## obj.FormatString(time.Time) (string, error)

Formats the time according to the pre-compiled pattern, and returns the result string.

# PERFORMANCE / OTHER LIBRARIES

This library is much faster than `github.com/tebeka/strftime` *IF* you can reuse the format pattern. Furthermore, depending on your pattern, we may not be able to achieve much speed gain. Patches, tests welcome.

This benchmark only uses the subset of conversion directives that are supported by *ALL* of the libraries compared.

Currently this library allocates the least (if pattern is reused), but can do better to catch up with other strftime implementations.

Somethings to consider: 

* Can it write to io.Writer?
* Which `%directive` does it handle?

```
hummingbird% go test -tags bench -benchmem -bench .
<snip>
BenchmarkTebeka-4                 300000          5094 ns/op         480 B/op         23 allocs/op
BenchmarkJehiah-4                1000000          1972 ns/op         256 B/op         17 allocs/op
BenchmarkLestrrat-4               200000          7206 ns/op        2600 B/op         75 allocs/op
BenchmarkLestrratCached-4         500000          2636 ns/op         176 B/op          2 allocs/op
PASS
ok      github.com/lestrrat/go-strftime 18.188s
```