package elastic

import (
	"encoding/json"
	"encoding/xml"
	"log/slog"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
)

var (
	_ option.Equality[Elastic[any]] = Elastic[any]{}
	_ option.Cloner[Elastic[any]]   = Elastic[any]{}
	_ json.Marshaler                = Elastic[any]{}
	_ json.Unmarshaler              = (*Elastic[any])(nil)
	_ jsonv2.MarshalerV2            = Elastic[any]{}
	_ xml.Marshaler                 = Elastic[any]{}
	_ xml.Unmarshaler               = (*Elastic[any])(nil)
	// We don't implement UnmarshalJSONV2 since there's variants that cannot be unmarshaled without
	// calling unmarshal twice or so.
	// there's 4 possible code paths
	//
	//   - input is T
	//   - input is []T
	//   - input starts with [ but T is []U
	//   - input starts with [ but T implements UnmarshalJSON v1 or v2; it's ambiguous.
	//
	// That'll needs unnecessary complexity to code base, e.g. teeing tokens and token stream decoder.
	//
	// _ jsonv2.UnmarshalerV2          = (*Elastic[any])(nil)
	_ slog.LogValuer = Elastic[any]{}
)

// Elastic[T] is a type that can express undefined | null | T | [](null | T).
//
// Elastic[T] can be a skippable struct field with omitempty option of `encoding/json`.
//
// Although it exposes its internal data structure,
// you should not mutate internal data.
// For more detail,
// See doc comment for github.com/ngicks/und/sliceund.Und[T].
type Elastic[T any] sliceund.Und[option.Options[T]]

// Null returns a null Elastic[T].
func Null[T any]() Elastic[T] {
	return Elastic[T](sliceund.Null[option.Options[T]]())
}

// Undefined returns an undefined Elastic[T].
func Undefined[T any]() Elastic[T] {
	return Elastic[T](sliceund.Undefined[option.Options[T]]())
}

// FromOptions converts slice of option.Option[T] into Elastic[T].
// options is retained by the returned value.
func FromOptions[T any, Opts ~[]option.Option[T]](options Opts) Elastic[T] {
	return Elastic[T](sliceund.Defined(option.Options[T](options)))
}

func (e Elastic[T]) inner() sliceund.Und[option.Options[T]] {
	return sliceund.Und[option.Options[T]](e)
}

// IsDefined returns true if e is a defined Elastic[T],
// which includes a slice with no element.
func (e Elastic[T]) IsDefined() bool {
	return e.inner().IsDefined()
}

// IsNull returns true if e is a null Elastic[T].
func (e Elastic[T]) IsNull() bool {
	return e.inner().IsNull()
}

// IsUndefined returns true if e is an undefined Elastic[T].
func (e Elastic[T]) IsUndefined() bool {
	return e.inner().IsUndefined()
}

// Equal implements option.Equality[Elastic[T]].
//
// Equal panics if T is uncomparable.
func (e Elastic[T]) Equal(other Elastic[T]) bool {
	return e.inner().Equal(other.inner())
}

// Clone implements option.Cloner[Elastic[T]].
//
// Clone clones its internal option.Option slice by copy.
// Or if T implements Cloner[T], each element is cloned.
func (e Elastic[T]) Clone() Elastic[T] {
	return Elastic[T](e.inner().Clone())
}

// Value returns a first value of its internal option slice if e is defined.
// Otherwise it returns zero value for T.
func (e Elastic[T]) Value() T {
	if e.IsDefined() {
		vs := e.inner().Value()
		if len(vs) > 0 {
			return vs[0].Value()
		}
	}
	var zero T
	return zero
}

// Values returns internal option slice as plain []T.
//
// If e is not defined, it returns nil.
// Any None value in its internal option slice will be converted
// to zero value of T.
func (e Elastic[T]) Values() []T {
	if !e.IsDefined() {
		return []T(nil)
	}
	opts := e.inner().Value()
	vs := make([]T, len(opts))
	for i, opt := range opts {
		vs[i] = opt.Value()
	}
	return vs
}

// Pointer returns a first value of its internal option slice as *T if e is defined.
//
// Pointer returns nil if
//   - e is not defined
//   - e has no element
//   - e's first element is None.
func (e Elastic[T]) Pointer() *T {
	if e.IsDefined() {
		vs := e.inner().Value()
		if len(vs) > 0 && vs[0].IsSome() {
			v := vs[0].Value()
			return &v
		}
	}
	return nil
}

// Pointer returns its internal option slice as []*T if e is defined.
func (e Elastic[T]) Pointers() []*T {
	if !e.IsDefined() {
		return nil
	}
	opts := e.inner().Value()
	ptrs := make([]*T, len(opts))
	for i, opt := range opts {
		ptrs[i] = opt.Pointer()
	}
	return ptrs
}

// Unwrap unwraps e.
func (u Elastic[T]) Unwrap() sliceund.Und[option.Options[T]] {
	return u.inner()
}

// Map returns a new Elastic[T] whose internal value is e's mapped by f.
//
// The internal slice of e is capped to its length before passed to f.
func (e Elastic[T]) Map(f func(sliceund.Und[option.Options[T]]) sliceund.Und[option.Options[T]]) Elastic[T] {
	return Elastic[T](
		f(e.inner().Map(func(o option.Option[option.Option[option.Options[T]]]) option.Option[option.Option[option.Options[T]]] {
			if !o.IsNone() {
				return o
			}
			v := o.Value()
			if v.IsNone() {
				return o
			}
			vv := v.Value()
			return option.Some(option.Some(vv[:len(vv):len(vv)]))
		})),
	)
}

// UnmarshalXML implements xml.Unmarshaler.
func (o *Elastic[T]) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var t option.Options[T]
	err := d.DecodeElement(&t, &start)
	if err != nil {
		return err
	}

	if len(o.inner().Value()) == 0 {
		*o = FromOptions(t)
	} else {
		*o = o.Map(func(u sliceund.Und[option.Options[T]]) sliceund.Und[option.Options[T]] {
			return u.Map(func(o option.Option[option.Option[option.Options[T]]]) option.Option[option.Option[option.Options[T]]] {
				return o.Map(func(v option.Option[option.Options[T]]) option.Option[option.Options[T]] {
					return v.Map(func(v option.Options[T]) option.Options[T] {
						return append(v, t...)
					})
				})
			})
		})
	}
	return nil
}
