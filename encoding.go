package main

import (
	"encoding/json"
)

// Encoder is an interface that defines the functions that an encoder should provide.
type Encoder interface {
	Encode(v interface{}) string
	EncodeMulti(v ...interface{}) string
	Decode(b []byte, i interface{}) error
}

// JSONEncoder extends the Encoder interface and provides JSON encoding.
type JSONEncoder struct{}

// Encode attempts to encode any struct into a JSON string.
func (e JSONEncoder) Encode(v interface{}) string {
	var data interface{} = v
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// EncodeMulti converts any collection of structs into a JSON string (array).
func (e JSONEncoder) EncodeMulti(v ...interface{}) string {
	var data interface{} = v
	if v == nil || len(v) == 0 || (len(v) == 1 && v[0] == nil) {
		// so empty results produce '[]' and not 'null'
		data = []interface{}{}
	}
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// Decode loads the specified struct with the given JSON byte array.
func (e JSONEncoder) Decode(b []byte, i interface{}) error {
	err := json.Unmarshal(b, i)
	if err != nil {
		return err
	}
	return nil
}
