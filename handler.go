package core

import (
	"appengine"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// A Client sends an Http request and receives any type which implements the interface or err in case of failure
// Handler is a interface and DO is a method that can be registered to a route to handle HTTP
// requests and has params for additioanl variables and quries ,al last username is used to lookup user

type Handler interface {
	Do(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError)
}

// Handler func is a function which implements Handler Interface.

type HandlerFunc func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError)

//  Handler Func is inturn a receiver which implements Do Method

func (f HandlerFunc) Do(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
	return f(r, ps, username)
}

// Decorator are functions which takes in Handler , applies enhancement to the handler, without modifying its existing properties and return the enhanced handler

type Decorator func(h Handler) Handler

// Decorate function is used to bind all the decorators togather with the handler,with every loop a decorator is applied, which adds it own enhancement
// and it is orthognoal to the other decorators

// Here order of decorators is very important , for e.g. if you want to implement search and pagination, Search decorator should go first and on the result of search decorator
//i.e  handler resposne with searched criteria , you can apply pagination to limit results.

func Decorate(h Handler, ds ...Decorator) Handler {
	decorated := h
	for _, decorate := range ds {
		decorated = decorate(decorated)
	}
	return decorated
}
