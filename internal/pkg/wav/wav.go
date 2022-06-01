package wav

import (
	"bytes"
	"encoding/binary"
)

var dataHeader = []byte{'d', 'a', 't', 'a'}
var expectedDataHeaderStart = 36

// IsValid test if wav binary data is OK, at least the size
// and contains data header key
func IsValid(data []byte) bool {
	return getDataHeaderPos(data, expectedDataHeaderStart) > 0
}

// data may start at dynamic position
// try search data location by running through headers
func getDataHeaderPos(data []byte, at int) int {
	if len(data) < at+8 {
		return 0
	}
	if bytes.Equal(data[at:at+4], dataHeader) {
		return at
	}
	hl := getSize(data, at+4)
	return getDataHeaderPos(data, at+8+int(hl))
}

// TakeHeader copy header
func TakeHeader(data []byte) []byte {
	return data[0:36]
}

func getSize(data []byte, at int) uint32 {
	return binary.LittleEndian.Uint32(data[at : at+4])
}

// GetSize get data bits size
func GetSize(data []byte) uint32 {
	dataPos := getDataHeaderPos(data, expectedDataHeaderStart)
	if dataPos > 0 {
		return getSize(data, dataPos+4)
	}
	return 0
}

// TakeData copy data
func TakeData(data []byte) []byte {
	dataPos := getDataHeaderPos(data, expectedDataHeaderStart)
	if dataPos > 0 {
		return data[dataPos+8:]
	}
	return nil
}

// SizeBytes return wav size bytes
func SizeBytes(data uint32) []byte {
	res := &bytes.Buffer{}
	binary.Write(res, binary.LittleEndian, data)
	return res.Bytes()
}

// GetSampleRate return sample rate from header
func GetSampleRate(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data[24:28])
}

// GetBitsPerSample return bits per sample
func GetBitsPerSample(data []byte) uint16 {
	return binary.LittleEndian.Uint16(data[34:36])
}

// GetChannels return channels
func GetChannels(data []byte) uint16 {
	return binary.LittleEndian.Uint16(data[22:24])
}

// GetBitsRate returns bit rate from header
func GetBitsRate(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data[28:32])
}

// GetBitsRateCalc calculates bits per sample - sometimes there is incorect value for bitsrate in header
func GetBitsRateCalc(data []byte) uint32 {
	return uint32(GetSampleRate(data) * uint32(GetChannels(data)) * uint32(GetBitsPerSample(data)/8))
}
