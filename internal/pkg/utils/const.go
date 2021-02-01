package utils

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
		return ""
	}
	return [...]string{"original", "normalized"}[e]
}
