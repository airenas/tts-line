package utils

//RequestTypeEnum represent possible request types
type RequestTypeEnum int

const (
	//RequestMain value
	RequestMain RequestTypeEnum = iota
)

func (e RequestTypeEnum) String() string {
	return [...]string{"", "main"}[e]
}
