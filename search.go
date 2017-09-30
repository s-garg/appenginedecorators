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

// Search returns a Decorator which will call the underlying handler and then search for a specified term in the
// http response. The 'searchable' attributes are passed in the function definition. The search term is read from the
// 'search' query paramater.
//
// Note: While this approach might be used in some cases, in most cases using the search support of app engine
// would be more efficient.
func Search(attribute []string) Decorator {
	return func(h Handler) Handler {
		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
			response, serverError := h.Do(r, ps, username)
			if serverError != nil {
				return response, serverError
			}
			queryValues := r.URL.Query()
			searchTerm := queryValues.Get("search")
			searchTerm = strings.TrimSpace(searchTerm)
			if searchTerm == "" {
				return response, nil
			}
			input := reflect.ValueOf(response)
			array := reflect.MakeSlice(reflect.TypeOf(response), 0, 10)
			for i := 0; i < input.Len(); i++ {
				var res string
				for j := 0; j < len(attribute); j++ {
					res = input.Index(i).Elem().FieldByName(attribute[j]).Interface().(string)
					if strings.Contains(res, searchTerm) == true && searchTerm != "" {
						array = reflect.Append(array, (input).Index(i))
					}
				}
			}
			return array.Interface(), nil
		})
	}
}
