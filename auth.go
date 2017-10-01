package core

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"model"
	"net/http"
	"strings"
)

// Claims struct has Username,UserType and jwt Token for authorization claims
type Claims struct {
	Username string `json:"username"`
	Type     string `json:"type"`
	jwt.StandardClaims
}

//PublicAuth is for non-user access in this case username is empty string,
func PublicAuth(privateKey string) Decorator {
	return func(h Handler) Handler {
		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
			unparsedToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if unparsedToken != "" {
				token, err := jwt.ParseWithClaims(unparsedToken,
					&Claims{},
					func(token *jwt.Token) (interface{}, error) {
						if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
							return nil, errors.New("Unexpected signing method")
						}
						return []byte(privateKey), nil
					})
				if err == nil {
					if claims, ok := token.Claims.(*Claims); ok && token.Valid {
						username = claims.Username
					}
				}
			}
			return h.Do(r, ps, username)
		})
	}
}

//ProtectedAuth is for registered users
func ProtectedAuth(privateKey string) Decorator {
	return func(h Handler) Handler {
		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
			unparsedToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if unparsedToken == "" {
				return nil, NewServerError("Unauthorized access", "", Unauthorized, nil)
			}
			token, err := jwt.ParseWithClaims(unparsedToken,
				&Claims{},
				func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, errors.New("Unexpected signing method")
					}
					return []byte(privateKey), nil
				})

			if err != nil {
				return nil, NewServerError("Unauthorized access", "", Unauthorized, nil)
			}
			if claims, ok := token.Claims.(*Claims); ok && token.Valid {
				return h.Do(r, ps, claims.Username)

			} else {
				return nil, NewServerError("Unauthorized access", claims.Username, Unauthorized, nil)
			}
		})
	}
}