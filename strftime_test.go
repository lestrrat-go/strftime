package strftime_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	envload "github.com/lestrrat-go/envload"
	"github.com/lestrrat-go/strftime"
	"github.com/stretchr/testify/assert"
)

var ref = time.Unix(1136239445, 123456789).UTC()

func TestExclusion(t *testing.T) {
	s, err := strftime.New("%p PM")
	if !assert.NoError(t, err, `strftime.New should succeed`) {
		return
	}

	var tm time.Time
	if !assert.Equal(t, "AM PM", s.FormatString(tm)) {
		return
	}
}

func TestInvalid(t *testing.T) {
	_, err := strftime.New("%")
	if !assert.Error(t, err, `strftime.New should return error`) {
		return
	}

	_, err = strftime.New(" %")
	if !assert.Error(t, err, `strftime.New should return error`) {
		return
	}
	_, err = strftime.New(" % ")
	if !assert.Error(t, err, `strftime.New should return error`) {
		return
	}
}

func TestFormatMethods(t *testing.T) {
	l := envload.New()
	defer l.Restore()

	os.Setenv("LC_ALL", "C")

	formatString := `%A %a %B %b %C %c %D %d %e %F %H %h %I %j %k %l %M %m %n %p %R %r %S %T %t %U %u %V %v %W %w %X %x %Y %y %Z %z`
	resultString := "Monday Mon January Jan 20 Mon Jan  2 22:04:05 2006 01/02/06 02  2 2006-01-02 22 Jan 10 002 22 10 04 01 \n PM 22:04 10:04:05 PM 05 22:04:05 \t 01 1 01  2-Jan-2006 01 1 22:04:05 01/02/06 2006 06 UTC +0000"

	s, err := strftime.Format(formatString, ref)
	if !assert.NoError(t, err, `strftime.Format succeeds`) {
		return
	}

	if !assert.Equal(t, resultString, s, `formatted result matches`) {
		return
	}

	formatter, err := strftime.New(formatString)
	if !assert.NoError(t, err, `strftime.New succeeds`) {
		return
	}

	if !assert.Equal(t, resultString, formatter.FormatString(ref), `formatted result matches`) {
		return
	}

	var buf bytes.Buffer
	err = formatter.Format(&buf, ref)
	if !assert.NoError(t, err, `Format method succeeds`) {
		return
	}
	if !assert.Equal(t, resultString, buf.String(), `formatted result matches`) {
		return
	}

	var dst []byte
	dst = formatter.FormatBuffer(dst, ref)
	if !assert.Equal(t, resultString, string(dst), `formatted result matches`) {
		return
	}

	dst = []byte("nonsense")
	dst = formatter.FormatBuffer(dst[:0], ref)
	if !assert.Equal(t, resultString, string(dst), `overwritten result matches`) {
		return
	}

	dst = []byte("nonsense")
	dst = formatter.FormatBuffer(dst, ref)
	if !assert.Equal(t, "nonsense"+resultString, string(dst), `appended result matches`) {
		return
	}

}

func TestFormatBlanks(t *testing.T) {
	l := envload.New()
	defer l.Restore()

	os.Setenv("LC_ALL", "C")

	{
		dt := time.Date(1, 1, 1, 18, 0, 0, 0, time.UTC)
		s, err := strftime.Format("%l", dt)
		if !assert.NoError(t, err, `strftime.Format succeeds`) {
			return
		}

		if !assert.Equal(t, " 6", s, "leading blank is properly set") {
			return
		}
	}
	{
		dt := time.Date(1, 1, 1, 6, 0, 0, 0, time.UTC)
		s, err := strftime.Format("%k", dt)
		if !assert.NoError(t, err, `strftime.Format succeeds`) {
			return
		}

		if !assert.Equal(t, " 6", s, "leading blank is properly set") {
			return
		}
	}
}

func TestFormatZeropad(t *testing.T) {
	l := envload.New()
	defer l.Restore()

	os.Setenv("LC_ALL", "C")

	{
		dt := time.Date(1, 1, 1, 1, 0, 0, 0, time.UTC)
		s, err := strftime.Format("%j", dt)
		if !assert.NoError(t, err, `strftime.Format succeeds`) {
			return
		}

		if !assert.Equal(t, "001", s, "padding is properly set") {
			return
		}
	}
	{
		dt := time.Date(1, 1, 10, 6, 0, 0, 0, time.UTC)
		s, err := strftime.Format("%j", dt)
		if !assert.NoError(t, err, `strftime.Format succeeds`) {
			return
		}

		if !assert.Equal(t, "010", s, "padding is properly set") {
			return
		}
	}
	{
		dt := time.Date(1, 6, 1, 6, 0, 0, 0, time.UTC)
		s, err := strftime.Format("%j", dt)
		if !assert.NoError(t, err, `strftime.Format succeeds`) {
			return
		}

		if !assert.Equal(t, "152", s, "padding is properly set") {
			return
		}
	}
	{
		dt := time.Date(100, 1, 1, 1, 0, 0, 0, time.UTC)
		s, err := strftime.Format("%C", dt)
		if !assert.NoError(t, err, `strftime.Format succeeds`) {
			return
		}

		if !assert.Equal(t, "01", s, "padding is properly set") {
			return
		}
	}
}

func TestGHIssue5(t *testing.T) {
	const expected = `apm-test/logs/apm.log.01000101`
	p, _ := strftime.New("apm-test/logs/apm.log.%Y%m%d")
	dt := time.Date(100, 1, 1, 1, 0, 0, 0, time.UTC)
	if !assert.Equal(t, expected, p.FormatString(dt), `patterns including 'pm' should be treated as verbatim formatter`) {
		return
	}
}

func TestGHPR7(t *testing.T) {
	const expected = `123`

	p, _ := strftime.New(`%L`, strftime.WithMilliseconds('L'))
	if !assert.Equal(t, expected, p.FormatString(ref), `patterns should match for custom specification`) {
		return
	}
}

func TestWithMicroseconds(t *testing.T) {
	const expected = `123456`

	p, _ := strftime.New(`%f`, strftime.WithMicroseconds('f'))
	if !assert.Equal(t, expected, p.FormatString(ref), `patterns should match for custom specification`) {
		return
	}
}

func TestWithUnixSeconds(t *testing.T) {
	const expected = `1136239445`

	p, _ := strftime.New(`%s`, strftime.WithUnixSeconds('s'))
	if !assert.Equal(t, expected, p.FormatString(ref), `patterns should match for custom specification`) {
		return
	}
}

func ExampleSpecificationSet() {
	{
		// I want %L as milliseconds!
		p, err := strftime.New(`%L`, strftime.WithMilliseconds('L'))
		if err != nil {
			fmt.Println(err)
			return
		}
		p.Format(os.Stdout, ref)
		os.Stdout.Write([]byte{'\n'})
	}

	{
		// I want %f as milliseconds!
		p, err := strftime.New(`%f`, strftime.WithMilliseconds('f'))
		if err != nil {
			fmt.Println(err)
			return
		}
		p.Format(os.Stdout, ref)
		os.Stdout.Write([]byte{'\n'})
	}

	{
		// I want %X to print out my name!
		a := strftime.Verbatim(`Daisuke Maki`)
		p, err := strftime.New(`%X`, strftime.WithSpecification('X', a))
		if err != nil {
			fmt.Println(err)
			return
		}
		p.Format(os.Stdout, ref)
		os.Stdout.Write([]byte{'\n'})
	}

	{
		// I want a completely new specification set, and I want %X to print out my name!
		a := strftime.Verbatim(`Daisuke Maki`)

		ds := strftime.NewSpecificationSet()
		ds.Set('X', a)
		p, err := strftime.New(`%X`, strftime.WithSpecificationSet(ds))
		if err != nil {
			fmt.Println(err)
			return
		}
		p.Format(os.Stdout, ref)
		os.Stdout.Write([]byte{'\n'})
	}

	{
		// I want %s as unix timestamp!
		p, err := strftime.New(`%s`, strftime.WithUnixSeconds('s'))
		if err != nil {
			fmt.Println(err)
			return
		}
		p.Format(os.Stdout, ref)
		os.Stdout.Write([]byte{'\n'})
	}

	// OUTPUT:
	// 123
	// 123
	// Daisuke Maki
	// Daisuke Maki
	// 1136239445
}

func TestGHIssue9(t *testing.T) {
	pattern, _ := strftime.New("/full1/test2/to3/proveIssue9isfixed/11%C22/12345%Y%m%d.%H.log.%C.log")
	testTime := time.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC)
	correctString := "/full1/test2/to3/proveIssue9isfixed/112022/1234520200101.01.log.20.log"

	var buf bytes.Buffer
	pattern.Format(&buf, testTime)

	// Using a fixed time should give us a fixed output.
	if !assert.True(t, buf.String() == correctString) {
		t.Logf("Buffer [%s] should be [%s]", buf.String(), correctString)
		return
	}
}

func TestGHIssue18(t *testing.T) {
	testIH := func(twelveHour bool) func(t *testing.T) {
		var patternString string
		if twelveHour {
			patternString = "%I"
		} else {
			patternString = "%H"
		}
		return func(t *testing.T) {
			t.Helper()
			var buf bytes.Buffer
			pattern, _ := strftime.New(patternString)
			for i := 0; i < 24; i++ {
				testTime := time.Date(2020, 1, 1, i, 1, 1, 1, time.UTC)
				var correctString string
				switch {
				case twelveHour && i == 0:
					correctString = fmt.Sprintf("%02d", 12)
				case twelveHour && i > 12:
					correctString = fmt.Sprintf("%02d", i-12)
				default:
					correctString = fmt.Sprintf("%02d", i)
				}

				buf.Reset()

				pattern.Format(&buf, testTime)
				if !assert.Equal(t, correctString, buf.String(), "Buffer [%s] should be [%s] for time %s", buf.String(), correctString, testTime) {
					return
				}
			}
		}
	}
	testR := func(t *testing.T) {
		t.Helper()
		patternString := "%r"
		var buf bytes.Buffer
		pattern, _ := strftime.New(patternString)
		for i := 0; i < 24; i++ {
			testTime := time.Date(2020, 1, 1, i, 1, 1, 1, time.UTC)

			var correctString string
			switch {
			case i == 0:
				correctString = fmt.Sprintf("%02d:%02d:%02d AM", 12, testTime.Minute(), testTime.Second())
			case i == 12:
				correctString = fmt.Sprintf("%02d:%02d:%02d PM", 12, testTime.Minute(), testTime.Second())
			case i > 12:
				correctString = fmt.Sprintf("%02d:%02d:%02d PM", i-12, testTime.Minute(), testTime.Second())
			default:
				correctString = fmt.Sprintf("%02d:%02d:%02d AM", i, testTime.Minute(), testTime.Second())
			}

			buf.Reset()

			t.Logf("%s", correctString)
			pattern.Format(&buf, testTime)
			if !assert.Equal(t, correctString, buf.String(), "Buffer [%s] should be [%s] for time %s", buf.String(), correctString, testTime) {
				continue
			}
		}
	}
	t.Run("12 hour zero pad %I", testIH(true))
	t.Run("24 hour zero pad %H", testIH(false))
	t.Run("12 hour zero pad %r", testR)
}

func TestFormat12AM(t *testing.T) {
	s, err := strftime.Format(`%H %I %l`, time.Time{})
	if !assert.NoError(t, err, `strftime.Format succeeds`) {
		return
	}

	if !assert.Equal(t, "00 12 12", s, "correctly format the hour") {
		return
	}
}

func TestFormat_WeekNumber(t *testing.T) {
	for y := 2000; y < 2020; y++ {
		sunday := "00"
		monday := "00"
		for d := 1; d < 8; d++ {
			base := time.Date(y, time.January, d, 0, 0, 0, 0, time.UTC)

			switch base.Weekday() {
			case time.Sunday:
				sunday = "01"
			case time.Monday:
				monday = "01"
			}

			if got, _ := strftime.Format("%U", base); got != sunday {
				t.Errorf("Format(%q, %d) = %q, want %q", "%U", base.Unix(), got, sunday)
			}
			if got, _ := strftime.Format("%W", base); got != monday {
				t.Errorf("Format(%q, %d) = %q, want %q", "%W", base.Unix(), got, monday)
			}
		}
	}
}


func TestFormat_WeekYear(t *testing.T) {
    // Note: According to ISO 8601, Week 1 is the week that contains the first Thursday of the year.
    testCases := []struct {
        name     string
        date     time.Time
        expectG  string // %G - ISO week year (4 digits)
        expectg  string // %g - ISO week year without century (2 digits)
    }{
        {
            name:     "New Year's Day 2005 (week belongs to 2004)",
            date:     time.Date(2005, 1, 1, 0, 0, 0, 0, time.UTC), // Saturday
            expectG:  "2004",
            expectg:  "04",
        },
        {
            name:     "New Year's Day 2006 (week belongs to 2005)", 
            date:     time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC), // Sunday
            expectG:  "2005",
            expectg:  "05",
        },
        {
            name:     "Dec 31, 2007 (week belongs to 2008)",
            date:     time.Date(2007, 12, 31, 0, 0, 0, 0, time.UTC), // Monday
            expectG:  "2008",
            expectg:  "08",
        },
        {
            name:     "Dec 30, 2008 (week belongs to 2009)",
            date:     time.Date(2008, 12, 30, 0, 0, 0, 0, time.UTC), // Tuesday
            expectG:  "2009",
            expectg:  "09",
        },
        {
            name:     "Jan 3, 2010 (first week of 2010)",
            date:     time.Date(2010, 1, 3, 0, 0, 0, 0, time.UTC), // Sunday
            expectG:  "2009",
            expectg:  "09",
        },
        {
            name:     "Jan 4, 2010 (first Monday, week 1 of 2010)",
            date:     time.Date(2010, 1, 4, 0, 0, 0, 0, time.UTC), // Monday
            expectG:  "2010",
            expectg:  "10",
        },
        {
            name:     "Regular date mid-year",
            date:     time.Date(2023, 7, 15, 0, 0, 0, 0, time.UTC),
            expectG:  "2023",
            expectg:  "23",
        },
        {
            name:     "Year 2000 (Y2K boundary)",
            date:     time.Date(2000, 6, 15, 0, 0, 0, 0, time.UTC),
            expectG:  "2000",
            expectg:  "00",
        },
        {
            name:     "Year 1999 to 2000 transition",
            date:     time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), // Saturday
            expectG:  "1999",
            expectg:  "99",
        },
        {
            name:     "Single digit year case",
            date:     time.Date(2009, 6, 15, 0, 0, 0, 0, time.UTC),
            expectG:  "2009",
            expectg:  "09",
        },
        {
            name:     "Edge case: Dec 29, 2014 (week 1 of 2015)",
            date:     time.Date(2014, 12, 29, 0, 0, 0, 0, time.UTC), // Monday
            expectG:  "2015",
            expectg:  "15",
        },
        {
            name:     "Edge case: Jan 4, 2021 (first Monday of year)",
            date:     time.Date(2021, 1, 4, 0, 0, 0, 0, time.UTC), // Monday  
            expectG:  "2021",
            expectg:  "21",
        },
        {
            name:     "Edge case: Jan 3, 2021 (belongs to 2020)",
            date:     time.Date(2021, 1, 3, 0, 0, 0, 0, time.UTC), // Sunday
            expectG:  "2020",
            expectg:  "20",
        },
        // Test cases for BCE years
        {
            name:     "Year 1 BCE mid-year",
            date:     time.Date(-1, 6, 15, 0, 0, 0, 0, time.UTC),
            expectG:  "-0001",
            expectg:  "-01",
        },
        {
            name:     "Year 10 BCE",
            date:     time.Date(-1, 3, 20, 0, 0, 0, 0, time.UTC),
            expectG:  "-0001",
            expectg:  "-01",
        },
        {
            name:     "Year 99 BCE",
            date:     time.Date(-99, 8, 10, 0, 0, 0, 0, time.UTC),
            expectG:  "-0099",
            expectg:  "-99",
        },
        {
            name:     "Year 100 BCE (century boundary)",
            date:     time.Date(-100, 12, 25, 0, 0, 0, 0, time.UTC),
            expectG:  "-0100",
            expectg:  "-00",
        },
        {
            name:     "Year 1234 BCE (4-digit BCE year)",
            date:     time.Date(-1234, 9, 15, 0, 0, 0, 0, time.UTC),
            expectG:  "-1234",
            expectg:  "-34",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test %G (4-digit ISO week year)
            gotG, err := strftime.Format("%G", tc.date)
            if !assert.NoError(t, err, "strftime.Format %%G should succeed") {
                return
            }
            if !assert.Equal(t, tc.expectG, gotG, "%%G format for %v", tc.date) {
                return
            }

            // Test %g (2-digit ISO week year)
            gotg, err := strftime.Format("%g", tc.date)
            if !assert.NoError(t, err, "strftime.Format %%g should succeed") {
                return
            }
            if !assert.Equal(t, tc.expectg, gotg, "%%g format for %v", tc.date) {
                return
            }

            // Verify consistency: last 2 digits of %G should equal %g
            expectedLastG := tc.expectG[len(tc.expectG)-2:]
            expectedLastg := tc.expectg[len(tc.expectg)-2:]
            if !assert.Equal(t, expectedLastG, expectedLastg, "%%g should be last 2 digits of %%G for %v", tc.date) {
                return
            }
        })
    }
}

func TestFormat_WeekYearBoundaries(t *testing.T) {
    // Test the tricky boundary cases where calendar year != ISO week year
    boundaryTests := []struct {
        calendarYear int
        month        time.Month
        day          int
        expectedWeekYear int
    }{
        // Cases where early January belongs to previous week year
        {2005, time.January, 1, 2004}, // Sat Jan 1, 2005 is in week 53 of 2004
        {2005, time.January, 2, 2004}, // Sun Jan 2, 2005 is in week 53 of 2004
        {2006, time.January, 1, 2005}, // Sun Jan 1, 2006 is in week 52 of 2005
        
        // Cases where late December belongs to next week year  
        {2007, time.December, 31, 2008}, // Mon Dec 31, 2007 is in week 1 of 2008
        {2008, time.December, 29, 2009}, // Mon Dec 29, 2008 is in week 1 of 2009
        {2008, time.December, 30, 2009}, // Tue Dec 30, 2008 is in week 1 of 2009
        {2008, time.December, 31, 2009}, // Wed Dec 31, 2008 is in week 1 of 2009
    }

    for _, bt := range boundaryTests {
        testDate := time.Date(bt.calendarYear, bt.month, bt.day, 0, 0, 0, 0, time.UTC)
        
        gotG, err := strftime.Format("%G", testDate)
        if !assert.NoError(t, err, "strftime.Format %%G should succeed") {
            continue
        }
        
        expectedG := fmt.Sprintf("%04d", bt.expectedWeekYear)
        if !assert.Equal(t, expectedG, gotG, "Calendar %d-%02d-%02d should be week year %d", 
            bt.calendarYear, bt.month, bt.day, bt.expectedWeekYear) {
            continue
        }
        
        // Test %g as well
        gotg, err := strftime.Format("%g", testDate)
        if !assert.NoError(t, err, "strftime.Format %%g should succeed") {
            continue
        }
        
        expectedg := fmt.Sprintf("%02d", bt.expectedWeekYear%100)
        assert.Equal(t, expectedg, gotg, "Week year without century should match for %v", testDate)
    }
}
