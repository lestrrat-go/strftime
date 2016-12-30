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

```
hummingbird% go test -tags bench -bench .
BenchmarkTebeka-4                 200000          7849 ns/op
BenchmarkLestrrat-4               200000         10452 ns/op
BenchmarkLestrratCached-4         500000          4116 ns/op
PASS
ok      github.com/lestrrat/go-strftime 15.869s
```