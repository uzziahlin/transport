package json

import "encoding/json"

type Serializer struct {
}

func (s *Serializer) Code() uint8 {
	return 1
}

func (s *Serializer) Serialize(src any) ([]byte, error) {
	return json.Marshal(src)
}

func (s *Serializer) Deserialize(data []byte, dest any) error {
	return json.Unmarshal(data, dest)
}
