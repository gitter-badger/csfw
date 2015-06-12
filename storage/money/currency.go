// Copyright 2015, Cyrill @ Schumacher.fm and the CoreStore contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package money

import (
	"errors"
	"math"

	"github.com/corestoreio/csfw/i18n"
	"github.com/corestoreio/csfw/utils/log"
)

var (
	// ErrOverflow occurs on integer overflow
	ErrOverflow = errors.New("Integer Overflow")

	RoundTo = .5
	//	RoundTo  = .5 + (1 / Guardf)
	RoundToN = RoundTo * -1
)

// Interval* constants http://en.wikipedia.org/wiki/Swedish_rounding
const (
	// Interval000 no swedish rounding (default)
	Interval000 Interval = iota
	// Interval005 rounding with 0.05 intervals
	Interval005
	// Interval010 rounding with 0.10 intervals
	Interval010
	// Interval015 same as Interval010 except that 5 will be rounded down.
	// 0.45 => 0.40 or 0.46 => 0.50
	// Special case for New Zealand (a must visit!), it is up to the business
	// to decide if they will round 5¢ intervals up or down. The majority of
	// retailers follow government advice and round it down. Use then Interval015.
	// otherwise use Interval010.
	Interval015
	// Interval025 rounding with 0.25 intervals
	Interval025
	// Interval050 rounding with 0.50 intervals
	Interval050
	// Interval100 rounding with 1.00 intervals
	Interval100
	interval999
)

type (
	// Interval defines the type for the Swedish rounding.
	Interval uint8

	// Currency represents a money aka currency type to avoid rounding errors with floats.
	// Takes also care of http://en.wikipedia.org/wiki/Swedish_rounding
	Currency struct {
		// m money in Guard/DP
		m int64
		// fmt to allow language specific output formats
		fmt i18n.CurrencyFormatter
		// Valid if false the internal value is NULL
		Valid bool
		// Interval defines how the swedish rounding can be applied.
		Interval Interval

		jm  JSONMarshaller
		jum JSONUnmarshaller

		guard  int64
		guardf float64
		dp     int64
		dpf    float64
		// bufC print buffer for number generation incl. locale settings ... or a sync.Pool ?
		bufC buf
	}

	// OptionFunc used to apply options to the Currency struct
	OptionFunc func(*Currency) OptionFunc
)

// Swedish sets the Swedish rounding
// http://en.wikipedia.org/wiki/Swedish_rounding
// Errors will be logged
func Swedish(i Interval) OptionFunc {
	if i >= interval999 {
		log.Error("Currency=SetSwedishRounding", "err", errors.New("Interval out of scope. Resetting."), "interval", i)
		i = Interval000
	}
	return func(c *Currency) OptionFunc {
		previous := c.Interval
		c.Interval = i
		return Swedish(previous)
	}
}

// SetGuard sets the guard
func Guard(g int) OptionFunc {
	if g == 0 {
		g = 1
	}
	return func(c *Currency) OptionFunc {
		previous := int(c.guard)
		c.guard = int64(g)
		c.guardf = float64(g)
		return Guard(previous)
	}
}

// Precision sets the precision.
// 2 decimal places => 10^2; 3 decimal places => 10^3; x decimal places => 10^x
// If not a decimal power then falls back to the default value.
func Precision(p int) OptionFunc {
	p64 := int64(p)
	l := int64(math.Log(float64(p64)))
	if p64 != 0 && (l%2) != 0 {
		p64 = dp
	}
	if p64 == 0 { // check for division by zero
		p64 = 1
	}
	return func(c *Currency) OptionFunc {
		previous := int(c.dp)
		c.dp = p64
		c.dpf = float64(p64)
		return Precision(previous)
	}
}

// Format sets the locale specific formatter. Panics if f is nil.
func Format(f i18n.CurrencyFormatter) OptionFunc {
	// @todo not sure if this function is needed an we simply can export field
	// fmt as Formatter ... but what if we need mutexes?
	if f == nil {
		log.Fatal("money=Format", "err", errors.New("Argument f cannot be nil"))
	}
	return func(c *Currency) OptionFunc {
		previous := c.fmt
		c.fmt = f
		return Format(previous)
	}
}

// JSONMarshal sets a custom JSON Marshaller
func JSONMarshal(m JSONMarshaller) OptionFunc {
	// @todo not sure if this function is needed an we simply can export field
	// jm as JSONMarshaller ... but what if we need mutexes?
	if m == nil {
		m = NewJSONEncoder()
	}
	return func(c *Currency) OptionFunc {
		previous := c.jm
		c.jm = m
		return JSONMarshal(previous)
	}
}

// JSONUnmarshal sets a custom JSON Unmmarshaller
func JSONUnmarshal(um JSONUnmarshaller) OptionFunc {
	// @todo not sure if this function is needed an we simply can export field
	// jum as JSONUnmarshaller ... but what if we need mutexes?
	if um == nil {
		um = NewJSONDecoder()
	}
	return func(c *Currency) OptionFunc {
		previous := c.jum
		c.jum = um
		return JSONUnmarshal(previous)
	}
}

// New creates a new empty Currency struct with package default values of
// Guard and decimal precision.
func New(opts ...OptionFunc) Currency {
	c := Currency{
		guard:  guard,
		guardf: guardf,
		dp:     dp,
		dpf:    dpf,
		fmt:    DefaultFormat,
		jm:     DefaultJSONEncode,
		jum:    DefaultJSONDecode,
	}
	c.Option(opts...)
	return c
}

// Options besides New() also Option() can apply options to the current
// struct. It returns the last set option. More info about the returned function:
// http://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
func (c *Currency) Option(opts ...OptionFunc) (previous OptionFunc) {
	for _, o := range opts {
		if o != nil {
			previous = o(c)
		}
	}
	return previous
}

// Abs Returns the absolute value of Currency
func (c Currency) Abs() Currency {
	if c.m < 0 {
		return c.Neg()
	}
	return c
}

// Getf gets the float64 value of money (see Raw() for int64)
func (c Currency) Getf() float64 {
	return float64(c.m) / c.dpf
}

// Geti gets value of money truncating after decimal precision (see Raw() for no truncation).
// Rounds always down
func (c Currency) Geti() int64 {
	return c.m / c.dp
}

// Dec returns the decimals
func (c Currency) Dec() int64 {
	return c.Abs().Raw() % c.dp
}

// Raw returns in int64 the value of Currency (also see Gett(), See Get() for float64)
func (c Currency) Raw() int64 {
	return c.m
}

// Set sets the raw Currency field m
func (c Currency) Set(i int64) Currency {
	c.m = i
	c.Valid = true
	return c
}

// Setf sets a float64 into a Currency type for precision calculations
func (c Currency) Setf(f float64) Currency {
	fDPf := f * c.dpf
	r := int64(f * c.dpf)
	c.Valid = true
	return c.Set(rnd(r, fDPf-float64(r)))
}

// Sign returns the Sign of Currency 1 if positive, -1 if negative
func (c Currency) Sign() int {
	if c.m < 0 {
		return -1
	}
	return 1
}

// Localize for money type representation in a specific locale. Owns the return value.
func (c Currency) Localize() []byte {
	// thread safe?
	c.bufC = c.bufC[:0]
	c.fmt.FmtCurrency(&c.bufC, c.Sign(), c.Geti(), c.Dec())
	return c.bufC
}

// String for money type representation i a specific locale.
func (c Currency) String() string {
	return string(c.Localize())

}

// Number prints the currency without any locale specific formatting. E.g. useful in JavaScript.
func (c Currency) Number() string {
	return string(c.NumberByte())
}

// NumberByte prints the currency without any locale specific formatting. Owns the result.
func (c Currency) NumberByte() []byte {
	// thread safe?
	c.bufC = c.bufC[:0]
	c.fmt.FmtNumber(&c.bufC, c.Sign(), c.Geti(), c.Dec())
	return c.bufC
}

// Add Adds two Currency types. Returns empty Currency on integer overflow.
// Errors will be logged and a trace is available when the level for tracing has been set.
func (c Currency) Add(d Currency) Currency {
	r := c.m + d.m
	if (r^c.m)&(r^d.m) < 0 {
		if log.IsTrace() {
			log.Trace("Currency=Add", "err", ErrOverflow, "m", c, "n", d)
		}
		log.Error("Currency=Add", "err", ErrOverflow, "m", c, "n", d)
		return New()
	}
	c.m = r
	c.Valid = true
	return c
}

// Sub subtracts one Currency type from another. Returns empty Currency on integer overflow.
// Errors will be logged and a trace is available when the level for tracing has been set.
func (c Currency) Sub(d Currency) Currency {
	r := c.m - d.m
	if (r^c.m)&^(r^d.m) < 0 {
		if log.IsTrace() {
			log.Trace("Currency=Sub", "err", ErrOverflow, "m", c, "n", d)
		}
		log.Error("Currency=Sub", "err", ErrOverflow, "m", c, "n", d)
		return New()
	}
	c.m = r
	return c
}

// Mul Multiplies two Currency types. Both types must have the same precision.
func (c Currency) Mul(d Currency) Currency {
	return c.Set(c.m * d.m / c.dp)
}

// Div Divides one Currency type from another
func (c Currency) Div(d Currency) Currency {
	f := (c.guardf * c.dpf * float64(c.m)) / float64(d.m) / c.guardf
	i := int64(f)
	return c.Set(rnd(i, f-float64(i)))
}

// Mulf Multiplies a Currency with a float to return a money-stored type
func (c Currency) Mulf(f float64) Currency {
	i := c.m * int64(f*c.guardf*c.dpf)
	r := i / c.guard / c.dp
	return c.Set(rnd(r, float64(i)/c.guardf/c.dpf-float64(r)))
}

// Neg Returns the negative value of Currency
func (c Currency) Neg() Currency {
	if c.m != 0 {
		c.m *= -1
	}
	return c
}

// Pow is the power of Currency
func (c Currency) Pow(f float64) Currency {
	return c.Setf(math.Pow(c.Getf(), f))
}

// rnd rounds int64 remainder rounded half towards plus infinity
// trunc = the remainder of the float64 calc
// r     = the result of the int64 cal
func rnd(r int64, trunc float64) int64 {
	//fmt.Printf("RND 1 r = % v, trunc = %v RoundTo = %v\n", r, trunc, RoundTo)
	if trunc > 0 {
		if trunc >= RoundTo {
			r++
		}
	} else {
		if trunc < RoundToN {
			r--
		}
	}
	//fmt.Printf("RND 2 r = % v, trunc = %v RoundTo = %v\n", r, trunc, RoundTo)
	return r
}

// Round rounds a value.
func Round(f float64) float64 {
	// @todo check out to round negative numbers https://gist.github.com/pelegm/c48cff315cd223f7cf7b
	return math.Floor(f + .5)
}

// Swedish applies the Swedish rounding. You may set the usual options.
func (c Currency) Swedish(opts ...OptionFunc) Currency {
	c.Option(opts...)
	switch c.Interval {
	case Interval005:
		// NL, SG, SA, CH, TR, CL, IE
		// 5 cent rounding
		return c.Setf(Round(c.Getf()*20) / 20) // base 5
	case Interval010:
		// New Zealand & Hong Kong
		// 10 cent rounding
		// In Sweden between 1985 and 1992, prices were rounded up for sales
		// ending in 5 öre.
		return c.Setf(Round(c.Getf()*10) / 10)
	case Interval015:
		// 10 cent rounding, special case
		// Special case: In NZ, it is up to the business to decide if they
		// will round 5¢ intervals up or down. The majority of retailers follow
		// government advice and round it down.
		if c.m%5 == 0 {
			c.m = c.m - 1
		}
		return c.Setf(Round(c.Getf()*10) / 10)
	case Interval025:
		// round to quarter
		return c.Setf(Round(c.Getf()*4) / 4)
	case Interval050:
		// 50 cent rounding
		// The system used in Sweden from 1992 to 2010, in Norway from 1993 to 2012,
		// and in Denmark since 1 October 2008 is the following:
		// Sales ending in 1–24 öre round down to 0 öre.
		// Sales ending in 25–49 öre round up to 50 öre.
		// Sales ending in 51–74 öre round down to 50 öre.
		// Sales ending in 75–99 öre round up to the next whole Krone/krona.
		return c.Setf(Round(c.Getf()*2) / 2)
	case Interval100:
		// The system used in Sweden since 30 September 2010 and used in Norway since 1 May 2012.
		// Sales ending in 1–49 öre/øre round down to 0 öre/øre.
		// Sales ending in 50–99 öre/øre round up to the next whole krona/krone.
		return c.Setf(Round(c.Getf()*1) / 1) // ;-)
	}
	return c
}
