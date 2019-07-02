// json package for KillOJ
package kjson

import "encoding/json"

func MarshalString(v interface{}) (res string, err error) {
	var byteRes []byte
	byteRes, err = json.Marshal(v)
	return string(byteRes), err
}

func UnmarshalString(data string, v interface{}) error {
	byteData := []byte(data)
	return json.Unmarshal(byteData, v)
}
