package strftime

import (
	"strconv"
	"time"
)

var Milliseconds = AppendFunc(func(b []byte, t time.Time) []byte {
	millisecond := int(t.Nanosecond()) / int(time.Millisecond)
	if millisecond < 100 {
		b = append(b, '0')
	}
	if millisecond < 10 {
		b = append(b, '0')
	}
	return append(b, strconv.Itoa(millisecond)...)
})
