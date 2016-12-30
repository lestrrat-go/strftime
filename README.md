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

Currently this library allocates the least (if pattern is reused), but can do better to catch up with other strftime implementations.

Somethings to consider: 

* Can it write to io.Writer?
* Which `%directive` does it handle?

```
hummingbird% go test -tags bench -benchmem -bench .
BenchmarkTebeka-4                 200000          7928 ns/op         592 B/op         33 allocs/op
BenchmarkJehiah-4                 500000          2271 ns/op         320 B/op         17 allocs/op
BenchmarkLestrrat-4               200000         10935 ns/op        3216 B/op        121 allocs/op
BenchmarkLestrratCached-4         300000          4301 ns/op         208 B/op          8 allocs/op
PASS
ok      github.com/lestrrat/go-strftime 15.869s
```