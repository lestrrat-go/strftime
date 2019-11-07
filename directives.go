package strftime

import (
	"sync"

	"github.com/pkg/errors"
)

// because there is no such thing was a sync.RWLocker
type rwLocker interface {
	RLock()
	RUnlock()
	sync.Locker
}

// DirectiveSet is a container for patterns that Strftime uses.
// If you want a custom strftime, you can copy the default
// DirectiveSet and tweak it
type DirectiveSet interface {
	Lookup(byte) (Appender, error)
	Delete(byte) error
	Set(byte, Appender) error
}

type directiveSet struct {
	lock  rwLocker
	store map[byte]Appender
}

type immutableDirectiveSet struct {
	store map[byte]Appender
}

// The default directive set does not need any locking as it is never
// accessed from the outside, and is never mutated.
var defaultDirectiveSet *immutableDirectiveSet

func init() {
	tmp := NewDirectiveSet()
	populateDefaultDirectives(tmp)
	
	defaultDirectiveSet = &immutableDirectiveSet{
		store: tmp.(*directiveSet).store,
	}
}

// NewDirectiveSet creates a directive set with the default directives
func NewDirectiveSet() DirectiveSet {
	ds := &directiveSet{
		lock:  &sync.RWMutex{},
		store: make(map[byte]Appender),
	}
	populateDefaultDirectives(ds)

	return ds
}

func populateDefaultDirectives(ds DirectiveSet) {
	ds.Set('A', fullWeekDayName)
	ds.Set('a', abbrvWeekDayName)
	ds.Set('B', fullMonthName)
	ds.Set('b', abbrvMonthName)
	ds.Set('h', abbrvMonthName)
	ds.Set('C', centuryDecimal)
	ds.Set('c', timeAndDate)
	ds.Set('D', mdy)
	ds.Set('d', dayOfMonthZeroPad)
	ds.Set('e', dayOfMonthSpacePad)
	ds.Set('F', ymd)
	ds.Set('H', twentyFourHourClockZeroPad)
	ds.Set('I', twelveHourClockZeroPad)
	ds.Set('j', dayOfYear)
	ds.Set('k', twentyFourHourClockSpacePad)
	ds.Set('l', twelveHourClockSpacePad)
	ds.Set('M', minutesZeroPad)
	ds.Set('m', monthNumberZeroPad)
	ds.Set('n', newline)
	ds.Set('p', ampm)
	ds.Set('R', hm)
	ds.Set('r', imsp)
	ds.Set('S', secondsNumberZeroPad)
	ds.Set('T', hms)
	ds.Set('t', tab)
	ds.Set('U', weekNumberSundayOrigin)
	ds.Set('u', weekdayMondayOrigin)
	ds.Set('V', weekNumberMondayOriginOneOrigin)
	ds.Set('v', eby)
	ds.Set('W', weekNumberMondayOrigin)
	ds.Set('w', weekdaySundayOrigin)
	ds.Set('X', natReprTime)
	ds.Set('x', natReprDate)
	ds.Set('Y', year)
	ds.Set('y', yearNoCentury)
	ds.Set('Z', timezone)
	ds.Set('z', timezoneOffset)
	ds.Set('%', percent)
}

func (ds *immutableDirectiveSet) Lookup(b byte) (Appender, error) {
	v, ok := ds.store[b]
	if !ok {
		return nil, errors.Errorf(`lookup failed: pattern %%%c was not found in directive set`, b)
	}
	return v, nil
}

func (ds *immutableDirectiveSet) Set(_ byte, _ Appender) error {
	return errors.New(`set failed: directive set is immutable`)
}

func (ds *immutableDirectiveSet) Delete(_ byte) error {
	return errors.New(`delete failed: directive set is immutable`)
}

func (ds *directiveSet) Lookup(b byte) (Appender, error) {
	ds.lock.RLock()
	defer ds.lock.RLock()
	v, ok := ds.store[b]
	if !ok {
		return nil, errors.Errorf(`lookup failed: pattern %%%c was not found in directive set`, b)
	}
	return v, nil
}

func (ds *directiveSet) Delete(b byte) error {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	delete(ds.store, b)
	return nil
}

func (ds *directiveSet) Set(b byte, a Appender) error {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	ds.store[b] = a
	return nil

}
