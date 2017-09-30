package core

import (
	"appengine"
	"appengine/memcache"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

//Set function is for setting item in memcache
func Set(ctx appengine.Context, key string, value []byte, expiry time.Duration) {
	// item is stored with response,key and expiration time
	item := &memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: expiry,
	}
	if err := memcache.Set(ctx, item); err != nil {
		ErrorMsg(ctx, fmt.Sprintf("Error setting key '%s': %v", key, err))
	}
}

//Get function is for retrieving item from memcache
//memcache.Get gets the item for the given key. ErrCacheMiss is returned for a memcache
//cache miss. The key must be at most 250 bytes in length.
func Get(ctx appengine.Context, key string) (*memcache.Item, error) {
	item, err := memcache.Get(ctx, key)
	if err != nil && err != memcache.ErrCacheMiss {
		ErrorMsg(ctx, fmt.Sprintf("Error getting key '%s': %v", key, err))
	}
	return item, err
}

//ToByteArray gives the byte array equivalent of data
func ToByteArray(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	// An Encoder manages the transmission of type and data information to the
	// other side of a connection.
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
	//Bytes returns a slice of length b.Len() holding the unread portion of the buffer.
	//The slice is valid for use only until the next buffer modification
}

//ToInterface converts byte response base to the data
func ToInterface(bts []byte, data interface{}) error {
	//NewBuffer creates and initializes a new Buffer using buf as its initial
	//contents.  It is intended to prepare a Buffer to read existing data
	buf := bytes.NewBuffer(bts)
	//NewDecoder returns a new decoder that reads from the io.Reader.
	dec := gob.NewDecoder(buf)
	return dec.Decode(data)
	//Decode reads the next value from the input stream and stores
	//it in the data represented by the empty interface value.
}

//Cache Decorator is used to get the Cacheed Resposne, it is our first Decorator in the Decorate list in the GET Request, we uses memcache for caching
func Cache(typ reflect.Type, ttl time.Duration) Decorator {
	return func(h Handler) Handler {
		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
			ctx := appengine.NewContext(r)
			//queryValues will have value for query params
			//In query params you can specify as ignore_cache=true , in this case it will ignore this decorator and
			//will return the original handler resposne as received
			queryValues := r.URL.Query()
			ignoreCache := queryValues.Get("ignore_cache")
			if ignoreCache == "true" {
				return h.Do(r, ps, username)
			}
			requestUri := r.RequestURI
			_url, err := url.Parse(requestUri)
			if err != nil {
				return nil, NewServerError("Failed to parse: " + err.Error(), username, MissingErrorCode, err)
			}
			// key is url path against which it will search the memcache for the response
			key := _url.Path
			//Get searches the response in memcache
			item, err := Get(ctx, key)
			if err == memcache.ErrCacheMiss {
				response, serverError := h.Do(r, ps, username)
				if serverError != nil {
					return response, serverError
				}
				byteArray, err := ToByteArray(response)
				if err != nil {
					return nil, NewServerError("Cache lookup failed: " + err.Error(), username, MissingErrorCode, err)
				}
				//Set caches the resposne against the key(_url.key) with the timeStamp to handle cache expiration
				Set(ctx, key, byteArray, ttl)
				return response, serverError
			} else if err != nil {
				return h.Do(r, ps, username)
			} else {
				// reflect.New returns a Value representing a pointer to a new zero value
				// for the specified type.  That is, the returned Value's Type is PtrTo(typ).
				response := reflect.New(typ)
				err := ToInterface(item.Value, response.Interface())
				if err != nil {
					return nil, NewServerError("Cache lookup failed v: " + err.Error(), username, MissingErrorCode, err)
				} else {
					if response.Elem().Kind() == reflect.Slice && response.Elem().IsNil() {
						// due to gob issue discussed here: https://github.com/golang/go/issues/10905
						return reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(response)), 0, 0).Interface(), nil
						// if resposne is of type slice we return this
					} else {
						// Elem returns the value that the interface v contains
						// returns value that interface contains
						return response.Elem().Interface(), nil
					}
				}
			}
		})
	}
}

// CacheForKey is the keyd version of Cache Decorator, it finds its use case ,in user specific responses
func CacheForKey(typ reflect.Type, ttl time.Duration,
getKey func(r *http.Request, ps httprouter.Params, username string) (string, *ServerError)) Decorator {
	return func(h Handler) Handler {
		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
			ctx := appengine.NewContext(r)
			queryValues := r.URL.Query()
			ignoreCache := queryValues.Get("ignore_cache")
			if ignoreCache == "true" {
				return h.Do(r, ps, username)
			}
			//getKey will return key for user specific key for cache
			key, serverError := getKey(r, ps, username)
			if serverError != nil {
				return nil, serverError
			}
			item, err := Get(ctx, key)
			if err == memcache.ErrCacheMiss {
				response, serverError := h.Do(r, ps, username)
				if serverError != nil {
					return response, serverError
				}
				// ToByteArray converts response to Byte Array
				byteArray, err := ToByteArray(response)
				if err != nil {
					return nil, NewServerError("Cache lookup failed: " + err.Error(), username, MissingErrorCode, err)
				}
				//Set caches the resposne against the key(_url.key) with the timeStamp to handle cache expiration
				Set(ctx, key, byteArray, ttl)
				return response, serverError
			} else if err != nil {
				return h.Do(r, ps, username)
			} else {
				// reflect.New returns a Value representing a pointer to a new zero value
				// for the specified type.  That is, the returned Value's Type is PtrTo(typ).
				response := reflect.New(typ)
				err := ToInterface(item.Value, response.Interface())
				if err != nil {
					return nil, NewServerError("Cache lookup failed v: " + err.Error(), username, MissingErrorCode, err)
				} else {
					// Elem returns the value that the interface v contains
					// returns value that interface contains
					return response.Elem().Interface(), nil
				}
			}
		})
	}
}
