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

// Paginate returns a Decorator which will call the underlying handler and then return a section of the response based on
// limit and timestamp query params. limit is used to set the number of records to be returned. timestamp is used for ignoring
// records newer than the specified timestamp.
//
// Note: While this approach might be used in some cases, in most cases using the search support of app engine
// would be more efficient.
func Paginate(attribute string) Decorator {
	return func(h Handler) Handler {
		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
			limit, err := getLimit(r, username)
			if err != nil {
				return nil, err
			}
			timestamp, err := getTimestamp(r, username)
			if err != nil {
				return nil, err
			}

			response, serverError := h.Do(r, ps, username)
			if serverError != nil {
				return response, serverError
			}
			result := reflect.ValueOf(response)
			array := reflect.MakeSlice(reflect.TypeOf(response), 0, 10)
			count := limit
			if limit > result.Len() || limit == 0 {
				count = result.Len()
			}
			if timestamp == nil && limit == 0 {
				return response, nil
			} else if timestamp != nil {
				limitedResult := getResultsByLimitAndTimestamp(attribute, count, timestamp, result, array)
				return limitedResult.Interface(), nil
			} else {
				limitedResult := getResultsByLimit(count, result, array)
				return limitedResult.Interface(), nil
			}
		})
	}
}

// getLimit get limit from the query params. If no limit is found, 0 is returned.
func getLimit(r *http.Request, username string) (int, *ServerError) {
	queryValues := r.URL.Query()
	limit := queryValues.Get("limit")
	if limit == "" {
		limit = "0"
	}
	limits, err := strconv.Atoi(limit)
	if err != nil {
		return 0, NewServerError(fmt.Sprintf("Parameter 'limit' has invalid value: %s. Error: %v", limit, err.Error()),
			username, BadRequest, err)
	}
	return limits, nil
}

// getTimestamp get timestamp from the query params.
func getTimestamp(r *http.Request, username string) (*time.Time, *ServerError) {
	queryValues := r.URL.Query()
	timestamp := queryValues.Get("timestamp")
	var t time.Time
	var err error
	if timestamp != "" {
		t, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, NewServerError(fmt.Sprintf("Parameter 'timestamp' has invalid value: %s. Error: %v", timestamp, err.Error()),
				username, BadRequest, err)
		}
		return &t, nil
	} else {
		return nil, nil
	}
}

func getResultsByLimitAndTimestamp(attribute string, count int, timestamp *time.Time, input reflect.Value, array reflect.Value) reflect.Value {
	countArray := array
	for i := 0; i < input.Len(); i++ {
		var updatedAt time.Time
		updatedAt = input.Index(i).Elem().FieldByName(attribute).Interface().(time.Time)
		if updatedAt.Before(*timestamp) {
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

func getResultsByLimit(count int, input reflect.Value, array reflect.Value) reflect.Value {
	for i := 0; i < count; i++ {
		array = reflect.Append(array, (input).Index(i))
	}
	return array
}
