package util

import (
	"encoding/json"
)

func Unmarshal(data []byte, v interface{}) error {
	//dec := msgpack.NewDecoder(bytes.NewReader(data))
	//dec.UseJSONTag(true)
	//return dec.Decode(v)
	return json.Unmarshal(data, v)
}

func Marshal(v interface{}) ([]byte, error) {
	//var buf bytes.Buffer
	//enc := msgpack.NewEncoder(&buf)
	//enc.UseJSONTag(true)
	//err := enc.Encode(v)
	//return buf.Bytes(), err
	return json.Marshal(v)
}
