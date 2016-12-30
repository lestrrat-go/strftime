# go-strftime

strftime for Go

# SYNOPSIS

```go
f := strftime.New(`.... pattern ...`)
if err := f.Format(buf, time.Now()); err != nil {
    log.Println(err.Error())
}
```

# DESCRIPTION

VERY ALPHA CODE. More tests needed

# PERFORMANCE

This library is much faster than `github.com/tebeka/strftime` *IF* you can reuse the format pattern. Furthermore, depending on your pattern, we may not be able to achieve much speed gain. Patches, tests welcome.

This benchmark only uses the subset of conversion directives that are supported by *ALL* of the libraries compared.

Currently this library allocates the least (if pattern is reused), but can do better to catch up with other strftime implementations.

Somethings to consider: 

* Can it write to io.Writer?
* Which `%directive` does it handle?

```
hummingbird% go test -tags bench -benchmem -bench .
<snip>
BenchmarkTebeka-4                 200000          5321 ns/op         480 B/op         23 allocs/op
BenchmarkJehiah-4                1000000          2020 ns/op         256 B/op         17 allocs/op
BenchmarkLestrrat-4               200000          7822 ns/op        3016 B/op         92 allocs/op
BenchmarkLestrratCached-4         500000          2588 ns/op         176 B/op          2 allocs/op
PASS
ok      github.com/lestrrat/go-strftime 17.835s
```