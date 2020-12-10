package wav

import (
	"bytes"
	"encoding/binary"
)

//IsValid test if wav binary data is OK, at least the size
func IsValid(data []byte) bool {
	return len(data) > 44
}

//TakeHeader copy header
func TakeHeader(data []byte) []byte {
	return data[0:40]
}

//GetSize get data bits size
func GetSize(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data[40:44])
}

//TakeData copy data
func TakeData(data []byte) []byte {
	return data[44:]
}

//SizeBytes return wav size bytes
func SizeBytes(data uint32) []byte {
	res := &bytes.Buffer{}
	binary.Write(res, binary.LittleEndian, data)
	return res.Bytes()
}
