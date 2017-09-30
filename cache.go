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

// Cache returns a Decorator which will check the memcache for a cached Response. If a cached response is not found, it
// will invoke the underlying handler. The path of request is used the cache key.
// If the request contains a query param 'ignore_cache', then memcache will not be checked and underlying handler will be invoked.
func Cache(typ reflect.Type, ttl time.Duration) Decorator {
	return CacheWithKey(typ, ttl, func(r *http.Request, ps httprouter.Params, username string) (*string, *ServerError) {
		requestUri := r.RequestURI
		parsedUrl, err := url.Parse(requestUri)
		if err != nil {
			return nil, NewServerError(fmt.Sprintf("Failed to parse '%s': %v", r.RequestURI, err.Error()), username, MissingErrorCode, err)
		} else {
			key := parsedUrl.Path
			return &key, nil
		}
	})
}

// CacheForKey returns a Decorator which will check the memcache for a cached Response. If a cached response is not found, it
// will invoke the underlying handler. getKey function is used for generating a memcache key from a request.
// If the request contains a query param 'ignore_cache', then memcache will not be checked and underlying handler will be invoked.
func CacheWithKey(typ reflect.Type, ttl time.Duration,
	getKey func(r *http.Request, ps httprouter.Params, username string) (*string, *ServerError)) Decorator {
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
			item, err := get(ctx, *key)
			if err == memcache.ErrCacheMiss {
				response, serverError := h.Do(r, ps, username)
				if serverError != nil {
					return response, serverError
				}
				byteArray, err := toByteArray(response)
				if err != nil {
					return nil, NewServerError(fmt.Sprintf("Conversion to byte array failed: %v", err.Error()), username, MissingErrorCode, err)
				}
				set(ctx, *key, byteArray, ttl)
				return response, serverError
			} else if err != nil {
				return h.Do(r, ps, username)
			} else {
				response := reflect.New(typ)
				err := toInterface(item.Value, response.Interface())
				if err != nil {
					return nil, NewServerError(fmt.Sprintf("Conversion to interface failed: %v", err.Error()), username, MissingErrorCode, err)
				} else {
					if response.Elem().Kind() == reflect.Slice && response.Elem().IsNil() {
						// due to gob issue discussed here: https://github.com/golang/go/issues/10905
						return reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(response)), 0, 0).Interface(), nil
					} else {
						return response.Elem().Interface(), nil
					}
				}
			}
		})
	}
}

// set function puts an item in memcache
func set(ctx appengine.Context, key string, value []byte, expiry time.Duration) {
	item := &memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: expiry,
	}
	if err := memcache.Set(ctx, item); err != nil {
		ErrorMsg(ctx, fmt.Sprintf("Error setting key '%s': %v", key, err))
	}
}

// get function gets an item in memcache
func get(ctx appengine.Context, key string) (*memcache.Item, error) {
	item, err := memcache.Get(ctx, key)
	if err != nil && err != memcache.ErrCacheMiss {
		ErrorMsg(ctx, fmt.Sprintf("Error getting key '%s': %v", key, err))
	}
	return item, err
}

// toByteArray converts an interface to an byte array
func toByteArray(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	} else {
		return buf.Bytes(), nil
	}
}

// toInterface converts a byte array to a specified interface object
func toInterface(bts []byte, data interface{}) error {
	buf := bytes.NewBuffer(bts)
	dec := gob.NewDecoder(buf)
	return dec.Decode(data)
}
