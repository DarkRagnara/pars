package pars

import (
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

func assertReader(t *testing.T, r *Reader, buf []byte, err error) {
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

func assertParse(t *testing.T, val interface{}, err error, expectedVal interface{}, expectedErr error) {
	if val != expectedVal {
		t.Errorf("Expected %v (%T), but got %v (%T)", expectedVal, expectedVal, val, val)
	}

	if err != expectedErr {
		t.Errorf("Expected error %v (%T), but got %v (%T)", expectedErr, expectedErr, err, err)
	}
}
