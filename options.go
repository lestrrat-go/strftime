package strftime

type Option interface {
	Name() string
	Value() any
}

type option struct {
	name  string
	value any
}

func (o *option) Name() string { return o.name }
func (o *option) Value() any   { return o.value }

const optSpecificationSet = `opt-specification-set`

// WithSpecificationSet allows you to specify a custom specification set
func WithSpecificationSet(ds SpecificationSet) Option {
	return &option{
		name:  optSpecificationSet,
		value: ds,
	}
}

const optLocale = `opt-locale`

// WithLocale overrides the name-producing conversion specifiers (%A, %a, %B,
// %b, %h, %p) with the localized names supplied by loc, which is typically
// built with NewLocale. Numeric specifiers (%d, %m, %Y, ...) are
// locale-invariant and are unaffected.
//
// Because a Locale supplies one form per name, format a context that needs
// inflected month names (e.g. "%d %B" vs "%B %Y" in Slavic languages) by
// compiling a separate Strftime object per context, each with the Locale
// holding the appropriate form. See the Locale documentation for details.
func WithLocale(loc Locale) Option {
	return &option{
		name:  optLocale,
		value: loc,
	}
}

type optSpecificationPair struct {
	name     byte
	appender Appender
}

const optSpecification = `opt-specification`

// WithSpecification allows you to create a new specification set on the fly,
// to be used only for that invocation.
func WithSpecification(b byte, a Appender) Option {
	return &option{
		name: optSpecification,
		value: &optSpecificationPair{
			name:     b,
			appender: a,
		},
	}
}

// WithMilliseconds is similar to WithSpecification, and specifies that
// the Strftime object should interpret the pattern `%b` (where b
// is the byte that you specify as the argument)
// as the zero-padded, 3 letter milliseconds of the time.
func WithMilliseconds(b byte) Option {
	return WithSpecification(b, Milliseconds())
}

// WithMicroseconds is similar to WithSpecification, and specifies that
// the Strftime object should interpret the pattern `%b` (where b
// is the byte that you specify as the argument)
// as the zero-padded, 3 letter microseconds of the time.
func WithMicroseconds(b byte) Option {
	return WithSpecification(b, Microseconds())
}

// WithUnixSeconds is similar to WithSpecification, and specifies that
// the Strftime object should interpret the pattern `%b` (where b
// is the byte that you specify as the argument)
// as the unix timestamp in seconds
func WithUnixSeconds(b byte) Option {
	return WithSpecification(b, UnixSeconds())
}
