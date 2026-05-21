package strftime_test

import (
	"testing"
	"time"

	"github.com/lestrrat-go/strftime"
	"github.com/stretchr/testify/require"
)

// french is a representative non-inflecting locale.
var french = strftime.NewLocale(
	strftime.WithMonths(strftime.MonthNames{
		"janvier", "février", "mars", "avril", "mai", "juin",
		"juillet", "août", "septembre", "octobre", "novembre", "décembre",
	}),
	strftime.WithShortMonths(strftime.MonthNames{
		"janv.", "févr.", "mars", "avr.", "mai", "juin",
		"juil.", "août", "sept.", "oct.", "nov.", "déc.",
	}),
	strftime.WithWeekdays(strftime.WeekdayNames{
		"dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi",
	}),
	strftime.WithShortWeekdays(strftime.WeekdayNames{
		"dim.", "lun.", "mar.", "mer.", "jeu.", "ven.", "sam.",
	}),
)

// 2006-01-02 03:04:05 UTC is a Monday in January, before noon.
var locref = time.Date(2006, time.January, 2, 3, 4, 5, 0, time.UTC)

func TestLocaleFrench(t *testing.T) {
	for _, tc := range []struct {
		pattern string
		want    string
	}{
		{"%A", "lundi"},
		{"%a", "lun."},
		{"%B", "janvier"},
		{"%b", "janv."},
		{"%h", "janv."},
		{"%A %d %B %Y", "lundi 02 janvier 2006"},
	} {
		t.Run(tc.pattern, func(t *testing.T) {
			got, err := strftime.Format(tc.pattern, locref, strftime.WithLocale(french))
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

// A Locale that sets only weekday names must still produce English month
// names rather than blanks.
func TestLocalePartialFallback(t *testing.T) {
	partial := strftime.NewLocale(
		strftime.WithWeekdays(strftime.WeekdayNames{
			"dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi",
		}),
	)
	got, err := strftime.Format("%A %B", locref, strftime.WithLocale(partial))
	require.NoError(t, err)
	require.Equal(t, "lundi January", got)
}

// Numeric specifiers are locale-invariant.
func TestLocaleNumericUnaffected(t *testing.T) {
	got, err := strftime.Format("%Y-%m-%d %H:%M:%S", locref, strftime.WithLocale(french))
	require.NoError(t, err)
	require.Equal(t, "2006-01-02 03:04:05", got)
}

// DefaultLocale reproduces the stock English output.
func TestLocaleDefaultMatchesEnglish(t *testing.T) {
	const pattern = "%A %a %B %b %p"
	english, err := strftime.Format(pattern, locref)
	require.NoError(t, err)
	withDefault, err := strftime.Format(pattern, locref, strftime.WithLocale(strftime.DefaultLocale()))
	require.NoError(t, err)
	require.Equal(t, english, withDefault)
}

func TestLocaleAMPM(t *testing.T) {
	loc := strftime.NewLocale(strftime.WithMeridiem("matin", "soir"))
	morning, err := strftime.Format("%p", locref, strftime.WithLocale(loc)) // 03:04 -> morning
	require.NoError(t, err)
	require.Equal(t, "matin", morning)

	afternoon := time.Date(2006, time.January, 2, 15, 0, 0, 0, time.UTC)
	evening, err := strftime.Format("%p", afternoon, strftime.WithLocale(loc))
	require.NoError(t, err)
	require.Equal(t, "soir", evening)
}

// WithSpecification applied alongside WithLocale must win.
func TestLocaleExplicitSpecificationOverrides(t *testing.T) {
	got, err := strftime.Format("%B", locref,
		strftime.WithLocale(french),
		strftime.WithSpecification('B', strftime.Verbatim("OVERRIDE")),
	)
	require.NoError(t, err)
	require.Equal(t, "OVERRIDE", got)
}

// Inflecting languages: the correct in-date vs stand-alone form is selected
// by compiling one Strftime per context, each with the matching Locale.
func TestLocaleInflectionViaObjectSwap(t *testing.T) {
	weekdays := strftime.WithWeekdays(strftime.WeekdayNames{
		"воскресенье", "понедельник", "вторник", "среда",
		"четверг", "пятница", "суббота",
	})

	inDate := strftime.NewLocale(weekdays, strftime.WithMonths(strftime.MonthNames{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}))

	standalone := strftime.NewLocale(weekdays, strftime.WithMonths(strftime.MonthNames{
		"январь", "февраль", "март", "апрель", "май", "июнь",
		"июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь",
	}))

	dateFmt, err := strftime.New("%d %B %Y", strftime.WithLocale(inDate))
	require.NoError(t, err)
	require.Equal(t, "02 января 2006", dateFmt.FormatString(locref))

	headerFmt, err := strftime.New("%B %Y", strftime.WithLocale(standalone))
	require.NoError(t, err)
	require.Equal(t, "январь 2006", headerFmt.FormatString(locref))
}
