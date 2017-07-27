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

// HandlerWrapper is used to wrap the http requests, we take handler and its corresponding path with the type of method for e.g. POST,PUT, GET, DELETE in params

func HandlerWrapper(handler Handler, path string, method string, gaTrackingId string,
	isDev bool) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// Here we have taken additional params for Google Analytics (GA) tracning to be used as  decorator
	// isDev bool is the check for Development enviroment used in GA Decorator

	handlerWithMetrics := Decorate(
		handler,
		Metrics(fmt.Sprintf("%s_%s", method, path), gaTrackingId, isDev),
	)

	// here we bind our first decorator (GA Metrics) with the handler
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		// the func wll return HTTP Resposnes,request and params

		ctx := appengine.NewContext(r)

		// From HTTP request r, we extracted appengine context
		// NewContext returns a context for an in-flight HTTP request.

		defer func(t time.Time) {
			DebugMsg(ctx, fmt.Sprintf("--- Time Elapsed for path '%s' : %v ---\n", path, time.Since(t)))
		}(time.Now())

		// Debug Message for time elapsed for the request

		response, serverError := handlerWithMetrics.Do(r, ps, "")

		// handlerWithMetrics (Decorate) will return a resposne and server Error

		if serverError != nil {
			if isDev {
				var stack [4096]byte
				runtime.Stack(stack[:], false)
				serverError.Stack = strings.Replace(fmt.Sprintf("%s", stack[:]), "\u0000", "", -1)
			}
			ErrorJson(w, r, serverError)

            //  ErrorJson allows us to control the format of the output based on the error code (err.code()) we can show error
			//   NewServerError(err.Error(), username,  ErrorCode, err) ,  we can control Errorcode based on response

		} else {
			err := JsonResponse(w, response)

			// JsonResponse sets content content Type as application/json and encodes it to Json format

			if err != nil {
				ErrorJson(w, r, NewServerError(fmt.Sprintf("Failed to encode to json for request: %v",
					r), "", MissingErrorCode, err))
			}
			// If encoding fails we return Server Error
		}
	}
}
