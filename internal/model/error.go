package model

type ErrorCode string

const (
	ErrorUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrorForbidden    ErrorCode = "FORBIDDEN"
	ErrorLoginTaken   ErrorCode = "LOGIN_TAKEN"
	ErrorNotFound     ErrorCode = "NOT_FOUND"
	ErrorInternal     ErrorCode = "INTERNAL"
	ErrorBadRequest   ErrorCode = "BAD_REQUEST"
)

type ErrorBody struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}
