package proto

import (
	"errors"
	"github.com/golang/protobuf/proto"
)

type Serializer struct {
}

func (s Serializer) Serialize(src any) ([]byte, error) {
	msg, ok := src.(proto.Message)

	if !ok {
		return nil, errors.New("序列化数据必须实现proto.Message接口")
	}

	return proto.Marshal(msg)
}

func (s Serializer) Deserialize(data []byte, dest any) error {
	msg, ok := dest.(proto.Message)

	if !ok {
		return errors.New("反序列化数据必须实现proto.Message接口")
	}

	return proto.Unmarshal(data, msg)
}

func (s Serializer) Code() uint8 {
	return 2
}
