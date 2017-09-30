package core

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"reflect"

	"appengine"
	"fmt"
	"strings"
	"time"
)

// Search decorator takes an array of attributes against which search paramater will be matched and it returns enhanced resposne filtered by search criteria
func Search(attribute []string) Decorator {
	return func(h Handler) Handler {
		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
			response, serverError := h.Do(r, ps, username)
			if serverError != nil {
				return response, serverError
			}
			search := getSearchFromQuery(r, username)
			if search == "" {
				return response, nil
			}
			limitedResult := limitBySearch(response, attribute, search)
			return limitedResult.Interface(), nil
		})
	}
}

// limitBySearch limits the results in response by mathcing search params aginst attribute array
// attributes are basically feild name
func limitBySearch(response interface{}, attribute []string, search string) reflect.Value {
	input := reflect.ValueOf(response)
	// We used reflect package from Go to make pagination Generic
	//reflect.ValueOf returns a new Value initialized to the concrete value
	// stored in the interface i.  ValueOf(nil) returns the zero Value.
	array := reflect.MakeSlice(reflect.TypeOf(response), 0, 10)
	for i := 0; i < input.Len(); i++ {
		var res string
		for j := 0; j < len(attribute); j++ {
			res = input.Index(i).Elem().FieldByName(attribute[j]).Interface().(string)
			if strings.Contains(res, search) == true && search != "" {
				array = reflect.Append(array, (input).Index(i))
			}
		}
	}
	return array
}

// getSearchFromQuery to get search term from query params
func getSearchFromQuery(r *http.Request, username string) string {
	queryValues := r.URL.Query()
	search := queryValues.Get("search")
	search = strings.TrimSpace(search)
	return search
}
