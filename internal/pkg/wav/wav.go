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

//GetSampleRate return sample rate from header
func GetSampleRate(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data[24:28])
}

//GetSampleRate return sample rate from header
func GetBitsPerSample(data []byte) uint16 {
	return binary.LittleEndian.Uint16(data[34:36])
}

//GetChannels return channels
func GetChannels(data []byte) uint16 {
	return binary.LittleEndian.Uint16(data[22:24])
}

//GetBitsRate returns bit rate from header
func GetBitsRate(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data[28:32])
}

//GetBitsRateCalc calculates bits per sample - sometimes there is incorect value for bitsrate in header
func GetBitsRateCalc(data []byte) uint32 {
	return uint32(GetSampleRate(data) * uint32(GetChannels(data)) * uint32(GetBitsPerSample(data)/8))
}
