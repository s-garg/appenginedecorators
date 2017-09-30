package core

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// A ServerError has details of an error
type ServerError struct {
	When   time.Time 
	What   string    
	Who    string    
	Code   ErrorCode 
	Source error     
	Stack  string    
}

func (e ServerError) Error() string {
	if e.Source == nil {
		return fmt.Sprintf("%v: Error %v for user '%v' -- %v", e.When, e.Code, e.Who, e.What)
	} else {
		return fmt.Sprintf("%v: Error %v for user '%v' -- %v -- %v ", e.When, e.Code, e.Who, e.What, e.Source.Error())
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

var _userTypes = [...]string{
	"MissingErrorCode",
	"BadRequest",
	"Unauthorized",
	"Forbidden",
	"NotFound",
}

var _httpErrorCodes = [...]int{
	500,
	400,
	401,
	403,
	404,
}

func (ut ErrorCode) String() string { return _userTypes[ut] }

func (ut ErrorCode) HttpErrorCode() int { return _httpErrorCodes[ut] }

func NewServerError(what string, who string, code ErrorCode, err error) *ServerError {

	if who == "" {
		who = "unknown"
	}
	var stack [4096]byte
	runtime.Stack(stack[:], false)
	return &ServerError{
		When:   time.Now(),
		What:   what,
		Who:    who,
		Code:   code,
		Source: err,
		Stack:  strings.Replace(fmt.Sprintf("%s", stack[:]), "\u0000", "", -1),
	}
}
