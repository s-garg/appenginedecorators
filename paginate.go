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

			// appengine contxt derived from request

			defer func(t time.Time) {
				DebugMsg(ctx, fmt.Sprintf("--- Time Elapsed for paginate decorator : %v ---\n", time.Since(t)))
			}(time.Now())

			// Shows time elpased in the Pagination Decorator

			limit, err := getLimitFromQuery(r, username)
			if err != nil {
				return nil, err
			}

			// getLimitFromQuery gets the value of limit parameter from ps httprouter.Params
			// e.g. localhost:8080/_a/assemblies?limit=3&&timestamp=2017-07-25 16:11:46 , it will return limit as 3

			timestamp, err := getTimeStampFromQuery(r, username)
			if err != nil {
				return nil, err
			}

			// getTimeStampFromQuery gets the value of timestamp parameter from ps httprouter.Params
			// e.g. localhost:8080/_a/assemblies?limit=3&&timestamp=2017-07-25 16:11:46 , it will return limit as 2017-07-25 16:11:46

			response, serverError := h.Do(r, ps, username)

			// here we get resposne of the handler on which decorator will be applied

			if serverError != nil {
				return response, serverError
			}

			// It will return serverError if there is some error in handler's response

			result := reflect.ValueOf(response)                         // got the input
			array := reflect.MakeSlice(reflect.TypeOf(response), 0, 10) // created a slice value

			// We used reflect package from Go to make pagination Generic
			//reflect.ValueOf returns a new Value initialized to the concrete value
			// stored in the interface i.  ValueOf(nil) returns the zero Value.

			count := limit
			if limit > result.Len() || limit == 0 {
				count = result.Len()
			}

			// if limit params is 0 or more than number of results  in response we set in to the length of resposne

			if timestamp == nil && limit == 0 {
				return response, nil

				// if params are not specified e.g.  localhost:8080/_a/assemblies, Paginate decorator will just return response from the handler with no change

			} else if timestamp != nil {
				limitedResult := limitResultsByTimeStamp(attribute, count, timestamp, result, array)

				// limitResultsByTimeStamp limits the result by TimeStamp and count

				return limitedResult.Interface(), nil

			} else {
				limitedResult := limitResultsOnlyByCount(count, result, array)

				// limitResultsByTimeStamp limits the result by count

				return limitedResult.Interface(), nil
			}
		})
	}
}

// getLimitFromQuery to get limit from params

func getLimitFromQuery(r *http.Request, username string) (int, *ServerError) {
	queryValues := r.URL.Query()

	// queryValues will have value for query params

	limit := queryValues.Get("limit")

	//   queryValues.Get gets the value of parameter in this case parameter is limit

	if limit == "" {
		limit = "0"
	}

	//  If limit is "" we default it to 0

	limits, err := strconv.Atoi(limit)

	//strconv.Atoi converts to integer datatype

	if err != nil {
		return 0, NewServerError("Parameter 'limit' has invalid value: "+limit+". Error: "+err.Error(),
			username, BadRequest, err)
	}

	// error handling

	return limits, nil

}

func getTimeStampFromQuery(r *http.Request, username string) (*time.Time, *ServerError) {
	queryValues := r.URL.Query()

	// queryValues will have value for query params

	timestamp := queryValues.Get("timestamp")

	//   queryValues.Get gets the value of parameter in this case parameter is timestamp

	var t time.Time
	var err error
	if timestamp != "" {
		t, err = time.Parse(time.RFC3339, timestamp)

		// time.Parse will convert the time stamp to unix.time()

		if err != nil {
			return nil, NewServerError("Parameter 'timestamp' has invalid value: "+timestamp+". Error: "+err.Error(),
				username, BadRequest, err)
		}
		// error handling

		return &t, nil

		// returns timeStamp

	} else {
		return nil, nil
	}

}

// limitResultsByTimeStamp limits the result by timestamp and  limit count

func limitResultsByTimeStamp(attribute string, count int, timestamp *time.Time, input reflect.Value, array reflect.Value) reflect.Value {

	// attribute : FeildName taken into consideration for timeStamp
	// count : value of limit
	// timeStamp : timeStamp value
	// input : Handler response
	// array : Slice of type handler Response

	countArray := array
	for i := 0; i < input.Len(); i++ {
		var UpdatedAt time.Time
		UpdatedAt = input.Index(i).Elem().FieldByName(attribute).Interface().(time.Time)

		// We exract timestamp from the input response struct by feild Name i.e attribute and typecast the inteerface to time

		if UpdatedAt.Before(*timestamp) {
			array = reflect.Append(array, (input).Index(i))
		}

		// array collects all the response whose timestamp is before the UpdatedAt TimeStamp
	}
	if array.Len() > count {
		for j := 0; j < count; j++ {
			countArray = reflect.Append(countArray, (array).Index(j))
		}
		// Here we limit results whcih we got from timestamp filtering to the limit we got in params

		return countArray

	} else {
		return array
	}

	// we return responsne array with Pagination decorator applied
}

// limitResultsOnlyByCount limits results when timestamp params is not specified
// e.g. localhost:8080/_a/assemblies?limit=3

func limitResultsOnlyByCount(count int, input reflect.Value, array reflect.Value) reflect.Value {
	for i := 0; i < count; i++ {
		array = reflect.Append(array, (input).Index(i))
	}
	// we loop thoruhg the array to the specified limit

	return array

	// we return responsne array with Pagination decorator applied
}
