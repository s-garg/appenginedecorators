# appenginedecorators
Useful decorators for App Engine

# Embrace the Interface
## Motivation
Concepts of decorators is highly based on Robert Pike's Statement "If Java and C++ are about types hierarchies and taxonomy Go is about composition"

# Introduction
Decorators: By Optimized use of abstraction, Design pattern and other software engineering principles we can design a generic function, and work with
Go's Philosophy of composability,

In decorators, we try to take small chunk of code and perform a particular operation and putting them back together to work in tandem. By using this we 
can simplify code for various common functionalities to be applied on our business logic. They are simple orthogonal constructs to make our entire logic simpler.


 
## Handler Interface

``` GO
//Handler sends an Http request and receives any type which implements the interface or err in case of failure 
type Handler interface {
	Do(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError)
}
```
## Handler Function

``` GO
//Handler func is a function which implements Handler Interface.
type HandlerFunc func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError)

func (f HandlerFunc) Do(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
	return f(r, ps, username)
}
```
## Decorator Pattern

It is a design pattern that allows behavior to be added to an instance of a type without affecting the behavior of the instances of the 
same type. A function takes a client and returns the client of same type with an enriched behavior, for example Authorization, load balancing,
logging, pagination or search and much more.

``` GO
type Decorator func(h Handler) Handler
//This goes well with the single responsibly principle and the open close principle of software engineering 
//(software entities should be open for extension and closed for modification)
```  
## How to bind this orthogonal codes together?

``` GO
func Decorate(h Handler, ds ...Decorator) Handler {
	decorated := h
	for _, decorate := range ds {
		decorated = decorate(decorated)
	}
	return decorated
}

// We can take a client and one layer of behavior on the client with every loop, in this the client 
// functionality typically our business logic remains unchanged ,this confirms with our open closes principle and separation of concern principle.
```

## Onion Rings Analogy

Consider the client (Handler) as the core of a onion ring and the decorators as the layers of the onion rings which
adds on the core to enhance it without modifying it. (Open Closes Rule!)

![Onion Slice](https://github.com/s-garg/appenginedecorators/blob/master/githubimage.png)

