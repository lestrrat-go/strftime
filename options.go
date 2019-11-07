package strftime

type Option interface {
	Name() string
	Value() interface{}
}

type option struct {
	name  string
	value interface{}
}

func (o *option) Name() string       { return o.name }
func (o *option) Value() interface{} { return o.value }

const optDirectiveSet = `opt-directive-set`

// WithDirective allows you to specify a custom directive set
func WithDirectiveSet(ds DirectiveSet) Option {
	return &option{
		name:  optDirectiveSet,
		value: ds,
	}
}

type optDirectivePair struct {
	name     byte
	appender Appender
}

const optDirective = `opt-directive`

// WithDirective allows you to create a new directive set on the fly,
// to be used only for that invocation.
func WithDirective(b byte, a Appender) Option {
	return &option{
		name: optDirective,
		value: &optDirectivePair{
			name:     b,
			appender: a,
		},
	}
}

// WithMilliseconds is similar to WithDirective, and specifies that
// the Strftime object should interpret the pattern `%b` (where b
// is the byte that you specify as the argument)
// as the zero-padded, 3 letter milliseconds of the time.
func WithMilliseconds(b byte) Option {
	return WithDirective(b, Milliseconds)
}
