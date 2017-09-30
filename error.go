package core

import (
	"fmt"
	"time"
)

// A ServerError has details of an error
type ServerError struct {
	When   time.Time
	What   string
	Who    string
	Code   ErrorCode
	Source error
}

func (e ServerError) Error() string {
	if e.Source == nil {
		return fmt.Sprintf("%v: Error %d for user '%s' -- %s", e.When, e.Code, e.Who, e.What)
	} else {
		return fmt.Sprintf("%v: Error %d for user '%s' -- %s (source: %v)", e.When, e.Code, e.Who, e.What, e.Source.Error())
	}
}

// Error codes for various scenarios
type ErrorCode int32

const (
	MissingErrorCode ErrorCode = iota
	BadRequest
	Unauthorized
	Forbidden
	NotFound
)

var userTypes = [...]string{
	"MissingErrorCode",
	"BadRequest",
	"Unauthorized",
	"Forbidden",
	"NotFound",
}

var httpErrorCodes = [...]int{
	500,
	400,
	401,
	403,
	404,
}

func (ut ErrorCode) String() string { return userTypes[ut] }

func (ut ErrorCode) HttpErrorCode() int { return httpErrorCodes[ut] }

func NewServerError(what string, who string, code ErrorCode, err error) *ServerError {
	if who == "" {
		who = "unknown"
	}
	return &ServerError{
		When:   time.Now(),
		What:   what,
		Who:    who,
		Code:   code,
		Source: err,
	}
}
