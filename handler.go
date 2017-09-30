package core

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// Handler is an interface that processes a http request and returns response as an interface.
type Handler interface {
	Do(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError)
}

// A HandlerFunc coverts an existing func to type Handler
type HandlerFunc func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError)

func (f HandlerFunc) Do(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
	return f(r, ps, username)
}

// A Decorator takes in a Handler, applies some enhancement to the existing handler logic and returns the enhanced handler
type Decorator func(h Handler) Handler

// Decorate takes a Handler and a list of Decorators, and converts them to a new Handler which will go through all the Decorators before executing the original Handler.
func Decorate(h Handler, ds ...Decorator) Handler {
	decorated := h
	for _, decorate := range ds {
		decorated = decorate(decorated)
	}
	return decorated
}
