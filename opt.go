package und

import (
	"encoding/json"
)

type Equality[T any] interface {
	Equal(T) bool
}

var _ Equality[Option[int]] = (*Option[int])(nil)

// Option represents an optional value.
type Option[T any] struct {
	some bool
	v    T
}

func Some[T any](v T) Option[T] {
	return Option[T]{
		some: true,
		v:    v,
	}
}

func None[T any]() Option[T] {
	return Option[T]{}
}

func (o Option[T]) IsZero() bool {
	return o.IsNone()
}

func (o Option[T]) IsSome() bool {
	return o.some
}

// IsSomeAnd returns true if o is some and calling f with value of o returns true.
// Otherwise it returns false.
func (o Option[T]) IsSomeAnd(f func(T) bool) bool {
	if o.IsSome() {
		return f(o.Value())
	}
	return false
}

func (o Option[T]) IsNone() bool {
	return !o.IsSome()
}

// Value returns its internal as T.
// T would be zero value if o is None.
func (o Option[T]) Value() T {
	return o.v
}

// Pointer transforms o to *T, the plain conventional Go representation of an optional value.
// The value is copied by assignment before returned from Pointer.
func (o Option[T]) Pointer() *T {
	if o.IsNone() {
		return nil
	}
	t := o.v
	return &t
}

// Equal implements Equality[Option[T]].
//
// Equal tests o and other if both are Some or None.
// If both have value, it tests equality of their values.
//
// Equal panics If T or dynamic type of T is not comparable.
//
// Option is comparable if T is comparable.
// Equal only exists for cases where T needs a special Equal method (e.g. time.Time, slice based types)
//
// Equal first checks if T implements Equality[T], then also for *T.
// If it does not, then Equal compares by the `==` comparison operator.
func (o Option[T]) Equal(other Option[T]) bool {
	if !o.some || !other.some { // inlined simple case
		return o.some == other.some
	}

	return equal(o.v, other.v)
}

func equal[T any](t, u T) bool {
	// Try type assertion first.
	// The implemented interface has precedence.

	// Check for T. Below *T is also checked but in case T is already a pointer type, when T = *U, *(*U) might not implement Equality.
	eq, ok := any(t).(Equality[T])
	if ok {
		return eq.Equal(u)
	}
	// check for *T so that we can find method implemented for *T not only ones for T.
	eq, ok = any(&t).(Equality[T])
	if ok {
		return eq.Equal(u)
	}

	return any(t) == any(u) // may panic if T or dynamic type of T is uncomparable.
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if !o.some {
		// same as bytes.Clone.
		return []byte(`null`), nil
	}
	return json.Marshal(o.v)
}

func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		o.some = false
		var zero T
		o.v = zero
		return nil
	}

	err := json.Unmarshal(data, &o.v)
	if err != nil {
		return err
	}
	o.some = true
	return nil
}

// And returns u if o is some, otherwise None[T].
func (o Option[T]) And(u Option[T]) Option[T] {
	if o.IsSome() {
		return u
	} else {
		return None[T]()
	}
}

// AndThen calls f with value of o if o is some, otherwise returns None[T].
func (o Option[T]) AndThen(f func(x T) Option[T]) Option[T] {
	if o.IsSome() {
		return f(o.Value())
	} else {
		return None[T]()
	}
}

// Filter returns o if o is some and calling pred against o's value returns true.
// Otherwise it returns None[T].
func (o Option[T]) Filter(pred func(t T) bool) Option[T] {
	if o.IsSome() && pred(o.Value()) {
		return o
	}
	return None[T]()
}

// FlattenOption converts Option[Option[T]] into Option[T].
func FlattenOption[T any](o Option[Option[T]]) Option[T] {
	if o.IsNone() {
		return None[T]()
	}
	v := o.Value()
	if v.IsNone() {
		return None[T]()
	}
	return v
}

// MapOption returns Some[U] whose inner value is o's value mapped by f if o is Some.
// Otherwise it returns None[U].
func MapOption[T, U any](o Option[T], f func(T) U) Option[U] {
	if o.IsSome() {
		return Some(f(o.Value()))
	}
	return None[U]()
}

// Map returns Option[T] whose inner value is o's value mapped by f if o is some.
// Otherwise it returns None[T].
func (o Option[T]) Map(f func(v T) T) Option[T] {
	return MapOption(o, f)
}

// MapOrOption returns value o's value applied by f if o is some.
// Otherwise it returns defaultValue.
func MapOrOption[T, U any](o Option[T], defaultValue U, f func(T) U) U {
	if o.IsNone() {
		return defaultValue
	}
	return f(o.Value())
}

// MapOr returns value o's value applied by f if o is some.
// Otherwise it returns defaultValue.
func (o Option[T]) MapOr(defaultValue T, f func(T) T) T {
	return MapOrOption(o, defaultValue, f)
}

// MapOrElseOption returns value o's value applied by f if o is some.
// Otherwise it returns a defaultFn result.
func MapOrElseOption[T, U any](o Option[T], defaultFn func() U, f func(T) U) U {
	if o.IsNone() {
		return defaultFn()
	}
	return f(o.Value())
}

// MapOrElse returns value o's value applied by f if o is some.
// Otherwise it returns a defaultFn result.
func (o Option[T]) MapOrElse(defaultFn func() T, f func(T) T) T {
	return MapOrElseOption(o, defaultFn, f)
}

// Or returns o if o is some, otherwise u.
func (o Option[T]) Or(u Option[T]) Option[T] {
	if o.IsSome() {
		return o
	} else {
		return u
	}
}

// OrElse returns o if o is some, otherwise calls f and returns the result.
func (o Option[T]) OrElse(f func() Option[T]) Option[T] {
	if o.IsSome() {
		return o
	} else {
		return f()
	}
}

// Xor returns o or u if either is some.
// If both are some or both none, it returns None[T].
func (o Option[T]) Xor(u Option[T]) Option[T] {
	if o.IsSome() && u.IsNone() {
		return o
	}
	if o.IsNone() && u.IsSome() {
		return u
	}
	return None[T]()
}
