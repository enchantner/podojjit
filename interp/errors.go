package interp

type RuntimeError struct {
	Err error
}

func (e *RuntimeError) Error() string {
	return e.Err.Error()
}
