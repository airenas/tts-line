package utils

import "strconv"

//RequestTypeEnum represent possible request types
type RequestTypeEnum int

const (
	//RequestOriginal value
	RequestOriginal RequestTypeEnum = iota + 1
	//RequestNormalized value - text after number replace
	RequestNormalized
	//RequestCleaned value - text after clean, normalize and url replace
	RequestCleaned
	//RequestUser value - data normalized by user
	RequestUser
	//RequestOriginalSSML value
	RequestOriginalSSML
)

func (e RequestTypeEnum) String() string {
	if e < RequestOriginal || e > RequestOriginalSSML {
		return "RequestType:" + strconv.Itoa(int(e))
	}
	return [...]string{"", "original", "normalized", "cleaned", "user", "originalSSML"}[e]
}
