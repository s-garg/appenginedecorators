package core

import (
	"appengine"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

// Paginate is a Decorator , its utility is to limit the number of results based on a user set limit and a attribute of type timeStamp,
// results before  the attribute will be displayed
func Paginate(attribute string) Decorator {
	return func(h Handler) Handler {
		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {

			// It takes a handler as param and ps httprouter.Params can have 2 query params limit and timestamp
			// e.g. localhost:8080/_a/assemblies?limit=3&&timestamp=2017-07-25 16:11:46 ,
			// we can set feild in attribute on which timestamp filter will be applied
			ctx := appengine.NewContext(r)
			// getLimitFromQuery gets the value of limit parameter from ps httprouter.Params
			// e.g. localhost:8080/_a/assemblies?limit=3&&timestamp=2017-07-25 16:11:46 , it will return limit as 3
			limit, err := getLimitFromQuery(r, username)
			if err != nil {
				return nil, err
			}
			// getTimeStampFromQuery gets the value of timestamp parameter from ps httprouter.Params
			// e.g. localhost:8080/_a/assemblies?limit=3&&timestamp=2017-07-25 16:11:46 , it will return limit as 2017-07-25 16:11:46
			timestamp, err := getTimeStampFromQuery(r, username)
			if err != nil {
				return nil, err
			}

			response, serverError := h.Do(r, ps, username)
			if serverError != nil {
				return response, serverError
			}
			// We used reflect package from Go to make pagination Generic
			//reflect.ValueOf returns a new Value initialized to the concrete value
			// stored in the interface i.  ValueOf(nil) returns the zero Value.
			result := reflect.ValueOf(response)
			array := reflect.MakeSlice(reflect.TypeOf(response), 0, 10)
			count := limit
			if limit > result.Len() || limit == 0 {
				count = result.Len()
			}
			if timestamp == nil && limit == 0 {
				return response, nil

			} else if timestamp != nil {
				limitedResult := limitResultsByTimeStamp(attribute, count, timestamp, result, array)
				return limitedResult.Interface(), nil

			} else {
				limitedResult := limitResultsOnlyByCount(count, result, array)
				return limitedResult.Interface(), nil
			}
		})
	}
}

//getLimitFromQuery to get limit from params
func getLimitFromQuery(r *http.Request, username string) (int, *ServerError) {
	queryValues := r.URL.Query()
	limit := queryValues.Get("limit")
	if limit == "" {
		limit = "0"
	}
	limits, err := strconv.Atoi(limit)
	if err != nil {
		return 0, NewServerError("Parameter 'limit' has invalid value: "+limit+". Error: "+err.Error(),
			username, BadRequest, err)
	}
	return limits, nil

}

func getTimeStampFromQuery(r *http.Request, username string) (*time.Time, *ServerError) {
	queryValues := r.URL.Query()
	timestamp := queryValues.Get("timestamp")
	var t time.Time
	var err error
	if timestamp != "" {
		t, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, NewServerError("Parameter 'timestamp' has invalid value: "+timestamp+". Error: "+err.Error(),
				username, BadRequest, err)
		}
		return &t, nil
	} else {
		return nil, nil
	}

}

//limitResultsByTimeStamp limits the result by timestamp and  limit count
//attribute : FeildName taken into consideration for timeStamp
//count : value of limit
//timeStamp : timeStamp value
//input : Handler response
//array : Slice of type handler Response
func limitResultsByTimeStamp(attribute string, count int, timestamp *time.Time, input reflect.Value, array reflect.Value) reflect.Value {
	countArray := array
	for i := 0; i < input.Len(); i++ {
		var UpdatedAt time.Time
		// We extract timestamp from the input response struct by feild Name i.e attribute and typecast the interface to time
		UpdatedAt = input.Index(i).Elem().FieldByName(attribute).Interface().(time.Time)
		if UpdatedAt.Before(*timestamp) {
			array = reflect.Append(array, (input).Index(i))
		}
	}
	if array.Len() > count {
		for j := 0; j < count; j++ {
			countArray = reflect.Append(countArray, (array).Index(j))
		}
		return countArray

	} else {
		return array
	}
}

// limitResultsOnlyByCount limits results when timestamp params is not specified
// e.g. localhost:8080/_a/assemblies?limit=3
func limitResultsOnlyByCount(count int, input reflect.Value, array reflect.Value) reflect.Value {
	for i := 0; i < count; i++ {
		array = reflect.Append(array, (input).Index(i))
	}
	return array
}
