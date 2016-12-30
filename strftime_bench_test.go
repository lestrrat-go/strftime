// +build bench

package strftime_test

import (
	"bytes"
	"log"
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"

	jehiah "github.com/jehiah/go-strftime"
	lestrrat "github.com/lestrrat/go-strftime"
	tebeka "github.com/tebeka/strftime"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:8080", nil))
	}()
}

const benchfmt = `%A %a %B %b %c %d %H %I %j %M %m %p %S %U %W %w %X %x %Y %y %Z`

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

func BenchmarkLestrrat(b *testing.B) {
	var t time.Time
	for i := 0; i < b.N; i++ {
		lestrrat.Format(benchfmt, t)
	}
}

func BenchmarkLestrratCached(b *testing.B) {
	var t time.Time
	f, _ := lestrrat.New(benchfmt)
	var buf bytes.Buffer
	b.ResetTimer()

	// This benchmark does not take into effect the compilation time
	// nor the buffer reset time
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		buf.Reset()
		b.StartTimer()
		f.Format(&buf, t)
	}
}
