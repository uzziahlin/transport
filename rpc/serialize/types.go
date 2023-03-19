package serialize

type Serializer interface {
	Serializable
	Deserializable
	Code() uint8
}

type Serializable interface {
	Serialize(src any) ([]byte, error)
}

type Deserializable interface {
	Deserialize(data []byte, dest any) error
}
