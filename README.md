# appenginedecorators
Useful decorators for App Engine

# Embrace the Interface
## Motivation
Concepts of decorators is highly based on Robert Pike's Statement "If Java and C++ are about types hierarchies and taxonomy Go is about composition"

# Introduction
Decorators:By Optimised use of abstraction,Design pattern and other software engineering principles we can design a genric functions, and work with
Go'sPhiosopy of composiblity,

In decorators we try to take small chunk of code and perform a particular operation and putting them back togather to work in tandem.By using this we 
can singlify code for various comman functionalities to be applied on our business logic. They are simple orthogonal constructs in order to make
our entire logic simpler.


 
## Handler Interface

``` GO
 // A Client sends an Http request and receivres any type which implements the interface or err in case of failure 

type Handler interface {
	Do(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError)
}
```
## Handler Function

``` GO
// Handler func is a function which implements Handler Interface.
type HandlerFunc func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError)

func (f HandlerFunc) Do(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
	return f(r, ps, username)
}
```
## Decorator Pattern

It is a design pattern that allows behavoiur to be added to an instacne of a type without afftecting the behavour of the instacnes of the 
same type. A function takes a client and returns the client of same type with an enriched behavour,for example Authorixzation ,load balcning,
logging,pagination or search and much more.

``` GO
type Decorator func(h Handler) Handler
//This goes well woth the single resposniblty principle and the open close principle of software engineering 
//(software entitties should be open for extension and closed for modification)
```  
## How to bind this orthogonal codes togather?

``` GO
func Decorate(h Handler, ds ...Decorator) Handler {
	decorated := h
	for _, decorate := range ds {
		decorated = decorate(decorated)
	}
	return decorated
}

// We can take a client and one layer of behavior on the client with every loop, in this the client 
// functioanlity typically our business logic remains unchanged ,this confirms with our open closes princilpe and separation of concern principle.

```


Onion Rings Analogy: Consier the client (Handler) are the core of a onion ring and the decorators as the layers of the onion rings which
adds on the core to enhance it without modification.
