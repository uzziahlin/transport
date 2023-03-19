package compress

type Type uint8

const (
	GZIP Type = iota + 1
	ZIP
)

type Compressor interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
	Code() uint8
}
