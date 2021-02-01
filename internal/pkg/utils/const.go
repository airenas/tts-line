package utils

import "strconv"

//RequestTypeEnum represent possible request types
type RequestTypeEnum int

const (
	//RequestOriginal value
	RequestOriginal RequestTypeEnum = iota + 1
	//RequestNormalized value
	RequestNormalized
)

func (e RequestTypeEnum) String() string {
	if e < RequestOriginal || e > RequestNormalized {
		return "RequestType:" + strconv.Itoa(int(e))
	}
	return [...]string{"", "original", "normalized"}[e]
}
