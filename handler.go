// package core contains types and methods for appengine decorators
package core

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// A Handler wraps a Do method that can be registered to a route to handle HTTP
// requests and has params for additional variables and queries ,at last username is used to lookup user
// is implemented
type Handler interface {
	Do(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError)
}

// A Handlerfunc wraps a method which implements Handler Interface
type HandlerFunc func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError)

func (f HandlerFunc) Do(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
	return f(r, ps, username)
}

// A Decorator takes in Handler , applies enhancement to the handler, without modifying its existing properties and return the enhanced handler
type Decorator func(h Handler) Handler

// Decorate is used to bind all the decorators together with the handler,with every loop a decorator is applied, which adds it own enhancement
// and it is orthogonal to the other decorators
// Here order of decorators is very important , for e.g. if you want to implement search and pagination, Search decorator should go first and on the result of search decorator
// i.e  handler resposne with searched criteria , you can apply pagination to limit results.
func Decorate(h Handler, ds ...Decorator) Handler {
	decorated := h
	for _, decorate := range ds {
		decorated = decorate(decorated)
	}
	return decorated
}
