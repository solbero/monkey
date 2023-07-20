// evaluator/evaluator_test.go

package evaluator

import (
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
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
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		checkIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
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

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	return Eval(program)
}

func checkIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	t.Helper()
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer, got %T(%+v)", obj, obj)
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
		t.Errorf("object is not Boolean, got %T(%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value, got %t want %t", result.Value, expected)
		return false
	}
	return true
}
