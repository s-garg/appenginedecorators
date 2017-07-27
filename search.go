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

// Search Decorator takes an array of attributes against which search paramater will be matched and it returns enhanced resposne filtered bt search criteria
// e.g. localhost:8080/_a/assemblies?search="MCIL"

func Search(attribute []string) Decorator {
	return func(h Handler) Handler {

		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {

			// It takes a handler as param and ps httprouter.Params
			// e.g. localhost:8080/_a/assemblies?search="MCIL"

			ctx := appengine.NewContext(r)

			// appengine contxt derived from request

			defer func(t time.Time) {
				DebugMsg(ctx, fmt.Sprintf("--- Time Elapsed for search decorator : %v ---\n", time.Since(t)))
			}(time.Now())

			// Shows time elapsed in the Pagination Decorator

			response, serverError := h.Do(r, ps, username)
			if serverError != nil {
				return response, serverError
			}

			// here we get resposne of the handler on which decorator will be applied

			search := getSearchFromQuery(r, username)

			// getSearchFromQuery will give us array of attrinbutes on which search will trigger

			search = strings.TrimSpace(search)

			//trims empty spaces from search string

			if search == "" {
				return response, nil
			}

			// If search string is  empty return handler response with no enhancement
			// e.g. localhost:8080/_a/assemblies

			limitedResult := limitBySearch(response, attribute, search)

			//limitBySearch will get limit results from resposnes after applying search decorator

			return limitedResult.Interface(), nil

			// enhanced resposne from search decorator

		})
	}
}

// limitBySearch limits the results in response by mathcing search params aginst attribute array
// attributes are basically feild name

func limitBySearch(response interface{}, attribute []string, search string) reflect.Value {

	input := reflect.ValueOf(response) // got the input

	// We used reflect package from Go to make pagination Generic
	//reflect.ValueOf returns a new Value initialized to the concrete value
	// stored in the interface i.  ValueOf(nil) returns the zero Value.

	array := reflect.MakeSlice(reflect.TypeOf(response), 0, 10) // created a slice value
	// we create a slcie of imput response

	for i := 0; i < input.Len(); i++ {
		var res string
		for j := 0; j < len(attribute); j++ {
			res = input.Index(i).Elem().FieldByName(attribute[j]).Interface().(string)
			if strings.Contains(res, search) == true && search != "" {
				array = reflect.Append(array, (input).Index(i))
			}
		}
	}

	//For the lenght of resposne we loop through the array of all the attributes of each resposne object to match the search params and
	// in the end we append the results in array

	return array
}

// getSearchFromQuery to get limit from params

func getSearchFromQuery(r *http.Request, username string) string {
	queryValues := r.URL.Query()

	// queryValues will have value for query params

	search := queryValues.Get("search")

	//   queryValues.Get gets the value of parameter in this case parameter is search

	return search
}
