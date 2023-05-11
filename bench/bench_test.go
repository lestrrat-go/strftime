package bench_test

import (
	"bytes"
	"log"
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"

	fastly "github.com/fastly/go-utils/strftime"
	jehiah "github.com/jehiah/go-strftime"
	lestrrat "github.com/lestrrat-go/strftime"
	ncruces "github.com/ncruces/go-strftime"
	tebeka "github.com/tebeka/strftime"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:8080", nil))
	}()
}

const benchfmt = `%A %a %B %b %d %H %I %M %m %p %S %Y %y %Z`

func BenchmarkTebeka(b *testing.B) {
	var t time.Time
	for i := 0; i < b.N; i++ {
		tebeka.Format(benchfmt, t)
	}
}

func BenchmarkJehiah(b *testing.B) {
	// Grr, uses byte slices, and does it faster, but with more allocs
	var t time.Time
	for i := 0; i < b.N; i++ {
		jehiah.Format(benchfmt, t)
	}
}

func BenchmarkFastly(b *testing.B) {
	var t time.Time
	for i := 0; i < b.N; i++ {
		fastly.Strftime(benchfmt, t)
	}
}

func BenchmarkNcruces(b *testing.B) {
	var t time.Time
	for i := 0; i < b.N; i++ {
		ncruces.Format(benchfmt, t)
	}
}

func BenchmarkNcrucesAppend(b *testing.B) {
	var d []byte
	var t time.Time
	for i := 0; i < b.N; i++ {
		d = ncruces.AppendFormat(d[:0], benchfmt, t)
	}
}

func BenchmarkLestrrat(b *testing.B) {
	var t time.Time
	for i := 0; i < b.N; i++ {
		lestrrat.Format(benchfmt, t)
	}
}

func BenchmarkLestrratCachedString(b *testing.B) {
	var t time.Time
	f, _ := lestrrat.New(benchfmt)
	// This benchmark does not take into effect the compilation time
	for i := 0; i < b.N; i++ {
		f.FormatString(t)
	}
}

func BenchmarkLestrratCachedWriter(b *testing.B) {
	var t time.Time
	f, _ := lestrrat.New(benchfmt)
	var buf bytes.Buffer
	b.ResetTimer()

	// This benchmark does not take into effect the compilation time
	for i := 0; i < b.N; i++ {
		buf.Reset()
		f.Format(&buf, t)
	}
}

func BenchmarkLestrratCachedFormatBuffer(b *testing.B) {
	var t time.Time
	f, _ := lestrrat.New(benchfmt)
	b.ResetTimer()

	var buf []byte
	for i := 0; i < b.N; i++ {
		buf = f.FormatBuffer(buf[:0], t)
	}
}
