package main

import (
	"fmt"
	"testing"
)

type encTestObj struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// ----------------------------------------------
// JSONEncoder.Encode TESTS
// ----------------------------------------------

// Tests the "happy path" for the JSONEncoder.Encode() function.
func Test_JSONEncoder_Encode(t *testing.T) {
	o := encTestObj{Name: "my name", Value: 50}
	expected := fmt.Sprintf(`{"name":"%s","value":%d}`, o.Name, o.Value)

	enc := JSONEncoder{}
	actual := enc.Encode(o)
	assert.Equal(t, actual, expected, "JSONEncoder.Encode() returned unexpected value")
}

// ----------------------------------------------
// JSONEncoder.EncodeMulti TESTS
// ----------------------------------------------

// Tests the "happy path" for the JSONEncoder.EncodeMulti() function.
func Test_JSONEncoder_EncodeMulti(t *testing.T) {
	o1 := encTestObj{Name: "my name", Value: 50}
	o2 := encTestObj{Name: "your name", Value: 25}
	col := []interface{}{o1, o2}

	str1 := fmt.Sprintf(`{"name":"%s","value":%d}`, o1.Name, o1.Value)
	str2 := fmt.Sprintf(`{"name":"%s","value":%d}`, o2.Name, o2.Value)
	expected := fmt.Sprintf("[%s,%s]", str1, str2)

	enc := JSONEncoder{}
	actual := enc.EncodeMulti(col...)
	assert.Equal(t, actual, expected, "JSONEncoder.EncodeMulti() returned unexpected value")
}

func Test_JSONEncoder_EncodeMulti_Empty(t *testing.T) {
	col := []interface{}{}
	enc := JSONEncoder{}
	actual := enc.EncodeMulti(col...)
	expected := "[]"
	assert.Equal(t, actual, expected, "JSONEncoder.EncodeMulti() returned unexpected value")
}

func Test_JSONEncoder_EncodeMulti_Nil(t *testing.T) {
	enc := JSONEncoder{}
	actual := enc.EncodeMulti(nil)
	expected := "[]"
	assert.Equal(t, actual, expected, "JSONEncoder.EncodeMulti() returned unexpected value")
}

func Test_JSONEncoder_EncodeMulti_NoValues(t *testing.T) {
	enc := JSONEncoder{}
	actual := enc.EncodeMulti()
	expected := "[]"
	assert.Equal(t, actual, expected, "JSONEncoder.EncodeMulti() returned unexpected value")
}

// ----------------------------------------------
// JSONEncoder.Decode TESTS
// ----------------------------------------------

// Tests the "happy path" for the JSONEncoder.Decode() function.
func Test_JSONEncoder_Decode(t *testing.T) {
	expected := encTestObj{Name: "my name", Value: 50}
	str := fmt.Sprintf(`{"name":"%s","value":%d}`, expected.Name, expected.Value)
	actual := encTestObj{}
	enc := JSONEncoder{}
	_ = enc.Decode([]byte(str), &actual)
	assert.Equal(t, actual, expected, "JSONEncoder.Decode() returned unexpected value")
}
