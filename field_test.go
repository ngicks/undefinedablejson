package undefinedablejson_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/ngicks/type-param-common/util"
	"github.com/ngicks/undefinedablejson"
)

type CustomizedEquality struct {
	V *int
}

func (e CustomizedEquality) Equal(other CustomizedEquality) bool {
	// Does it look something like ring buffer index?
	return *e.V%30 == *other.V%30
}

type NonComparableButEquality []string

func (e NonComparableButEquality) Equal(other NonComparableButEquality) bool {
	if len(e) != len(other) {
		return false
	}
	for idx := range e {
		if e[idx] != other[idx] {
			return false
		}
	}
	return true
}

type pairNullable[T any] struct {
	l, r  undefinedablejson.Nullable[T]
	equal bool
}

func runNullableTests[T any](t *testing.T, pairs []pairNullable[T]) bool {
	t.Helper()
	didError := false
	for idx, testCase := range pairs {
		isEqual := testCase.l.Equal(testCase.r)
		if isEqual != testCase.equal {
			var shouldBe string
			if testCase.equal {
				shouldBe = "be equal"
			} else {
				shouldBe = "not be equal"
			}
			didError = true
			t.Errorf(
				"case number = %d. should %s: type = %T left = %v, right = %v",
				idx, shouldBe, testCase.l, formatValue[T](testCase.l), formatValue[T](testCase.r),
			)
		}
	}

	return !didError
}

type pairUndefinedable[T any] struct {
	l, r  undefinedablejson.Undefinedable[T]
	equal bool
}

func runUndefinedableTests[T any](t *testing.T, pairs []pairUndefinedable[T]) bool {
	t.Helper()
	didError := false
	for idx, testCase := range pairs {
		isEqual := testCase.l.Equal(testCase.r)
		if isEqual != testCase.equal {
			didError = true
			t.Errorf(
				"case number = %d. not equal: type = %T left = %v, right = %v",
				idx, testCase.l, formatValue[T](testCase.l), formatValue[T](testCase.r),
			)
		}
	}

	return !didError
}

func formatValue[T any](v interface {
	IsNull() bool
	Value() T
}) string {
	if und, ok := v.(interface{ IsUndefined() bool }); ok && und.IsUndefined() {
		return `<undefined>`
	}
	if v.IsNull() {
		return `<null>`
	} else {
		return fmt.Sprintf("%+v", v.Value())
	}
}

// case 1: comparable.
var caseComparable = []pairNullable[int]{
	{
		undefinedablejson.NonNull(123), undefinedablejson.NonNull(123),
		true,
	},
	{
		undefinedablejson.NonNull(123), undefinedablejson.NonNull(224),
		false,
	},
	{
		undefinedablejson.Null[int](), undefinedablejson.Null[int](),
		true,
	},
	{
		undefinedablejson.NonNull(123), undefinedablejson.Null[int](),
		false,
	},
	{
		undefinedablejson.Null[int](), undefinedablejson.NonNull(123),
		false,
	},
}

// case 2: non comparable
var caseNonComparable = []pairNullable[[]string]{
	{
		undefinedablejson.NonNull([]string{"foo"}), undefinedablejson.NonNull([]string{"foo"}),
		false,
	},
	{
		undefinedablejson.NonNull([]string{"foo"}), undefinedablejson.NonNull([]string{"bar"}),
		false,
	},
	{
		undefinedablejson.Null[[]string](), undefinedablejson.Null[[]string](),
		true,
	},
	{
		undefinedablejson.NonNull([]string{"foo"}), undefinedablejson.Null[[]string](),
		false,
	},
	{
		undefinedablejson.Null[[]string](), undefinedablejson.NonNull([]string{"foo"}),
		false,
	},
}

var sampleSlice = []string{"foo", "bar", "baz"}

// case 3: pointer value
var casePointer = []pairNullable[*[]string]{
	{
		undefinedablejson.NonNull(&[]string{"foo"}), undefinedablejson.NonNull(&[]string{"foo"}),
		false,
	},
	{
		undefinedablejson.NonNull(&[]string{"foo"}), undefinedablejson.NonNull(&[]string{"bar"}),
		false,
	},
	{ // same pointer = true (of course).
		undefinedablejson.NonNull(&sampleSlice), undefinedablejson.NonNull(&sampleSlice),
		true,
	},
	{
		undefinedablejson.Null[*[]string](), undefinedablejson.Null[*[]string](),
		true,
	},
	{
		undefinedablejson.NonNull(&[]string{"foo"}), undefinedablejson.Null[*[]string](),
		false,
	},
	{
		undefinedablejson.Null[*[]string](), undefinedablejson.NonNull(&[]string{"foo"}),
		false,
	},
}

// case 4: non comparable but implements Equality.
var caseNonComparableButCustomEquality = []pairNullable[NonComparableButEquality]{
	{
		undefinedablejson.NonNull(NonComparableButEquality{"foo"}), undefinedablejson.NonNull(NonComparableButEquality{"foo"}),
		true,
	},
	{
		undefinedablejson.NonNull(NonComparableButEquality{"foo"}), undefinedablejson.NonNull(NonComparableButEquality{"bar"}),
		false,
	},
	{
		undefinedablejson.Null[NonComparableButEquality](), undefinedablejson.Null[NonComparableButEquality](),
		true,
	},
	{
		undefinedablejson.NonNull(NonComparableButEquality{"foo"}), undefinedablejson.Null[NonComparableButEquality](),
		false,
	},
	{
		undefinedablejson.Null[NonComparableButEquality](), undefinedablejson.NonNull(NonComparableButEquality{"foo"}),
		false,
	},
}

// case 5: comparable but has customized equality.
var caseComparableButCustomEquality = []pairNullable[CustomizedEquality]{
	{
		undefinedablejson.NonNull(CustomizedEquality{util.Escape(123)}), undefinedablejson.NonNull(CustomizedEquality{util.Escape(123)}),
		true,
	},
	{ // uses customized equality method
		undefinedablejson.NonNull(CustomizedEquality{util.Escape(1)}), undefinedablejson.NonNull(CustomizedEquality{util.Escape(31)}),
		true,
	},
	{
		undefinedablejson.NonNull(CustomizedEquality{util.Escape(123)}), undefinedablejson.NonNull(CustomizedEquality{util.Escape(124)}),
		false,
	},
	{
		undefinedablejson.Null[CustomizedEquality](), undefinedablejson.Null[CustomizedEquality](),
		true,
	},
	{
		undefinedablejson.NonNull(CustomizedEquality{util.Escape(123)}), undefinedablejson.Null[CustomizedEquality](),
		false,
	},
	{
		undefinedablejson.Null[CustomizedEquality](), undefinedablejson.NonNull(CustomizedEquality{util.Escape(123)}),
		false,
	},
}

func TestFields_equality(t *testing.T) {
	runNullableTests(t, caseComparable)
	runNullableTests(t, caseNonComparable)
	runNullableTests(t, casePointer)
	runNullableTests(t, caseNonComparableButCustomEquality)
	runNullableTests(t, caseComparableButCustomEquality)

	runUndefinedableTests(t, convertNullableCasesToUndefined(caseComparable))
	runUndefinedableTests(t, convertNullableCasesToUndefined(caseNonComparable))
	runUndefinedableTests(t, convertNullableCasesToUndefined(casePointer))
	runUndefinedableTests(t, convertNullableCasesToUndefined(caseNonComparableButCustomEquality))
	runUndefinedableTests(t, convertNullableCasesToUndefined(caseComparableButCustomEquality))

	runUndefinedableTests(t, []pairUndefinedable[int]{
		{ // undefined - undefined
			undefinedablejson.UndefinedField[int](), undefinedablejson.UndefinedField[int](),
			true,
		},
		// undefined - value
		{
			undefinedablejson.Field(123), undefinedablejson.UndefinedField[int](),
			false,
		}, {
			undefinedablejson.UndefinedField[int](), undefinedablejson.Field(123),
			false,
		},
		// undefined - null
		{
			undefinedablejson.UndefinedField[int](), undefinedablejson.NullField[int](),
			false,
		},
		{
			undefinedablejson.NullField[int](), undefinedablejson.UndefinedField[int](),
			false,
		},
	})
}
func convertNullableCasesToUndefined[T any](cases []pairNullable[T]) []pairUndefinedable[T] {
	ret := make([]pairUndefinedable[T], len(cases))

	for idx, testCase := range cases {
		var l undefinedablejson.Undefinedable[T]
		if testCase.l.IsNull() {
			l = undefinedablejson.NullField[T]()
		} else {
			l = undefinedablejson.Field(testCase.l.Value())
		}

		var r undefinedablejson.Undefinedable[T]
		if testCase.r.IsNull() {
			r = undefinedablejson.NullField[T]()
		} else {
			r = undefinedablejson.Field(testCase.r.Value())
		}

		ret[idx] = pairUndefinedable[T]{
			l, r,
			testCase.equal,
		}
	}
	return ret
}

type RaceTestA struct {
	Foo undefinedablejson.Undefinedable[int] `und:"string"`
}

func (r RaceTestA) F() int {
	return r.Foo.Value()
}

type RaceTestB struct {
	Foo undefinedablejson.Undefinedable[int] `und:"string"`
}

func (r RaceTestB) F() int {
	return r.Foo.Value()
}

type RaceTestC struct {
	Foo undefinedablejson.Undefinedable[int] `und:"string"`
	ErroneousEmbedded
}

func (r RaceTestC) F() int {
	return r.Foo.Value()
}

type ErroneousEmbedded struct {
	Bar string
}

func (e ErroneousEmbedded) MarshalJSON() ([]byte, error) {
	return []byte(`null`), nil
}

func unmarshal[T interface{ F() int }]() error {
	var (
		t                 T
		err, unmarshalErr error
	)
	if rand.Int31n(10) >= 5 {
		err = undefinedablejson.UnmarshalFieldsJSON([]byte(`{"Foo":"123"}`), &t)
		if err != nil {
			return err
		}
		if t.F() != 123 {
			unmarshalErr = fmt.Errorf("error")
		}
	} else {
		err = undefinedablejson.UnmarshalFieldsJSON([]byte(`{}`), &t)
		if err != nil {
			return err
		}
		if t.F() != 0 {
			unmarshalErr = fmt.Errorf("error")
		}
	}

	if unmarshalErr != nil {
		return unmarshalErr
	}
	return nil
}

func TestSerde_serde_info_cache_race(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			time.Sleep(time.Duration(rand.Int31n(5)))
			if err := unmarshal[RaceTestA](); err != nil {
				t.Errorf("%+v", err)
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			time.Sleep(time.Duration(rand.Int31n(5)))
			if err := unmarshal[RaceTestB](); err != nil {
				t.Errorf("%+v", err)
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			time.Sleep(time.Duration(rand.Int31n(5)))
			if err := unmarshal[RaceTestC](); err == nil {
				t.Error("must be error but nil")
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
