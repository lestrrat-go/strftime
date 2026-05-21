package strftime_test

import (
	"testing"
	"time"

	"github.com/lestrrat-go/strftime"
	"github.com/stretchr/testify/require"
)

// unpadref is chosen so that every numeric field has a single significant
// digit, making the effect of the '-'/'#' (no-pad) flag observable.
// 2006-01-02 03:04:05 UTC: Jan(01) 2nd(02) 03:04:05, year-of-century 06,
// day-of-year 002, UTC offset +0000.
var unpadref = time.Date(2006, time.January, 2, 3, 4, 5, 0, time.UTC)

func TestUnpaddedFlag(t *testing.T) {
	for _, tc := range []struct {
		spec     byte
		padded   string
		unpadded string
	}{
		{'d', "02", "2"},
		{'e', " 2", "2"},
		{'H', "03", "3"},
		{'I', "03", "3"},
		{'j', "002", "2"},
		{'k', " 3", "3"},
		{'l', " 3", "3"},
		{'M', "04", "4"},
		{'m', "01", "1"},
		{'S', "05", "5"},
		{'y', "06", "6"},
		{'z', "+0000", "+0"},
	} {
		t.Run(string(tc.spec), func(t *testing.T) {
			padded, err := strftime.Format("%"+string(tc.spec), unpadref)
			require.NoError(t, err, "padded form should compile")
			require.Equal(t, tc.padded, padded, "padded baseline")

			for _, flag := range []string{"-", "#"} {
				got, err := strftime.Format("%"+flag+string(tc.spec), unpadref)
				require.NoErrorf(t, err, "%%%s%c should compile", flag, tc.spec)
				require.Equalf(t, tc.unpadded, got, "%%%s%c", flag, tc.spec)
			}
		})
	}
}

func TestUnpaddedNonNumericUnchanged(t *testing.T) {
	// Non-numeric fields are emitted unchanged when the flag is present.
	for _, spec := range []string{"A", "B", "Z", "p"} {
		plain, err := strftime.Format("%"+spec, unpadref)
		require.NoError(t, err)
		flagged, err := strftime.Format("%-"+spec, unpadref)
		require.NoError(t, err)
		require.Equalf(t, plain, flagged, "%%-%s should match %%%s", spec, spec)
	}
}

func TestUnpaddedInContext(t *testing.T) {
	got, err := strftime.Format("%Y-%-m-%-d %-H:%-M:%-S", unpadref)
	require.NoError(t, err)
	require.Equal(t, "2006-1-2 3:4:5", got)
}

func TestUnpaddedStrayFlag(t *testing.T) {
	for _, p := range []string{"%-", "%#", "foo %-"} {
		_, err := strftime.Format(p, unpadref)
		require.Errorf(t, err, "pattern %q should error", p)
	}
}
