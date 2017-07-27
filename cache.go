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

// Set function is for setting item in memcache

func Set(ctx appengine.Context, key string, value []byte, expiry time.Duration) {
	defer func(t time.Time) {
		DebugMsg(ctx, fmt.Sprintf("--- Time Elapsed for setting memcache key '%s': %v ---\n", key, time.Since(t)))
	}(time.Now())

	// Shows time elapsed in the Set function

	item := &memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: expiry,
	}

	// item is stores with respsone,key and expiration time

	if err := memcache.Set(ctx, item); err != nil {
		ErrorMsg(ctx, fmt.Sprintf("Error setting key '%s': %v", key, err))

		// throws error if there is an error seting key in memcache
	}
}

// Get function is for retriving item from memcache

func Get(ctx appengine.Context, key string) (*memcache.Item, error) {
	defer func(t time.Time) {
		DebugMsg(ctx, fmt.Sprintf("--- Time Elapsed for getting memcache key '%s': %v ---\n", key, time.Since(t)))
	}(time.Now())

	// Shows time elapsed in the Get function

	item, err := memcache.Get(ctx, key)

	// memcache.Get gets the item for the given key. ErrCacheMiss is returned for a memcache
	// cache miss. The key must be at most 250 bytes in length.

	if err != nil && err != memcache.ErrCacheMiss {
		ErrorMsg(ctx, fmt.Sprintf("Error getting key '%s': %v", key, err))
	}
	return item, err
}

func ToByteArray(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	// An Encoder manages the transmission of type and data information to the
	// other side of a connection.

	err := enc.Encode(data)

	// Encode transmits the data item represented by the empty interface value,
	// guaranteeing that all necessary type information has been transmitted first.

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil

	// Bytes returns a slice of length b.Len() holding the unread portion of the buffer.
	// The slice is valid for use only until the next buffer modification
}

// ToInterface converts byte respsonce base to the data

func ToInterface(bts []byte, data interface{}) error {
	buf := bytes.NewBuffer(bts)

	// NewBuffer creates and initializes a new Buffer using buf as its initial
	// contents.  It is intended to prepare a Buffer to read existing data

	dec := gob.NewDecoder(buf)

	// NewDecoder returns a new decoder that reads from the io.Reader.

	return dec.Decode(data)

	// Decode reads the next value from the input stream and stores
	// it in the data represented by the empty interface value.
}

// Cache Decorator is used to get the Cacheed Resposne, it is our first Decorator in the Decorate list in the GET Request, we uses memcache for caching

func Cache(typ reflect.Type, ttl time.Duration) Decorator {
	return func(h Handler) Handler {
		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
			ctx := appengine.NewContext(r)

			// appengine context derived from request

			defer func(t time.Time) {
				DebugMsg(ctx, fmt.Sprintf("--- Time Elapsed for cache decorator : %v ---\n", time.Since(t)))
			}(time.Now())

			// Shows time elapsed in the Cache Decorator

			queryValues := r.URL.Query()
			ignoreCache := queryValues.Get("ignore_cache")
			if ignoreCache == "true" {
				return h.Do(r, ps, username)
			}

			// queryValues will have value for query params
			// In query params you can specify as ignore_cache=true , in this case it will ignore this decorator and
			// will return the original handler resposne as received

			requestUri := r.RequestURI
			_url, err := url.Parse(requestUri)

			//  Parse parses rawurl into a URL structure.
			// The rawurl may be relative or absolute. This is useful in ingoring query params while caching

			if err != nil {
				return nil, NewServerError("Failed to parse: "+err.Error(), username, MissingErrorCode, err)
			}

			// return error if it fails to parse url

			key := _url.Path

			// key is url path against which it will search the memcache for the response

			item, err := Get(ctx, key)

			// Get searchs the respsone in memcache

			if err == memcache.ErrCacheMiss {
				response, serverError := h.Do(r, ps, username)

				// it the error is ErrCacheMiss means that an operation failed
				// because the item wasn't present. we returb the hander func to get the respsone

				if serverError != nil {
					return response, serverError
				}

				// set the value in cache
				byteArray, err := ToByteArray(response)
				if err != nil {
					return nil, NewServerError("Cache lookup failed: "+err.Error(), username, MissingErrorCode, err)
				}

				// ToByteArray converts response to Byte Array

				Set(ctx, key, byteArray, ttl)

				// Once we get the ByteArray we cache the resposne using Set against the key(_url.key) with the timeStamp to handle cache expiration

				return response, serverError
			} else if err != nil {
				return h.Do(r, ps, username)

				// If err is other than ErrCacheMiss we invoke handler without seting the respsonse in cache

			} else {
				response := reflect.New(typ)

				// reflect.New returns a Value representing a pointer to a new zero value
				// for the specified type.  That is, the returned Value's Type is PtrTo(typ).

				err := ToInterface(item.Value, response.Interface())
				if err != nil {
					return nil, NewServerError("Cache lookup failed v: "+err.Error(), username, MissingErrorCode, err)
				} else {
					if response.Elem().Kind() == reflect.Slice && response.Elem().IsNil() {
						// due to gob issue discussed here: https://github.com/golang/go/issues/10905
						return reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(response)), 0, 0).Interface(), nil

						// if resposne is of type slice we return this

					} else {
						return response.Elem().Interface(), nil
						// Elem returns the value that the interface v contains
						// returns value that interface contains
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

			// appengine context derived from request

			defer func(t time.Time) {
				DebugMsg(ctx, fmt.Sprintf("--- Time Elapsed for cache decorator : %v ---\n", time.Since(t)))
			}(time.Now())

			// Shows time elapsed in the CacheForKey Decorator

			queryValues := r.URL.Query()
			ignoreCache := queryValues.Get("ignore_cache")
			if ignoreCache == "true" {
				return h.Do(r, ps, username)
			}

			// queryValues will have value for query params
			// In query params you can specify as ignore_cache=true , in this case it will ignore this decorator and
			// will return the original handler resposne as received

			key, serverError := getKey(r, ps, username)
			if serverError != nil {
				return nil, serverError
			}

			// getKey will return key for user specific key for cachce

			item, err := Get(ctx, key)
			if err == memcache.ErrCacheMiss {
				response, serverError := h.Do(r, ps, username)
				if serverError != nil {
					return response, serverError
				}

				// it the error is ErrCacheMiss means that an operation failed
				// because the item wasn't present. we returb the hander func to get the respsone

				// set the value in cache
				byteArray, err := ToByteArray(response)
				if err != nil {
					return nil, NewServerError("Cache lookup failed: "+err.Error(), username, MissingErrorCode, err)
				}

				// ToByteArray converts response to Byte Array

				Set(ctx, key, byteArray, ttl)

				// Once we get the ByteArray we cache the resposne using Set against the key(_url.key) with the timeStamp to handle cache expiration

				return response, serverError
			} else if err != nil {
				return h.Do(r, ps, username)

				// If err is other than ErrCacheMiss we invoke handler without seting the respsonse in cache

			} else {
				response := reflect.New(typ)

				// reflect.New returns a Value representing a pointer to a new zero value
				// for the specified type.  That is, the returned Value's Type is PtrTo(typ).

				err := ToInterface(item.Value, response.Interface())
				if err != nil {
					return nil, NewServerError("Cache lookup failed v: "+err.Error(), username, MissingErrorCode, err)
				} else {
					return response.Elem().Interface(), nil

					// Elem returns the value that the interface v contains
					// returns value that interface contains

				}
			}
		})
	}
}
