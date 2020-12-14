package middleware

import (
	"net/http"
)

// CheckHeader ...
func CheckHeader(header, value string) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if value != request.Header.Get(header) {
				http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			handler.ServeHTTP(writer, request)
		})
	}
}
