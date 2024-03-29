// evaluator/evaluator_test.go

package evaluator

import (
	"github.com/solbero/monkey/lexer"
	"github.com/solbero/monkey/object"
	"github.com/solbero/monkey/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30}, // 2 * (15)
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},                // 3 * (9) + 10
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50}, // (5 + 20 + 5) * 2 + -10
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		checkIntegerObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10}, // 1 is truthy
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			checkIntegerObject(t, evaluated, int64(integer))
		} else {
			checkNullObject(t, evaluated)
		}
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{`"hello" == "hello"`, true},
		{`"hello" == "goodbye"`, false},
		{`"hello" != "hello"`, false},
		{`"hello" != "goodbye"`, true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		checkBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false}, // 5 is truthy
		{"!!true", true},
		{"!!false", false},
		{"!!5", true}, // 5 is truthy
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		checkBooleanObject(t, evaluated, tt.expected)
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10}, // 9; is ignored
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10}, // 9; is ignored
		{"if (10 > 1) {if (10 > 1) {return 10;} return 1}", 10},
		{"let f = fn(x) {return x; x + 10;}; f(10);", 10}, // x + 10; is ignored
		{"let f = fn(x) {let result = x + 10; return result; return 10;}; f(10);", 20}, // return 10; is ignored
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		checkIntegerObject(t, evaluated, tt.expected)
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		checkIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"
	evaluated := testEval(input)

	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function, got %T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters, got %d, want 1", len(fn.Parameters))
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x', got %q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q, got %q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5}, // return statement
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20}, // nested function calls
		{"fn(x) { x; }(5)", 5},                                        // immediately invoked function expression
	}

	for _, tt := range tests {
		checkIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
	let newAdder = fn(x) {
		fn(y) { x + y };
	};

	let addTwo = newAdder(2);
	addTwo(2);
	`
	checkIntegerObject(t, testEval(input), 4)
}

func TestStringLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: `"hello world"`, expected: "hello world"},
		{input: `"hello \"world\""`, expected: "hello \"world\""},
		{input: `"hello\nworld"`, expected: "hello\nworld"},
		{input: `"hello\t\t\tworld"`, expected: "hello\t\t\tworld"},
		{input: `"hello\\world"`, expected: "hello\\world"},
		{input: `"hello\bworld"`, expected: "helloworld"},
		{input: `"Hello" + " " + "World!"`, expected: "Hello World!"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		str, ok := evaluated.(*object.String)
		if !ok {
			t.Fatalf("object is not String, got %T (%+v)", evaluated, evaluated)
		}

		if str.Value != tt.expected {
			t.Errorf("String has wrong value, expected %q got %q", tt.expected, str.Value)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{input: `len("")`, expected: 0},
		{input: `len("four")`, expected: 4},
		{input: `len("hello world")`, expected: 11},
		{input: `len(1)`, expected: "argument to 'len' not supported, got INTEGER"},
		{input: `len("one", "two")`, expected: "wrong number of arguments, got 2, want 1"},
		{input: `first([1, 2, 3])`, expected: 1},
		{input: `first([])`, expected: nil},
		{input: `first(1)`, expected: "argument to 'first' must be ARRAY, got INTEGER"},
		{input: `first([1, 2], [3, 4])`, expected: "wrong number of arguments, got 2, want 1"},
		{input: `last([1, 2, 3])`, expected: 3},
		{input: `last([])`, expected: nil},
		{input: `last(1)`, expected: "argument to 'last' must be ARRAY, got INTEGER"},
		{input: `last([1, 2], [3, 4])`, expected: "wrong number of arguments, got 2, want 1"},
		{input: `rest([1, 2, 3])`, expected: []int64{2, 3}},
		{input: `rest([])`, expected: nil},
		{input: `rest(1)`, expected: "argument to 'rest' must be ARRAY, got INTEGER"},
		{input: `rest([1, 2], [3, 4])`, expected: "wrong number of arguments, got 2, want 1"},
		{input: `push([], 1)`, expected: []int64{1}},
		{input: `push(1, 1)`, expected: "argument to 'push' must be ARRAY, got INTEGER"},
		{input: `push([1, 2], 1, 2)`, expected: "wrong number of arguments, got 3, want 2"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			checkIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error, got %T (%+v)", evaluated, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message, expected %q, got %q", expected, errObj.Message)
			}
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array, got %T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements, got %d", len(result.Elements))
	}

	checkIntegerObject(t, result.Elements[0], 1)
	checkIntegerObject(t, result.Elements[1], 4)
	checkIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{input: "[1, 2, 3][0]", expected: 1},
		{input: "[1, 2, 3][1]", expected: 2},
		{input: "[1, 2, 3][2]", expected: 3},
		{input: "let i = 0; [1][i];", expected: 1},
		{input: "[1, 2, 3][1 + 1];", expected: 3},
		{input: "let myArray = [1, 2, 3]; myArray[2];", expected: 3},
		{input: "let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];", expected: 6},
		{input: "let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]", expected: 2},
		{input: "[1, 2, 3][3]", expected: nil},
		{input: "[1, 2, 3][-1]", expected: nil},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			checkIntegerObject(t, evaluated, int64(integer))
			continue
		}

		checkNullObject(t, evaluated)
	}
}

func TestHashLiterals(t *testing.T) {
	input := `let two = "two";
	{
		"one": 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 / 2,
		4: 4,
		true: 5,
		false: 6
	}`

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("Eval didn't return Hash, got %T (%+v)", evaluated, evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs, got %d", len(result.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}

		checkIntegerObject(t, pair.Value, expectedValue)
	}
}

func TestHashIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{input: `{"foo": 5}["foo"]`, expected: 5},
		{input: `{"foo": 5}["bar"]`, expected: nil},
		{input: `let key = "foo"; {"foo": 5}[key]`, expected: 5},
		{input: `{5: 5}[5]`, expected: 5},
		{input: `{true: 5}[true]`, expected: 5},
		{input: `{false: 5}[false]`, expected: 5},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			checkIntegerObject(t, evaluated, int64(integer))
			continue
		}

		checkNullObject(t, evaluated)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input       string
		expectedMsg string
	}{
		{"5 + true;", "type mismatch: INTEGER + BOOLEAN"},
		{"5 + true; 5;", "type mismatch: INTEGER + BOOLEAN"},
		{"-true", "unknown operator: -BOOLEAN"},
		{"true + false;", "unknown operator: BOOLEAN + BOOLEAN"},
		{"5; true + false; 5", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) { true + false; }", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) { if (10 > 1) { return true + false; } return 1 }", "unknown operator: BOOLEAN + BOOLEAN"},
		{"foobar", "identifier not found: foobar"},
		{`"Hello" - "World!"`, "unknown operator: STRING - STRING"},
		{`{"name": "Monkey"}[fn(x) { x }];`, "unusable as hash key: FUNCTION"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned, got %T (%+v)", evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expectedMsg {
			t.Errorf("wrong error message, expected %q, got %q", tt.expectedMsg, errObj.Message)
		}
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	env := object.NewEnvironment()
	program := p.ParseProgram()
	return Eval(program, env)
}

func checkNullObject(t *testing.T, obj object.Object) bool {
	t.Helper()
	if obj != NULL {
		t.Errorf("object is not NULL, got %T (%+v)", obj, obj)
		return false
	}
	return true
}

func checkIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	t.Helper()
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer, got %T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value, got %d want %d", result.Value, expected)
		return false
	}
	return true
}

func checkBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	t.Helper()
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean, got %T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value, got %t want %t", result.Value, expected)
		return false
	}
	return true
}
