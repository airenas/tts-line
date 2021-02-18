package utils

import "strconv"

//RequestTypeEnum represent possible request types
type RequestTypeEnum int

const (
	//RequestOriginal value
	RequestOriginal RequestTypeEnum = iota + 1
	//RequestNormalized value
	RequestNormalized
	//RequestCleaned value
	RequestCleaned
	//RequestUser value - data normalized by user
	RequestUser
)

func (e RequestTypeEnum) String() string {
	if e < RequestOriginal || e > RequestUser {
		return "RequestType:" + strconv.Itoa(int(e))
	}
	return [...]string{"", "original", "normalized", "cleaned", "user"}[e]
}
