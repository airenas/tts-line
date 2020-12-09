package wav

import (
	"bytes"
	"encoding/binary"
)

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
	res := make([]byte, 4)
	binary.Write(bytes.NewBuffer(res), binary.LittleEndian, data)
	return res
}
