package dasel

import (
	"reflect"
	"testing"
)

func assertErrResult(t *testing.T, expErr error, gotErr error) bool {
	if expErr == nil && gotErr != nil {
		t.Errorf("expected err %v, got %v", expErr, gotErr)
		return false
	}
	if expErr != nil && gotErr == nil {
		t.Errorf("expected err %v, got %v", expErr, gotErr)
		return false
	}
	if expErr != nil && gotErr != nil && gotErr.Error() != expErr.Error() {
		t.Errorf("expected err %v, got %v", expErr, gotErr)
		return false
	}
	return true
}

func assertQueryResult(t *testing.T, exp reflect.Value, expErr error, got reflect.Value, gotErr error) bool {
	if !assertErrResult(t, expErr, gotErr) {
		return false
	}

	var res bool
	if exp.IsValid() && got.IsValid() {
		res = reflect.DeepEqual(exp.Interface(), got.Interface())
	} else {
		res = reflect.DeepEqual(exp, got)
	}

	if !res {
		t.Errorf("expected result %v, got %v", exp, got)
		return false
	}
	return true
}

func getNodeWithValue(value interface{}) *Node {
	rootNode := New(value)
	nextNode := &Node{
		Previous: rootNode,
	}
	rootNode.Next = nextNode
	return nextNode
}

func TestFindValueProperty(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		n := getNodeWithValue(nil)
		n.Previous.Selector.Current = "."
		got, err := findValueProperty(n, false)
		assertQueryResult(t, nilValue(), &UnexpectedPreviousNilValue{Selector: "."}, got, err)
	})
	t.Run("NotFound", func(t *testing.T) {
		n := getNodeWithValue(map[string]interface{}{})
		n.Selector.Current = "x"
		got, err := findValueProperty(n, false)
		assertQueryResult(t, nilValue(), &ValueNotFound{Selector: n.Selector.Current, PreviousValue: n.Previous.Value}, got, err)
	})
	t.Run("UnsupportedType", func(t *testing.T) {
		val := 0
		n := getNodeWithValue(val)
		n.Selector.Current = "x"
		got, err := findValueProperty(n, false)
		assertQueryResult(t, nilValue(), &UnsupportedTypeForSelector{Selector: n.Selector, Value: reflect.ValueOf(val)}, got, err)
	})
}

func TestFindValueIndex(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		n := getNodeWithValue(nil)
		n.Previous.Selector.Current = "."
		got, err := findValueIndex(n, false)
		assertQueryResult(t, nilValue(), &UnexpectedPreviousNilValue{Selector: "."}, got, err)
	})
	t.Run("NotFound", func(t *testing.T) {
		n := getNodeWithValue([]interface{}{})
		n.Selector.Current = "[0]"
		n.Selector.Index = 0
		got, err := findValueIndex(n, false)
		assertQueryResult(t, nilValue(), &ValueNotFound{Selector: n.Selector.Current, PreviousValue: n.Previous.Value}, got, err)
	})
	t.Run("UnsupportedType", func(t *testing.T) {
		val := map[string]interface{}{}
		n := getNodeWithValue(val)
		n.Selector.Current = "[0]"
		n.Selector.Index = 0
		got, err := findValueIndex(n, false)
		assertQueryResult(t, nilValue(), &UnsupportedTypeForSelector{Selector: n.Selector, Value: reflect.ValueOf(val)}, got, err)
	})
}

func TestFindValueNextAvailableIndex(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		n := getNodeWithValue([]interface{}{})
		n.Selector.Current = "[0]"
		n.Selector.Index = 0
		got, err := findNextAvailableIndex(n, false)
		assertQueryResult(t, nilValue(), &ValueNotFound{Selector: n.Selector.Current, PreviousValue: n.Previous.Value}, got, err)
	})
}

func TestFindValueDynamic(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		n := getNodeWithValue(nil)
		n.Previous.Selector.Current = "."
		got, err := findValueDynamic(n, false)
		assertQueryResult(t, nilValue(), &UnexpectedPreviousNilValue{Selector: "."}, got, err)
	})
	t.Run("NotFound", func(t *testing.T) {
		n := getNodeWithValue([]interface{}{})
		n.Selector.Current = "(name=x)"
		n.Selector.Conditions = []Condition{
			&EqualCondition{Key: "name", Value: "x"},
		}
		got, err := findValueDynamic(n, false)
		assertQueryResult(t, nilValue(), &ValueNotFound{Selector: n.Selector.Current, PreviousValue: n.Previous.Value}, got, err)
	})
	t.Run("NotFoundMap", func(t *testing.T) {
		n := getNodeWithValue(map[string]interface{}{})
		n.Selector.Current = "(name=x)"
		n.Selector.Conditions = []Condition{
			&EqualCondition{Key: "name", Value: "x"},
		}
		got, err := findValueDynamic(n, false)
		assertQueryResult(t, nilValue(), &ValueNotFound{Selector: n.Selector.Current, PreviousValue: n.Previous.Value}, got, err)
	})
	t.Run("NotFoundWithCreate", func(t *testing.T) {
		n := getNodeWithValue([]interface{}{})
		n.Selector.Current = "(name=x)"
		n.Selector.Conditions = []Condition{
			&EqualCondition{Key: "name", Value: "x"},
		}
		got, err := findValueDynamic(n, true)
		if !assertQueryResult(t, nilValue(), nil, got, err) {
			return
		}
		if exp, got := "NEXT_AVAILABLE_INDEX", n.Selector.Type; exp != got {
			t.Errorf("expected type of %s, got %s", exp, got)
			return
		}
	})
	t.Run("UnsupportedType", func(t *testing.T) {
		val := 0
		n := getNodeWithValue(val)
		n.Selector.Current = "(name=x)"
		n.Selector.Conditions = []Condition{
			&EqualCondition{Key: "name", Value: "x"},
		}
		got, err := findValueDynamic(n, false)
		assertQueryResult(t, nilValue(), &UnsupportedTypeForSelector{Selector: n.Selector, Value: reflect.ValueOf(val)}, got, err)
	})
}

func TestFindValueLength(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		n := getNodeWithValue(nil)
		n.Previous.Selector.Current = ".[#]"
		got, err := findValueLength(n, false)
		assertQueryResult(t, nilValue(), &UnexpectedPreviousNilValue{Selector: ".[#]"}, got, err)
	})
	t.Run("UnsupportedTypeInt", func(t *testing.T) {
		val := 0
		n := getNodeWithValue(val)
		n.Selector.Current = ".[#]"
		got, err := findValueLength(n, false)
		assertQueryResult(t, nilValue(), &UnsupportedTypeForSelector{Selector: n.Selector, Value: reflect.ValueOf(val)}, got, err)
	})
	t.Run("UnsupportedTypeBool", func(t *testing.T) {
		val := false
		n := getNodeWithValue(val)
		n.Selector.Current = ".[#]"
		got, err := findValueLength(n, false)
		assertQueryResult(t, nilValue(), &UnsupportedTypeForSelector{Selector: n.Selector, Value: reflect.ValueOf(val)}, got, err)
	})
	t.Run("SliceType", func(t *testing.T) {
		n := getNodeWithValue([]interface{}{"x", "y"})
		n.Previous.Selector.Current = ".[#]"
		got, err := findValueLength(n, false)
		assertQueryResult(t, reflect.ValueOf(2), nil, got, err)
	})
	t.Run("MapType", func(t *testing.T) {
		n := getNodeWithValue(map[string]interface{}{
			"x": 1,
			"y": 2,
		})
		n.Previous.Selector.Current = ".[#]"
		got, err := findValueLength(n, false)
		assertQueryResult(t, reflect.ValueOf(2), nil, got, err)
	})
	t.Run("StringType", func(t *testing.T) {
		n := getNodeWithValue("hello")
		n.Previous.Selector.Current = ".[#]"
		got, err := findValueLength(n, false)
		assertQueryResult(t, reflect.ValueOf(5), nil, got, err)
	})
}

func TestFindValueType(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		n := getNodeWithValue(nil)
		n.Previous.Selector.Current = ".[#]"
		got, err := findValueType(n, false)
		assertQueryResult(t, nilValue(), &UnexpectedPreviousNilValue{Selector: ".[#]"}, got, err)
	})
	t.Run("Int", func(t *testing.T) {
		val := 0
		n := getNodeWithValue(val)
		n.Selector.Current = ".[#]"
		got, err := findValueType(n, false)
		assertQueryResult(t, reflect.ValueOf("int"), nil, got, err)
	})
	t.Run("Float", func(t *testing.T) {
		val := 1.1
		n := getNodeWithValue(val)
		n.Selector.Current = ".[#]"
		got, err := findValueType(n, false)
		assertQueryResult(t, reflect.ValueOf("float"), nil, got, err)
	})
	t.Run("Bool", func(t *testing.T) {
		val := true
		n := getNodeWithValue(val)
		n.Selector.Current = ".[#]"
		got, err := findValueType(n, false)
		assertQueryResult(t, reflect.ValueOf("bool"), nil, got, err)
	})
	t.Run("String", func(t *testing.T) {
		val := "x"
		n := getNodeWithValue(val)
		n.Selector.Current = ".[#]"
		got, err := findValueType(n, false)
		assertQueryResult(t, reflect.ValueOf("string"), nil, got, err)
	})
	t.Run("Map", func(t *testing.T) {
		val := map[string]interface{}{"x": 1}
		n := getNodeWithValue(val)
		n.Selector.Current = ".[#]"
		got, err := findValueType(n, false)
		assertQueryResult(t, reflect.ValueOf("map"), nil, got, err)
	})
	t.Run("Array", func(t *testing.T) {
		val := []interface{}{1}
		n := getNodeWithValue(val)
		n.Selector.Current = ".[#]"
		got, err := findValueType(n, false)
		assertQueryResult(t, reflect.ValueOf("array"), nil, got, err)
	})
}

func TestFindValue(t *testing.T) {
	t.Run("MissingPreviousNode", func(t *testing.T) {
		n := New(nil)
		got, err := findValue(n, false)
		assertQueryResult(t, nilValue(), ErrMissingPreviousNode, got, err)
	})
	t.Run("UnsupportedSelector", func(t *testing.T) {
		n := getNodeWithValue([]interface{}{})
		n.Selector.Raw = "BAD"
		got, err := findValue(n, false)
		assertQueryResult(t, nilValue(), &UnsupportedSelector{Selector: "BAD"}, got, err)
	})
}
