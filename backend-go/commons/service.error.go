package commons

type ServiceError struct {
	Code string // Error code
}

func (e *ServiceError) Error() string {
	return e.Code
}
