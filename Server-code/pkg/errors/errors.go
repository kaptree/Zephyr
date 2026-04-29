package errors

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user is inactive")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("invalid token")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrNoteNotFound       = errors.New("note not found")
	ErrTagNotFound        = errors.New("tag not found")
	ErrTemplateNotFound   = errors.New("template not found")
	ErrDepartmentNotFound = errors.New("department not found")
	ErrGroupNotFound      = errors.New("group not found")
	ErrRoomNotFound       = errors.New("collaboration room not found")
	ErrDuplicateUsername  = errors.New("username already exists")
	ErrTagInUse           = errors.New("tag is in use and cannot be deleted")
	ErrInvalidOperation   = errors.New("invalid operation")
	ErrLoginLocked        = errors.New("account locked due to too many login attempts")
)

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func WrapError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
