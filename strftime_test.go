package strftime_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	envload "github.com/lestrrat/go-envload"
	"github.com/lestrrat/go-strftime"
	"github.com/stretchr/testify/assert"
)

var ref = time.Unix(1136239445, 0).UTC()

func TestFormat(t *testing.T) {
	l := envload.New()
	defer l.Restore()

	os.Setenv("LC_ALL", "C")

	s, err := strftime.Format(`%A %a %B %b %C %c %D %d %e %F %H %h %I %j %k %l %M %m %n %p %R %r %S %T %t %U %u %V %v %W %w %X %x %Y %y %Z %z`, ref)
	if !assert.NoError(t, err, `strftime.Format succeeds`) {
		return
	}

	if !assert.Equal(t, "Monday Mon January Jan 20 Mon Jan  2 22:04:05 2006 01/02/06 02  2 2006-01-02 22 Jan 10 002 22 10 04 01 \n PM 22:04 10:04:05 PM 05 22:04:05 \t 01 1 01  2-Jan-2006 01 1 22:04:05 01/02/06 2006 06 UTC +0000", s, `formatted result matches`) {
		return
	}
}

func TestGenerate(t *testing.T) {
	s, err := strftime.New(`Hello, %A World %H:%M:%S`)
	if !assert.NoError(t, err, `strftime.New should succeed`) {
		return
	}

	var buf bytes.Buffer
	err = s.Generate(&buf, "Format")
	if !assert.NoError(t, err, `s.Generate should succeed`) {
		return
	}
	t.Logf("%s", buf.String())
}
