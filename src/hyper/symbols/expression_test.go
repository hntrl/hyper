package symbols_test

import (
	"bufio"
	"strings"
	"testing"

	"github.com/hntrl/hyper/src/hyper/ast"
	"github.com/hntrl/hyper/src/hyper/parser"
	"github.com/hntrl/hyper/src/hyper/symbols"
	"github.com/hntrl/hyper/src/hyper/symbols/errors"
	"github.com/stretchr/testify/assert"
)

func textParser(lit string) *parser.Parser {
	reader := bufio.NewReader(strings.NewReader(lit + "\n"))
	lexer := parser.NewLexer(reader)
	return parser.NewParser(lexer)
}

func EqualErrorCode(t *testing.T, err error, errCode errors.Code, msgAndArgs ...interface{}) bool {
	if interpreterError, ok := err.(errors.InterpreterError); ok {
		if interpreterError.Code == errCode {
			return true
		}
		assert.Fail(t, "Code does not match", msgAndArgs...)
		return false
	}
	assert.Fail(t, "Error is not an InterpreterError", msgAndArgs...)
	return false
}
func EqualStandardError(t *testing.T, err error, errCode errors.Code, errString string, msgAndArgs ...interface{}) bool {
	if interpreterError, ok := err.(errors.InterpreterError); ok {
		if interpreterError.Code == errCode {
			if interpreterError.Msg == errString {
				return true
			}
			assert.Fail(t, "Message does not match", msgAndArgs...)
			return false
		}
		assert.Fail(t, "Code does not match", msgAndArgs...)
		return false
	}
	assert.Fail(t, "Error is not an InterpreterError", msgAndArgs...)
	return false
}

func SymbolTable(local map[string]symbols.ScopeValue) *symbols.SymbolTable {
	return &symbols.SymbolTable{
		Root:      nil,
		Immutable: make(map[string]symbols.ScopeValue),
		Local:     local,
		LoopState: nil,
	}
}

func TestEvaluateTypeExpression(t *testing.T) {
	fooClass := GenericClass{
		Name: "Foo",
		Properties: map[string]symbols.Class{
			"a": symbols.String,
			"b": symbols.Integer,
		},
	}
	st := SymbolTable(map[string]symbols.ScopeValue{
		"Foo":    fooClass,
		"Bar":    symbols.StringValue("Bar"),
		"String": symbols.String,
	})
	t.Run("can return basic type", func(t *testing.T) {
		node, err := ast.ParseTypeExpression(textParser("Foo"))
		if err != nil {
			t.Fatal(err)
		}
		class, err := st.EvaluateTypeExpression(*node)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, fooClass, class, "should return Foo class")
	})
	t.Run("will reject if selector isn't a class", func(t *testing.T) {
		node, err := ast.ParseTypeExpression(textParser("Bar"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = st.EvaluateTypeExpression(*node)
		EqualErrorCode(t, err, errors.InvalidClass, "should return invalid class error")
	})
	t.Run("will reject if the selector doesn't exist", func(t *testing.T) {
		node, err := ast.ParseTypeExpression(textParser("Baz"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = st.EvaluateTypeExpression(*node)
		EqualErrorCode(t, err, errors.UnknownSelector, "should return unknown selector error")
	})
	t.Run("can create partial class", func(t *testing.T) {
		node, err := ast.ParseTypeExpression(textParser("Partial<Foo>"))
		if err != nil {
			t.Fatal(err)
		}
		class, err := st.EvaluateTypeExpression(*node)
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(
			t,
			symbols.NewPartialClass(fooClass),
			class,
			"should return Partial<Foo> class",
		)
	})
	t.Run("will panic if partial class doesn't have properties", func(t *testing.T) {
		assert.PanicsWithError(t, "cannot create partial class without properties", func() {
			node, err := ast.ParseTypeExpression(textParser("Partial<String>"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = st.EvaluateTypeExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
		})
	})
	t.Run("can create array class", func(t *testing.T) {
		node, err := ast.ParseTypeExpression(textParser("[]Foo"))
		if err != nil {
			t.Fatal(err)
		}
		class, err := st.EvaluateTypeExpression(*node)
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(
			t,
			symbols.NewArrayClass(fooClass),
			class,
			"should return []Foo class",
		)
	})
	t.Run("can create partial array class", func(t *testing.T) {
		node, err := ast.ParseTypeExpression(textParser("[]Partial<Foo>"))
		if err != nil {
			t.Fatal(err)
		}
		class, err := st.EvaluateTypeExpression(*node)
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(
			t,
			symbols.NewArrayClass(
				symbols.NewPartialClass(fooClass),
			),
			class,
			"should return []Partial<Foo> class",
		)
	})
	t.Run("can create nilable class", func(t *testing.T) {
		node, err := ast.ParseTypeExpression(textParser("Foo?"))
		if err != nil {
			t.Fatal(err)
		}
		class, err := st.EvaluateTypeExpression(*node)
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(
			t,
			symbols.NewNilableClass(fooClass),
			class,
			"should return Foo? class",
		)
	})
	t.Run("can create nilable array class", func(t *testing.T) {
		node, err := ast.ParseTypeExpression(textParser("[]Foo?"))
		if err != nil {
			t.Fatal(err)
		}
		class, err := st.EvaluateTypeExpression(*node)
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(
			t,
			symbols.NewNilableClass(
				symbols.NewArrayClass(fooClass),
			),
			class,
			"should return []Foo? class",
		)
	})
	t.Run("can create nilable partial class", func(t *testing.T) {
		node, err := ast.ParseTypeExpression(textParser("Partial<Foo>?"))
		if err != nil {
			t.Fatal(err)
		}
		class, err := st.EvaluateTypeExpression(*node)
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(
			t,
			symbols.NewNilableClass(
				symbols.NewPartialClass(fooClass),
			),
			class,
			"should return Partial<Foo>? class",
		)
	})
	t.Run("can create nilable array partial class", func(t *testing.T) {
		node, err := ast.ParseTypeExpression(textParser("[]Partial<Foo>?"))
		if err != nil {
			t.Fatal(err)
		}
		class, err := st.EvaluateTypeExpression(*node)
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(
			t,
			symbols.NewArrayClass(
				symbols.NewNilableClass(
					symbols.NewPartialClass(fooClass),
				),
			),
			class,
			"should return []Partial<Foo>? class",
		)
	})
}

// ResolveExpression
// EvaluateExpression

func TestArrayExpression(t *testing.T) {
	fooClass := GenericClass{
		Name: "Foo",
		Properties: map[string]symbols.Class{
			"a": symbols.String,
			"b": symbols.Integer,
		},
	}
	st := SymbolTable(map[string]symbols.ScopeValue{
		"Foo":    fooClass,
		"Bar":    symbols.StringValue("Bar"),
		"String": symbols.String,
	})
	t.Run("can create empty array", func(t *testing.T) {
		node, err := ast.ParseArrayExpression(textParser("[]Foo{}"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveArrayExpression", func(t *testing.T) {
			value, err := st.ResolveArrayExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(
				t,
				symbols.NewArray(fooClass, 0),
				value,
				"should return []Foo value",
			)
		})
		t.Run("EvaluateArrayExpression", func(t *testing.T) {
			value, err := st.EvaluateArrayExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(
				t,
				&symbols.ExpectedValueObject{symbols.NewArrayClass(fooClass)},
				value,
				"should return expected []Foo value",
			)
		})
	})
	t.Run("will reject if the array type is invalid", func(t *testing.T) {
		node, err := ast.ParseArrayExpression(textParser("[]Bar{}"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveArrayExpression", func(t *testing.T) {
			_, err := st.ResolveArrayExpression(*node)
			EqualErrorCode(t, err, errors.InvalidClass, "should return invalid class error")
		})
		t.Run("EvaluateArrayExpression", func(t *testing.T) {
			_, err = st.EvaluateArrayExpression(*node)
			EqualErrorCode(t, err, errors.InvalidClass, "should return invalid class error")
		})
	})
	t.Run("will reject if the array type is unknown", func(t *testing.T) {
		node, err := ast.ParseArrayExpression(textParser("[]Baz{}"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveArrayExpression", func(t *testing.T) {
			_, err := st.ResolveArrayExpression(*node)
			EqualErrorCode(t, err, errors.UnknownSelector, "should return unknown selector error")
		})
		t.Run("EvaluateArrayExpression", func(t *testing.T) {
			_, err = st.EvaluateArrayExpression(*node)
			EqualErrorCode(t, err, errors.UnknownSelector, "should return unknown selector error")
		})
	})
	t.Run("can create array with single value types", func(t *testing.T) {
		t.Run("with primitive array type", func(t *testing.T) {
			node, err := ast.ParseArrayExpression(textParser("[]String{\"abc\", \"def\"}"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveArrayExpression", func(t *testing.T) {
				value, err := st.ResolveArrayExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				expectedValue := symbols.NewArray(symbols.String, 2)
				expectedValue.Set(0, symbols.StringValue("abc"))
				expectedValue.Set(1, symbols.StringValue("def"))
				assert.EqualValues(
					t,
					expectedValue,
					value,
					"should return []String{\"abc\",\"def\"} value",
				)
			})
			t.Run("EvaluateArrayExpression", func(t *testing.T) {
				value, err := st.EvaluateArrayExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.EqualValues(
					t,
					&symbols.ExpectedValueObject{symbols.NewArrayClass(symbols.String)},
					value,
					"should return expected []String value",
				)
			})
		})
		t.Run("with object array type", func(t *testing.T) {
			node, err := ast.ParseArrayExpression(textParser("[]Foo{Foo{a: \"abc\", b: 0}, Foo{a: \"def\", b: 1}}"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveArrayExpression", func(t *testing.T) {
				value, err := st.ResolveArrayExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				expectedValue := symbols.NewArray(fooClass, 2)
				expectedValue.Set(0, GenericValue{
					class: fooClass,
					data: map[string]symbols.ValueObject{
						"a": symbols.StringValue("abc"),
						"b": symbols.IntegerValue(0),
					},
				})
				expectedValue.Set(1, GenericValue{
					class: fooClass,
					data: map[string]symbols.ValueObject{
						"a": symbols.StringValue("def"),
						"b": symbols.IntegerValue(1),
					},
				})
				assert.EqualValues(
					t,
					expectedValue,
					value,
					"should return []Foo{Foo{a: \"abc\", b: 0}, Foo{a: \"def\", b: 1}} value",
				)
			})
			t.Run("EvaluateArrayExpression", func(t *testing.T) {
				value, err := st.EvaluateArrayExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.EqualValues(
					t,
					&symbols.ExpectedValueObject{symbols.NewArrayClass(fooClass)},
					value,
					"should return expected []Foo value",
				)
			})
		})
	})
	t.Run("can create array with multiple value types", func(t *testing.T) {
		node, err := ast.ParseArrayExpression(textParser("[]Foo{{a: \"abc\", b: 0}, Foo{a: \"def\", b: 1}}"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveArrayExpression", func(t *testing.T) {
			value, err := st.ResolveArrayExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			expectedValue := symbols.NewArray(fooClass, 2)
			expectedValue.Set(0, GenericValue{
				class: fooClass,
				data: map[string]symbols.ValueObject{
					"a": symbols.StringValue("abc"),
					"b": symbols.IntegerValue(0),
				},
			})
			expectedValue.Set(1, GenericValue{
				class: fooClass,
				data: map[string]symbols.ValueObject{
					"a": symbols.StringValue("def"),
					"b": symbols.IntegerValue(1),
				},
			})
			assert.EqualValues(
				t,
				expectedValue,
				value,
				"should return []Foo{Foo{a: \"abc\", b: 0}, Foo{a: \"def\", b: 1}} value",
			)
		})
		t.Run("EvaluateArrayExpression", func(t *testing.T) {
			value, err := st.EvaluateArrayExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(
				t,
				&symbols.ExpectedValueObject{symbols.NewArrayClass(fooClass)},
				value,
				"should return expected []Foo value",
			)
		})
	})
	t.Run("will reject if a value cannot be used", func(t *testing.T) {
		node, err := ast.ParseArrayExpression(textParser("[]String{\"abc\", Foo{a: \"def\", b: 1}}"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveArrayExpression", func(t *testing.T) {
			_, err := st.ResolveArrayExpression(*node)
			EqualErrorCode(t, err, errors.CannotConstruct, "should return cannot construct error")
		})
		t.Run("EvaluateArrayExpression", func(t *testing.T) {
			_, err = st.EvaluateArrayExpression(*node)
			EqualErrorCode(t, err, errors.CannotConstruct, "should return cannot construct error")
		})
	})
}

func TestInstanceExpression(t *testing.T) {
	fooClass := GenericClass{
		Name: "Foo",
		Properties: map[string]symbols.Class{
			"a": symbols.String,
			"b": symbols.Integer,
		},
	}
	st := SymbolTable(map[string]symbols.ScopeValue{
		"Foo":    fooClass,
		"Bar":    symbols.StringValue("Bar"),
		"String": symbols.String,
	})
	t.Run("can create instance of class", func(t *testing.T) {
		node, err := ast.ParseInstanceExpression(textParser("Foo{a: \"abc\", b: 0}"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveInstanceExpression", func(t *testing.T) {
			value, err := st.ResolveInstanceExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(
				t,
				GenericValue{
					class: fooClass,
					data: map[string]symbols.ValueObject{
						"a": symbols.StringValue("abc"),
						"b": symbols.IntegerValue(0),
					},
				},
				value,
				"should return Foo{a: \"abc\", b: 0} value",
			)
		})
		t.Run("EvaluateInstanceExpression", func(t *testing.T) {
			value, err := st.EvaluateInstanceExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(
				t,
				&symbols.ExpectedValueObject{fooClass},
				value,
				"should return expected Foo value",
			)
		})
	})
	t.Run("will reject if the class doesn't have properties", func(t *testing.T) {
		node, err := ast.ParseInstanceExpression(textParser("String{a: \"abc\", b: 0}"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveInstanceExpression", func(t *testing.T) {
			_, err := st.ResolveInstanceExpression(*node)
			EqualErrorCode(t, err, errors.CannotConstruct, "should return cannot construct error")
		})
		t.Run("EvaluateInstanceExpression", func(t *testing.T) {
			_, err = st.EvaluateInstanceExpression(*node)
			EqualErrorCode(t, err, errors.CannotConstruct, "should return cannot construct error")
		})
	})
	t.Run("will reject if the class is not a class", func(t *testing.T) {
		node, err := ast.ParseInstanceExpression(textParser("Bar{a: \"abc\", b: 0}"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveInstanceExpression", func(t *testing.T) {
			_, err := st.ResolveInstanceExpression(*node)
			EqualErrorCode(t, err, errors.InvalidInstanceableTarget, "should return not instanceable error")
		})
		t.Run("EvaluateInstanceExpression", func(t *testing.T) {
			_, err = st.EvaluateInstanceExpression(*node)
			EqualErrorCode(t, err, errors.InvalidInstanceableTarget, "should return not instanceable error")
		})
	})
	t.Run("will reject if the class is unknown", func(t *testing.T) {
		node, err := ast.ParseInstanceExpression(textParser("Baz{a: \"abc\", b: 0}"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveInstanceExpression", func(t *testing.T) {
			_, err := st.ResolveInstanceExpression(*node)
			EqualErrorCode(t, err, errors.UnknownSelector, "should return unknown selector error")
		})
		t.Run("EvaluateInstanceExpression", func(t *testing.T) {
			_, err = st.EvaluateInstanceExpression(*node)
			EqualErrorCode(t, err, errors.UnknownSelector, "should return unknown selector error")
		})
	})
}

func TestUnaryExpression(t *testing.T) {
	st := SymbolTable(map[string]symbols.ScopeValue{
		"Foo":    symbols.IntegerValue(1),
		"Bar":    symbols.IntegerValue(-1),
		"Baz":    symbols.BooleanValue(false),
		"String": symbols.String,
	})
	t.Run("+Foo (1)", func(t *testing.T) {
		node, err := ast.ParseUnaryExpression(textParser("+Foo"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveUnaryExpression", func(t *testing.T) {
			value, err := st.ResolveUnaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, symbols.IntegerValue(1), value)
		})
		t.Run("EvaluateUnaryExpression", func(t *testing.T) {
			value, err := st.EvaluateUnaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, &symbols.ExpectedValueObject{symbols.Integer}, value)
		})
	})
	t.Run("+Bar (-1)", func(t *testing.T) {
		node, err := ast.ParseUnaryExpression(textParser("+Bar"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveUnaryExpression", func(t *testing.T) {
			value, err := st.ResolveUnaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, symbols.IntegerValue(1), value)
		})
		t.Run("EvaluateUnaryExpression", func(t *testing.T) {
			value, err := st.EvaluateUnaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, &symbols.ExpectedValueObject{symbols.Integer}, value)
		})
	})
	t.Run("+Baz (false)", func(t *testing.T) {
		node, err := ast.ParseUnaryExpression(textParser("+Baz"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveUnaryExpression", func(t *testing.T) {
			_, err := st.ResolveUnaryExpression(*node)
			EqualStandardError(t, err, errors.CannotConstruct, "cannot construct Number from Bool")
		})
		t.Run("EvaluateUnaryExpression", func(t *testing.T) {
			_, err := st.EvaluateUnaryExpression(*node)
			EqualStandardError(t, err, errors.CannotConstruct, "cannot construct Number from Bool")
		})
	})
	t.Run("+String", func(t *testing.T) {
		node, err := ast.ParseUnaryExpression(textParser("+String"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveUnaryExpression", func(t *testing.T) {
			_, err := st.ResolveUnaryExpression(*node)
			EqualErrorCode(t, err, errors.InvalidValueExpression)
		})
		t.Run("EvaluateUnaryExpression", func(t *testing.T) {
			_, err := st.EvaluateUnaryExpression(*node)
			EqualErrorCode(t, err, errors.InvalidValueExpression)
		})
	})
	t.Run("-Foo (1)", func(t *testing.T) {
		node, err := ast.ParseUnaryExpression(textParser("-Foo"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveUnaryExpression", func(t *testing.T) {
			value, err := st.ResolveUnaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, symbols.IntegerValue(-1), value)
		})
		t.Run("EvaluateUnaryExpression", func(t *testing.T) {
			value, err := st.EvaluateUnaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, &symbols.ExpectedValueObject{symbols.Integer}, value)
		})
	})
	t.Run("-Bar (-1)", func(t *testing.T) {
		node, err := ast.ParseUnaryExpression(textParser("-Bar"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveUnaryExpression", func(t *testing.T) {
			value, err := st.ResolveUnaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, symbols.IntegerValue(1), value)
		})
		t.Run("EvaluateUnaryExpression", func(t *testing.T) {
			value, err := st.EvaluateUnaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, &symbols.ExpectedValueObject{symbols.Integer}, value)
		})
	})
	t.Run("-Baz (false)", func(t *testing.T) {
		node, err := ast.ParseUnaryExpression(textParser("-Baz"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveUnaryExpression", func(t *testing.T) {
			_, err := st.ResolveUnaryExpression(*node)
			EqualStandardError(t, err, errors.CannotConstruct, "cannot construct Number from Bool")
		})
		t.Run("EvaluateUnaryExpression", func(t *testing.T) {
			_, err := st.EvaluateUnaryExpression(*node)
			EqualStandardError(t, err, errors.CannotConstruct, "cannot construct Number from Bool")
		})
	})
	t.Run("-String", func(t *testing.T) {
		node, err := ast.ParseUnaryExpression(textParser("-String"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveUnaryExpression", func(t *testing.T) {
			_, err := st.ResolveUnaryExpression(*node)
			EqualErrorCode(t, err, errors.InvalidValueExpression)
		})
		t.Run("EvaluateUnaryExpression", func(t *testing.T) {
			_, err := st.EvaluateUnaryExpression(*node)
			EqualErrorCode(t, err, errors.InvalidValueExpression)
		})
	})
	t.Run("!Foo (1)", func(t *testing.T) {
		node, err := ast.ParseUnaryExpression(textParser("!Foo"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveUnaryExpression", func(t *testing.T) {
			_, err := st.ResolveUnaryExpression(*node)
			EqualErrorCode(t, err, errors.InvalidUnaryOperand)
		})
		t.Run("EvaluateUnaryExpression", func(t *testing.T) {
			_, err := st.EvaluateUnaryExpression(*node)
			EqualErrorCode(t, err, errors.InvalidUnaryOperand)
		})
	})
	t.Run("!Baz (false)", func(t *testing.T) {
		node, err := ast.ParseUnaryExpression(textParser("!Baz"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveUnaryExpression", func(t *testing.T) {
			value, err := st.ResolveUnaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, symbols.BooleanValue(true), value)
		})
		t.Run("EvaluateUnaryExpression", func(t *testing.T) {
			value, err := st.EvaluateUnaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, &symbols.ExpectedValueObject{symbols.Boolean}, value)
		})
	})
}

func TestBinaryExpression(t *testing.T) {
	st := SymbolTable(map[string]symbols.ScopeValue{
		"Foo":    symbols.IntegerValue(1),
		"Bar":    symbols.IntegerValue(-1),
		"Baz":    symbols.DoubleValue(1.05),
		"Lorem":  symbols.StringValue("Lorem ipsum"),
		"String": symbols.String,
	})
	left, err := ast.ParseExpression(textParser("Foo"))
	if err != nil {
		t.Fatal(err)
	}
	t.Run("will parse binary expression with operator", func(t *testing.T) {
		node, err := ast.ParseBinaryExpression(textParser("+ Bar"), *left)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveBinaryExpression", func(t *testing.T) {
			value, err := st.ResolveBinaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, symbols.IntegerValue(0), value)
		})
		t.Run("EvaluateBinaryExpression", func(t *testing.T) {
			value, err := st.EvaluateBinaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, &symbols.ExpectedValueObject{symbols.Integer}, value)
		})
	})
	t.Run("will parse binary expression with comparator", func(t *testing.T) {
		node, err := ast.ParseBinaryExpression(textParser("> Bar"), *left)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveBinaryExpression", func(t *testing.T) {
			value, err := st.ResolveBinaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, symbols.BooleanValue(true), value)
		})
		t.Run("EvaluateBinaryExpression", func(t *testing.T) {
			value, err := st.EvaluateBinaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, &symbols.ExpectedValueObject{symbols.Boolean}, value)
		})
	})
	t.Run("will parse binary expression with different operand types that can be operated on", func(t *testing.T) {
		node, err := ast.ParseBinaryExpression(textParser("+ Baz"), *left)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveBinaryExpression", func(t *testing.T) {
			value, err := st.ResolveBinaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, symbols.IntegerValue(2), value)
		})
		t.Run("EvaluateBinaryExpression", func(t *testing.T) {
			value, err := st.EvaluateBinaryExpression(*node)
			if err != nil {
				t.Fatal(err)
			}
			assert.EqualValues(t, &symbols.ExpectedValueObject{symbols.Integer}, value)
		})
	})
	t.Run("will reject binary expression if operands are not operable", func(t *testing.T) {
		node, err := ast.ParseBinaryExpression(textParser("+ Lorem"), *left)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveBinaryExpression", func(t *testing.T) {
			_, err := st.ResolveBinaryExpression(*node)
			EqualErrorCode(t, err, errors.UndefinedOperator)
		})
		t.Run("EvaluateBinaryExpression", func(t *testing.T) {
			_, err := st.EvaluateBinaryExpression(*node)
			EqualErrorCode(t, err, errors.UndefinedOperator)
		})
	})
	t.Run("will reject binary expression if operands are not comparable", func(t *testing.T) {
		node, err := ast.ParseBinaryExpression(textParser("> Lorem"), *left)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveBinaryExpression", func(t *testing.T) {
			_, err := st.ResolveBinaryExpression(*node)
			EqualErrorCode(t, err, errors.UndefinedOperator)
		})
		t.Run("EvaluateBinaryExpression", func(t *testing.T) {
			_, err := st.EvaluateBinaryExpression(*node)
			EqualErrorCode(t, err, errors.UndefinedOperator)
		})
	})
	t.Run("will reject binary expression with invalid operator", func(t *testing.T) {
		node, err := ast.ParseBinaryExpression(textParser("= Bar"), *left)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveBinaryExpression", func(t *testing.T) {
			_, err := st.ResolveBinaryExpression(*node)
			EqualErrorCode(t, err, errors.InvalidOperator)
		})
		t.Run("EvaluateBinaryExpression", func(t *testing.T) {
			_, err := st.EvaluateBinaryExpression(*node)
			EqualErrorCode(t, err, errors.InvalidOperator)
		})
	})
	t.Run("will reject binary expression with invalid left operand", func(t *testing.T) {
		left, err := ast.ParseExpression(textParser("String"))
		if err != nil {
			t.Fatal(err)
		}
		node, err := ast.ParseBinaryExpression(textParser("+ Lorem"), *left)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveBinaryExpression", func(t *testing.T) {
			_, err := st.ResolveBinaryExpression(*node)
			EqualErrorCode(t, err, errors.InvalidValueExpression)
		})
		t.Run("EvaluateBinaryExpression", func(t *testing.T) {
			_, err := st.EvaluateBinaryExpression(*node)
			EqualErrorCode(t, err, errors.InvalidValueExpression)
		})
	})
	t.Run("will reject binary expression with invalid right operand", func(t *testing.T) {
		node, err := ast.ParseBinaryExpression(textParser("+ String"), *left)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveBinaryExpression", func(t *testing.T) {
			_, err := st.ResolveBinaryExpression(*node)
			EqualErrorCode(t, err, errors.InvalidValueExpression)
		})
		t.Run("EvaluateBinaryExpression", func(t *testing.T) {
			_, err := st.EvaluateBinaryExpression(*node)
			EqualErrorCode(t, err, errors.InvalidValueExpression)
		})
	})
}

func TestResolveValueExpression(t *testing.T) {
	fooClass := GenericClass{
		Name: "Foo",
		Properties: map[string]symbols.Class{
			"a": symbols.String,
			"b": symbols.Integer,
		},
		PropertyOverrides: symbols.ClassPropertyMap{
			"failedPropertyGetter": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.String,
				Getter: func(val *GenericValue) (symbols.ValueObject, error) {
					return nil, errors.StandardError(-1, "failed property getter")
				},
			}),
		},
		Prototype: symbols.ClassPrototypeMap{
			"upper": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class:     GenericClass{},
				Arguments: []symbols.Class{symbols.String},
				Returns:   symbols.String,
				Handler: func(val *GenericValue, a symbols.StringValue) (symbols.StringValue, error) {
					return symbols.StringValue(strings.ToUpper(string(a))), nil
				},
			}),
			"noReturn": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class:     GenericClass{},
				Arguments: []symbols.Class{symbols.String},
				Returns:   nil,
				Handler: func(val *GenericValue, a symbols.StringValue) error {
					return nil
				},
			}),
			"failedMethod": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class:     GenericClass{},
				Arguments: []symbols.Class{symbols.String},
				Returns:   symbols.String,
				Handler: func(val *GenericValue, a symbols.StringValue) (symbols.StringValue, error) {
					return "", errors.StandardError(-1, "failed method")
				},
			}),
		},
		ClassProperties: symbols.ClassObjectPropertyMap{
			"classMethod": symbols.NewFunction(symbols.FunctionOptions{
				Arguments: []symbols.Class{symbols.String},
				Returns:   symbols.String,
				Handler: func(a symbols.StringValue) (symbols.StringValue, error) {
					return symbols.StringValue("bar" + string(a)), nil
				},
			}),
			"bar": symbols.StringValue("bar"),
		},
	}
	barObject := GenericObject{
		handler: func(key string) (symbols.ScopeValue, error) {
			switch key {
			case "pass":
				return symbols.StringValue("pass"), nil
			case "fail":
				return nil, errors.StandardError(-1, "fail")
			default:
				return nil, nil
			}
		},
	}
	stringArray := symbols.NewArray(symbols.String, 2)
	stringArray.Set(0, symbols.StringValue("abc"))
	stringArray.Set(1, symbols.StringValue("def"))
	st := SymbolTable(map[string]symbols.ScopeValue{
		"String":      symbols.String,
		"stringArray": stringArray,
		"Foo":         fooClass,
		"Bar":         barObject,
	})
	instanceNode, err := ast.ParseInstanceExpression(textParser("Foo{a: \"abc\", b: 0}"))
	if err != nil {
		t.Fatal(err)
	}
	fooValue, err := st.ResolveInstanceExpression(*instanceNode)
	if err != nil {
		t.Fatal(err)
	}
	st.Local["FooValue"] = fooValue

	t.Run("_ValueExpressionMember", func(t *testing.T) {
		t.Run("can access class property", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("Foo.bar"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				value, err := st.ResolveValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.EqualValues(t, symbols.StringValue("bar"), value)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				value, err := st.EvaluateValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.EqualValues(t, &symbols.ExpectedValueObject{symbols.String}, value)
			})
		})
		t.Run("will reject if class property cannot be found", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("Foo.baz"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualErrorCode(t, err, errors.UnknownProperty)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualErrorCode(t, err, errors.UnknownProperty)
			})
		})
		t.Run("can access value object prototype method", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("FooValue.upper"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				value, err := st.ResolveValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.IsType(t, symbols.Function{}, value)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				value, err := st.EvaluateValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.IsType(t, &symbols.ClassMethod{}, value)
			})
		})
		t.Run("can access value object properties", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("FooValue.a"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				value, err := st.ResolveValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.EqualValues(t, symbols.StringValue("abc"), value)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				value, err := st.EvaluateValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.EqualValues(t, &symbols.ExpectedValueObject{symbols.String}, value)
			})
		})
		t.Run("will reject if accessing value object property throws an error", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("FooValue.failedPropertyGetter"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualErrorCode(t, err, -1)
			})
		})
		t.Run("will reject if value object property cannot be found", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("FooValue.c"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualErrorCode(t, err, errors.UnknownProperty)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualErrorCode(t, err, errors.UnknownProperty)
			})
		})
		t.Run("can access object property through getter", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("Bar.pass"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				value, err := st.ResolveValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.EqualValues(t, symbols.StringValue("pass"), value)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				value, err := st.EvaluateValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.EqualValues(t, &symbols.ExpectedValueObject{symbols.String}, value)
			})
		})
		t.Run("will reject if accessing object property throws an error", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("Bar.fail"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualErrorCode(t, err, -1)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualErrorCode(t, err, -1)
			})
		})
		t.Run("will reject if object property cannot be found", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("Bar.baz"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualErrorCode(t, err, errors.UnknownProperty)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualErrorCode(t, err, errors.UnknownProperty)
			})
		})
	})
	t.Run("_ValueExpressionCallMember", func(t *testing.T) {
		t.Run("can call value object prototype method", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("FooValue.upper(\"abc\")"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				value, err := st.ResolveValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, symbols.StringValue("ABC"), value)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				value, err := st.EvaluateValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, &symbols.ExpectedValueObject{symbols.String}, value)
			})
		})
		t.Run("can call class object property function", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("Foo.classMethod(\"abc\")"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				value, err := st.ResolveValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, symbols.StringValue("barabc"), value)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				value, err := st.EvaluateValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, &symbols.ExpectedValueObject{symbols.String}, value)
			})
		})
		t.Run("will reject if there is mismatched argument length", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("Foo.classMethod(\"abc\", \"def\")"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualErrorCode(t, err, errors.InvalidArgumentLength)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualErrorCode(t, err, errors.InvalidArgumentLength)
			})
		})
		t.Run("will reject if there is mismatched argument type", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("Foo.classMethod(FooValue)"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualErrorCode(t, err, errors.CannotConstruct)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualErrorCode(t, err, errors.CannotConstruct)
			})
		})
		t.Run("will reject if the target callable returns an error", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("FooValue.failedMethod(\"abc\")"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualStandardError(t, err, -1, "failed method")
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualStandardError(t, err, -1, "failed method")
			})
		})
		t.Run("can construct class", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("String(\"abc\")"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				value, err := st.ResolveValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, symbols.StringValue("abc"), value)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				value, err := st.EvaluateValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, symbols.ExpectedValueObject{symbols.String}, value)
			})
		})
		t.Run("will reject if there is not one argument in class construction", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("String()"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualErrorCode(t, err, errors.InvalidClassConstruction)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualErrorCode(t, err, errors.InvalidClassConstruction)
			})
		})
		t.Run("will reject if the class cannot be constructed from the argument", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("String(FooValue)"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualStandardError(t, err, errors.CannotConstruct, "cannot construct String from Foo")
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualStandardError(t, err, errors.CannotConstruct, "cannot construct String from Foo")
			})
		})

		t.Run("will reject if the target is not callable", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("FooValue.a(\"abc\")"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualErrorCode(t, err, errors.InvalidCallExpression)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualErrorCode(t, err, errors.InvalidCallExpression)
			})
		})
	})
	t.Run("_ValueExpression", func(t *testing.T) {
		t.Run("will reject if the selector cannot be found", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("Baz"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualErrorCode(t, err, errors.UnknownSelector)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualErrorCode(t, err, errors.UnknownSelector)
			})
		})
		t.Run("can take index of enumerable class", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("Foo.bar[0]"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				value, err := st.ResolveValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, symbols.StringValue("a"), value)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				value, err := st.EvaluateValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, symbols.ExpectedValueObject{symbols.String}, value)
			})
		})
		t.Run("can take index of array class", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("stringArray[0]"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				value, err := st.ResolveValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, symbols.StringValue("a"), value)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				value, err := st.EvaluateValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, symbols.ExpectedValueObject{symbols.String}, value)
			})
		})
		t.Run("can take range of enumerable class", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("Foo.bar[0:1]"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				value, err := st.ResolveValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, symbols.StringValue("ab"), value)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				value, err := st.EvaluateValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, symbols.ExpectedValueObject{stringArray.Class()}, value)
			})
		})
		t.Run("can take range of array class", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("stringArray[0:1]"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				value, err := st.ResolveValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, stringArray, value)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				value, err := st.EvaluateValueExpression(*node)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, symbols.ExpectedValueObject{stringArray.Class()}, value)
			})
		})
		t.Run("will reject if the target in an index expression is not enumerable", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("FooValue.b[0]"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualErrorCode(t, err, errors.InvalidIndexTarget)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualErrorCode(t, err, errors.InvalidIndexTarget)
			})
		})
		t.Run("will reject if the ending value expression is not a value object", func(t *testing.T) {
			node, err := ast.ParseValueExpression(textParser("String"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("ResolveValueExpression", func(t *testing.T) {
				_, err := st.ResolveValueExpression(*node)
				EqualErrorCode(t, err, errors.InvalidValueExpression)
			})
			t.Run("EvaluateValueExpression", func(t *testing.T) {
				_, err := st.EvaluateValueExpression(*node)
				EqualErrorCode(t, err, errors.InvalidValueExpression)
			})
		})
	})
}

func TestIndexExpression(t *testing.T) {
	st := SymbolTable(map[string]symbols.ScopeValue{
		"String": symbols.String,
		"Bar": GenericClass{
			ClassProperties: symbols.ClassObjectPropertyMap{
				"index": symbols.NewFunction(symbols.FunctionOptions{
					Arguments: []symbols.Class{},
					Returns:   symbols.Integer,
					Handler: func() (symbols.IntegerValue, error) {
						return symbols.IntegerValue(1), nil
					},
				}),
			},
		},
	})
	t.Run("can take index of value object", func(t *testing.T) {
		node, err := ast.ParseIndexExpression(textParser("[0]"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveIndexExpression", func(t *testing.T) {
			startIndex, endIndex, err := st.ResolveIndexExpression(*node, symbols.StringValue("foo"))
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, symbols.IntegerValue(0), startIndex)
			assert.Equal(t, symbols.IntegerValue(-1), endIndex)
		})
		t.Run("EvaluateIndexExpression", func(t *testing.T) {
			err := st.EvaluateIndexExpression(*node, &symbols.ExpectedValueObject{symbols.String})
			if err != nil {
				t.Fatal(err)
			}
		})
	})
	t.Run("will reject if the target in an index expression is not enumerable", func(t *testing.T) {
		node, err := ast.ParseIndexExpression(textParser("[0]"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveIndexExpression", func(t *testing.T) {
			_, _, err := st.ResolveIndexExpression(*node, symbols.BooleanValue(false))
			EqualErrorCode(t, err, errors.InvalidIndexTarget)
		})
		t.Run("EvaluateIndexExpression", func(t *testing.T) {
			err := st.EvaluateIndexExpression(*node, &symbols.ExpectedValueObject{symbols.Boolean})
			EqualErrorCode(t, err, errors.InvalidIndexTarget)
		})
	})
	t.Run("will reject if the index is not a value object", func(t *testing.T) {
		node, err := ast.ParseIndexExpression(textParser("[String]"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveIndexExpression", func(t *testing.T) {
			_, _, err := st.ResolveIndexExpression(*node, symbols.StringValue("foo"))
			EqualErrorCode(t, err, errors.InvalidValueExpression)
		})
		t.Run("EvaluateIndexExpression", func(t *testing.T) {
			err := st.EvaluateIndexExpression(*node, &symbols.ExpectedValueObject{symbols.String})
			EqualErrorCode(t, err, errors.InvalidValueExpression)
		})
	})
	t.Run("will reject if the index is not an integer", func(t *testing.T) {
		node, err := ast.ParseIndexExpression(textParser("[\"baz\"]"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveIndexExpression", func(t *testing.T) {
			_, _, err := st.ResolveIndexExpression(*node, symbols.StringValue("foo"))
			EqualErrorCode(t, err, errors.InvalidIndex)
		})
		t.Run("EvaluateIndexExpression", func(t *testing.T) {
			err := st.EvaluateIndexExpression(*node, &symbols.ExpectedValueObject{symbols.String})
			EqualErrorCode(t, err, errors.InvalidIndex)
		})
	})
	t.Run("can take range of value object", func(t *testing.T) {
		node, err := ast.ParseIndexExpression(textParser("[0:1]"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveIndexExpression", func(t *testing.T) {
			startIndex, endIndex, err := st.ResolveIndexExpression(*node, symbols.StringValue("foo"))
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, symbols.IntegerValue(0), startIndex)
			assert.Equal(t, symbols.IntegerValue(1), endIndex)
		})
		t.Run("EvaluateIndexExpression", func(t *testing.T) {
			err := st.EvaluateIndexExpression(*node, &symbols.ExpectedValueObject{symbols.String})
			if err != nil {
				t.Fatal(err)
			}
		})
	})
	t.Run("will reject if end index is not a value object", func(t *testing.T) {
		node, err := ast.ParseIndexExpression(textParser("[0:String]"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveIndexExpression", func(t *testing.T) {
			_, _, err := st.ResolveIndexExpression(*node, symbols.StringValue("foo"))
			EqualErrorCode(t, err, errors.InvalidValueExpression)
		})
		t.Run("EvaluateIndexExpression", func(t *testing.T) {
			err := st.EvaluateIndexExpression(*node, &symbols.ExpectedValueObject{symbols.String})
			EqualErrorCode(t, err, errors.InvalidValueExpression)
		})
	})
	t.Run("will reject if end index is not an integer", func(t *testing.T) {
		node, err := ast.ParseIndexExpression(textParser("[0:\"baz\"]"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveIndexExpression", func(t *testing.T) {
			_, _, err := st.ResolveIndexExpression(*node, symbols.StringValue("foo"))
			EqualErrorCode(t, err, errors.InvalidIndex)
		})
		t.Run("EvaluateIndexExpression", func(t *testing.T) {
			err := st.EvaluateIndexExpression(*node, &symbols.ExpectedValueObject{symbols.String})
			EqualErrorCode(t, err, errors.InvalidIndex)
		})
	})
	t.Run("can take range of value object without explicit start index", func(t *testing.T) {
		node, err := ast.ParseIndexExpression(textParser("[:Bar.index()]"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveIndexExpression", func(t *testing.T) {
			startIndex, endIndex, err := st.ResolveIndexExpression(*node, symbols.StringValue("foo"))
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, symbols.IntegerValue(0), startIndex)
			assert.Equal(t, symbols.IntegerValue(1), endIndex)
		})
		t.Run("EvaluateIndexExpression", func(t *testing.T) {
			err := st.EvaluateIndexExpression(*node, &symbols.ExpectedValueObject{symbols.String})
			if err != nil {
				t.Fatal(err)
			}
		})
	})
	t.Run("can take range of value object without explicit end index", func(t *testing.T) {
		node, err := ast.ParseIndexExpression(textParser("[Bar.index():]"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run("ResolveIndexExpression", func(t *testing.T) {
			startIndex, endIndex, err := st.ResolveIndexExpression(*node, symbols.StringValue("foo"))
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, symbols.IntegerValue(1), startIndex)
			assert.Equal(t, symbols.IntegerValue(2), endIndex)
		})
		t.Run("EvaluateIndexExpression", func(t *testing.T) {
			err := st.EvaluateIndexExpression(*node, &symbols.ExpectedValueObject{symbols.String})
			if err != nil {
				t.Fatal(err)
			}
		})
	})
}

// ResolveIndexExpression
// EvaluateIndexExpression
