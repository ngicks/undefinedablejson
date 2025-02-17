package elastic

import (
	"encoding/json"
	"encoding/xml"
	"iter"
	"log/slog"
	"slices"

	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
)

// portable methods that can be copied from github.com/ngicks/und/elastic into github.com/ngicks/und/sliceund/elastic

// FromValue returns Elastic[T] with single some value.
func FromValue[T any](t T) Elastic[T] {
	return FromOptions(option.Some(t))
}

// FromPointer converts nil to undefined Elastic[T],
// or defined one whose internal value is dereferenced t.
//
// If you need to keep t as pointer, use [WrapPointer] instead.
func FromPointer[T any](t *T) Elastic[T] {
	if t == nil {
		return Undefined[T]()
	}
	return FromValue(*t)
}

// WrapPointer converts *T into Elastic[*T].
// The elastic value is defined if t is non nil, undefined otherwise.
//
// If you want t to be dereferenced, use [FromPointer] instead.
func WrapPointer[T any](t *T) Elastic[*T] {
	if t == nil {
		return Undefined[*T]()
	}
	return FromValue(t)
}

// FromValues converts variadic T values into an Elastic[T].
func FromValues[T any](ts ...T) Elastic[T] {
	opts := make(option.Options[T], len(ts))
	for i, value := range ts {
		opts[i] = option.Some(value)
	}
	return FromOptions(opts...)
}

// FromPointers converts variadic *T values into an Elastic[T],
// treating nil as None[T], and non-nil as Some[T].
//
// If you need to keep t-s as pointer, use [WrapPointers] instead.
func FromPointers[T any](ps ...*T) Elastic[T] {
	opts := make(option.Options[T], len(ps))
	for i, p := range ps {
		opts[i] = option.FromPointer(p)
	}
	return FromOptions(opts...)
}

// FromPointers converts variadic *T values into an Elastic[*T],
// treating nil as None[*T], and non-nil as Some[*T].
//
// If you need t-s to be dereferenced, use [FromPointers] instead.
func WrapPointers[T any](ps ...*T) Elastic[*T] {
	opts := make(option.Options[*T], len(ps))
	for i, p := range ps {
		opts[i] = option.WrapPointer(p)
	}
	return FromOptions(opts...)
}

// IsZero is an alias for IsUndefined.
func (e Elastic[T]) IsZero() bool {
	return e.IsUndefined()
}

func (e Elastic[T]) UndValidate() error {
	return e.inner().Value().UndValidate()
}

func (e Elastic[T]) UndCheck() error {
	return e.inner().Value().UndCheck()
}

// MarshalJSON implements json.Marshaler.
func (u Elastic[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.inner())
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Elastic[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*e = Null[T]()
		return nil
	}

	if len(data) >= 2 && data[0] == '[' {
		var t option.Options[T]
		err := json.Unmarshal(data, &t)
		// might be T is []U, and this fails
		// since it should've been [[...data...],[...data...]]
		if err == nil {
			*e = FromOptions(t...)
			return nil
		}
	}

	var t option.Option[T]
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}
	*e = FromOptions(t)
	return nil
}

// MarshalXML implements xml.Marshaler.
func (e Elastic[T]) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return e.Unwrap().MarshalXML(enc, start)
}

// LogValue implements slog.LogValuer.
func (e Elastic[T]) LogValue() slog.Value {
	return e.Unwrap().LogValue()
}

// Map returns a new Elastic value whose internal value is mapped by f.
//
// Be cautious that f will only be applied to some value; null values remain as null.
func Map[T, U any](e Elastic[T], f func(t T) U) Elastic[U] {
	switch {
	case e.IsUndefined():
		return Undefined[U]()
	case e.IsNull():
		return Null[U]()
	default:
		return FromOptionSeq(mapSeq(f, slices.Values(e.Unwrap().Value())))
	}
}

func mapSeq[T, U any](f func(T) U, seq iter.Seq[option.Option[T]]) iter.Seq[option.Option[U]] {
	return func(yield func(option.Option[U]) bool) {
		for opt := range seq {
			if !yield(option.Map(opt, f)) {
				return
			}
		}
	}
}

// State returns e's value state.
func (e Elastic[T]) State() und.State {
	switch {
	case e.IsUndefined():
		return und.StateUndefined
	case e.IsNull():
		return und.StateNull
	default:
		return und.StateDefined
	}
}
