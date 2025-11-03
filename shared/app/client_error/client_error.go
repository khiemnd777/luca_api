package client_error

import "fmt"

type ClientError struct {
	StatusCode int
	Err        error
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("client error %d: %v", e.StatusCode, e.Err)
}
