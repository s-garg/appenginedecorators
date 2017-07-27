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

// Auth Decorator checks for authorization on the response ,it is places as the last decorator once the handler is ready with all the decorations or ehancements we apply
// Auth Decorator to validate the access
//we are using OAuth 2 token, for Auth Check
// PublicAuth is for non-user access in this case username is empty string,

func PublicAuth(privateKey string) Decorator {
	return func(h Handler) Handler {
		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
			unparsedToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

			// unparsedToken, it sets the header for Authorization

			if unparsedToken != "" {

				// ParseWithClaims generates token for Authentication

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

			// returns resposne with updated username
		})
	}
}

// ProtectedAuth is for registed users

func ProtectedAuth(privateKey string) Decorator {
	return func(h Handler) Handler {
		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {
			unparsedToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

			// unparsedToken, it sets the header for Authorization
			if unparsedToken == "" {
				return nil, NewServerError("Unauthorized access", "", Unauthorized, nil)
			}

			// If token is empty we return server error

			token, err := jwt.ParseWithClaims(unparsedToken,
				&Claims{},
				func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, errors.New("Unexpected signing method")
					}
					return []byte(privateKey), nil
				})

			// ParseWithClaims genrates a token for login i.e a Private key for user

			if err != nil {
				return nil, NewServerError("Unauthorized access", "", Unauthorized, nil)
			}

			if claims, ok := token.Claims.(*Claims); ok && token.Valid {
				return h.Do(r, ps, claims.Username)

				// if the token is valid and verified then we return response

			} else {
				return nil, NewServerError("Unauthorized access", claims.Username, Unauthorized, nil)

				// if there is err in claims validation we return ServerError
			}
		})
	}
}

// AdminAuth is for admins

func AdminAuth(privateKey string) Decorator {
	return func(h Handler) Handler {
		return HandlerFunc(func(r *http.Request, ps httprouter.Params, username string) (interface{}, *ServerError) {

			// unparsedToken, it sets the header for Authorization

			unparsedToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if unparsedToken == "" {
				return nil, NewServerError("Unauthorized access", "", Unauthorized, nil)
			}

			// If token is empty we return server error

			token, err := jwt.ParseWithClaims(unparsedToken,
				&Claims{},
				func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, errors.New("Unexpected signing method")
					}
					return []byte(privateKey), nil
				})

			// ParseWithClaims genrates a token for login i.e a Private key for user

			if err != nil {
				return nil, NewServerError("Unauthorized access", "", Unauthorized, nil)
			}

			if claims, ok := token.Claims.(*Claims); ok && token.Valid {
				if claims.Type == model.UserType_Admin.String() {
					return h.Do(r, ps, claims.Username)

					// if the token is valid and verified then we return response and we do an additional check for user type as Admin

				} else {
					return nil, NewServerError(fmt.Sprintf("User %s is not admin", claims.Username),
						claims.Username, Forbidden, nil)

					// if user is not admin we return serverError
				}
			} else {
				return nil, NewServerError("Unauthorized access", "", Unauthorized, nil)

				// if there is err in claims validation we return ServerError

			}
		})
	}
}
