package pars

import (
	"math/big"
	"testing"
)

func assertRead(t *testing.T, n int, err error, expectedN int, expectedErr error) {
	if n != expectedN {
		t.Errorf("Expected read count %v, but got %v", expectedN, n)
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, but got %v", expectedErr, err)
	}
}

func assertReader(t *testing.T, r *reader, buf []byte, err error) {
	if r.lastErr != err {
		t.Errorf("Expected lastErr in reader %v, but got %v", err, r.lastErr)
	}

	assertBytes(t, r.buf.current, buf)
}

func assertBufferLen(t *testing.T, buf buffer, expected int) {
	if buf.Len() != expected {
		t.Errorf("Expected buffer to have length %v, but got %v", expected, buf.Len())
	}
}

func assertBytes(t *testing.T, buf []byte, expectedBuf []byte) {
	if len(expectedBuf) != len(buf) {
		t.Errorf("Expected len in reader %v, but got %v, buffer contains %v", len(expectedBuf), len(buf), buf)
		return
	}

	for i, b := range expectedBuf {
		if buf[i] != b {
			t.Errorf("Expected byte %v in reader buffer to be %v, but got %v", i, b, buf[i])
		}
	}
}

func assertError(t *testing.T, err error, expectedErr error) {
	if err != expectedErr && (err == nil || expectedErr == nil || err.Error() != expectedErr.Error()) {
		t.Errorf("\nExpected error '%v' (%T),\n"+
			"       but got '%v' (%T)", expectedErr, expectedErr, err, err)
	}
}

func assertValue(t *testing.T, val interface{}, expectedVal interface{}) {
	if val != expectedVal {
		t.Errorf("Expected %v (%T), but got %v (%T)", expectedVal, expectedVal, val, val)
	}
}

func assertValueSlice(t *testing.T, val interface{}, expectedValues []interface{}) {
	var values []interface{}
	var ok bool
	if values, ok = val.([]interface{}); ok == false {
		t.Errorf("Excepcted %v (%T), but got %v (%T)", expectedValues, expectedValues, val, val)
		return
	}

	for i, v := range values {
		if i >= len(expectedValues) {
			t.Errorf("Index %v: Unexpected element %v (%T)", i, v, v)
		} else {
			expectedVal := expectedValues[i]
			if v != expectedVal {
				t.Errorf("Index %v: Expected %v (%T), but got %v (%T)", i, expectedVal, expectedVal, v, v)
			}
		}
	}

	for i := len(values); i < len(expectedValues); i++ {
		expectedVal := expectedValues[i]
		t.Errorf("Missing element at index %v: %v (%T)", i, expectedVal, expectedVal)
	}
}

func assertValueBigInt(t *testing.T, val interface{}, expectedValStr interface{}) {
	if expectedValStr == nil {
		if val != nil {
			t.Errorf("Expected nil, but got %v (%T)", val, val)
		}
		return
	}

	expectedVal := big.NewInt(0)
	expectedVal.SetString(expectedValStr.(string), 10)

	var valBigInt *big.Int
	var ok bool

	if valBigInt, ok = val.(*big.Int); ok != true {
		t.Errorf("Expected %v (%T), but got %v (%T)", expectedVal, expectedVal, val, val)
		return
	}

	if valBigInt.Cmp(expectedVal) != 0 {
		t.Errorf("Expected %v (%T), but got %v (%T)", expectedVal, expectedVal, val, val)
	}
}

func assertParse(t *testing.T, val interface{}, err error, expectedVal interface{}, expectedErr error) {
	assertValue(t, val, expectedVal)
	assertError(t, err, expectedErr)
}

func assertParseSlice(t *testing.T, val interface{}, err error, expectedValues []interface{}, expectedErr error) {
	assertValueSlice(t, val, expectedValues)
	assertError(t, err, expectedErr)
}

func assertParseBigInt(t *testing.T, val interface{}, err error, expectedVal interface{}, expectedErr error) {
	assertValueBigInt(t, val, expectedVal)
	assertError(t, err, expectedErr)
}

func assertRunesInSlice(t *testing.T, vals []interface{}, expected string) {
	for i, val := range vals {
		if i >= len(expected) {
			t.Errorf("More values (%v) found than expected (%v)", len(vals), len(expected))
			return
		}
		expRune := []rune(expected)[i]
		if r, ok := val.(rune); ok {
			if r != expRune {
				t.Errorf("Expected rune '%c' at index '%v', got '%c'", expRune, i, r)
			}
		} else {
			t.Errorf("Expected rune '%c' at index '%v', got '%v' (%T)", expRune, i, val, val)
		}
	}
}
