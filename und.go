package und

import (
	"encoding/json"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/ngicks/und/option"
)

var (
	_ option.Equality[Und[any]] = Und[any]{}
	_ json.Marshaler            = Und[any]{}
	_ json.Unmarshaler          = (*Und[any])(nil)
	_ jsonv2.MarshalerV2        = Und[any]{}
	_ jsonv2.UnmarshalerV2      = (*Und[any])(nil)
)

// Und[T] is a type that can express a value (`T`), empty (`null`), or absent (`undefined`).
// Und[T] is comparable if T is comparable. And it can be copied by assign.
//
// When using Und[T] as a struct field,
// it can be skipped while marshaling if
//   - the field is `undefined`
//   - marshaler is "github.com/go-json-experiment/json/jsontext"
//   - omitzero options is set.
//
// If you need to stick with encoding/json v1, you can use github.com/ngicks/und/sliceund,
// a slice based version of Und[T] whish is already skppable by v1.
type Und[T any] struct {
	opt option.Option[option.Option[T]]
}

func Defined[T any](t T) Und[T] {
	return Und[T]{
		opt: option.Some(option.Some(t)),
	}
}

func Null[T any]() Und[T] {
	return Und[T]{
		opt: option.Some(option.None[T]()),
	}
}

func Undefined[T any]() Und[T] {
	return Und[T]{}
}

// FromPointer converts *T into Und[T].
// If v is nil, it returns an undefined Und.
// Otherwise, it returns Defined[T] whose value is the dereferenced v.
func FromPointer[T any](v *T) Und[T] {
	if v == nil {
		return Undefined[T]()
	}
	return Defined(*v)
}

func FromOption[T any](opt option.Option[option.Option[T]]) Und[T] {
	return Und[T]{opt: opt}
}

func (u Und[T]) IsZero() bool {
	return u.IsUndefined()
}

func (u Und[T]) IsDefined() bool {
	return u.opt.IsSome() && u.opt.Get().IsSome()
}

func (u Und[T]) IsNull() bool {
	return u.opt.IsSome() && u.opt.Get().IsNone()
}

func (u Und[T]) IsUndefined() bool {
	return u.opt.IsNone()
}

func (u Und[T]) Equal(other Und[T]) bool {
	return u.opt.Equal(other.opt)
}

func (u Und[T]) Get() T {
	if u.IsDefined() {
		return u.opt.Get().Get()
	}
	var zero T
	return zero
}

func (u Und[T]) Pointer() *T {
	if !u.IsDefined() {
		return nil
	}
	return u.opt.Get().Pointer()
}

func (u Und[T]) DoublePointer() **T {
	switch {
	case u.IsUndefined():
		return nil
	case u.IsNull():
		var t *T
		return &t
	default:
		t := u.opt.Get().Get()
		tt := &t
		return &tt
	}
}

func (u Und[T]) Unwrap() option.Option[option.Option[T]] {
	return u.opt
}

func (u Und[T]) Map(f func(option.Option[option.Option[T]]) option.Option[option.Option[T]]) Und[T] {
	return Und[T]{opt: f(u.opt)}
}

func (u Und[T]) MarshalJSON() ([]byte, error) {
	if u.IsUndefined() || u.IsNull() {
		return []byte(`null`), nil
	}
	return json.Marshal(u.opt.Get().Get())
}

func (u *Und[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*u = Null[T]()
		return nil
	}

	var t T
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	*u = Defined(t)
	return nil
}

func (u Und[T]) MarshalJSONV2(enc *jsontext.Encoder, opts jsonv2.Options) error {
	if !u.IsDefined() {
		return enc.WriteToken(jsontext.Null)
	}
	return jsonv2.MarshalEncode(enc, u.opt.Get().Get(), opts)
}

func (u *Und[T]) UnmarshalJSONV2(dec *jsontext.Decoder, opts jsonv2.Options) error {
	if dec.PeekKind() == 'n' {
		err := dec.SkipValue()
		if err != nil {
			return err
		}
		*u = Null[T]()
		return nil
	}
	var t T
	err := jsonv2.UnmarshalDecode(dec, &t, opts)
	if err != nil {
		return err
	}
	*u = Defined(t)
	return nil
}
