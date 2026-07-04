package middleware

import (
	"context"
	"net"
	"reflect"
)

type contextKey string

const ConnKey contextKey = "ziti-connection"
const ClaimsKey contextKey = "token-claims"

// GetZitiIdentity extracts the source identity identifier from a net.Conn using reflection
func GetZitiIdentity(conn net.Conn) string {
	if conn == nil {
		return ""
	}

	// Use reflection to check if the connection has a SourceIdentifier method
	val := reflect.ValueOf(conn)

	// If it is a pointer, get the underlying value
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Look up method in value or pointer receiver
	method := val.MethodByName("SourceIdentifier")
	if !method.IsValid() {
		// Try addressable value pointer receiver
		if reflect.PointerTo(val.Type()).NumMethod() > 0 {
			ptrVal := reflect.New(val.Type())
			ptrVal.Elem().Set(val)
			method = ptrVal.MethodByName("SourceIdentifier")
		}
	}

	if method.IsValid() {
		results := method.Call(nil)
		if len(results) > 0 && results[0].Kind() == reflect.String {
			return results[0].String()
		}
	}

	return ""
}

// GetConnFromContext extracts net.Conn from request context
func GetConnFromContext(ctx context.Context) net.Conn {
	if conn, ok := ctx.Value(ConnKey).(net.Conn); ok {
		return conn
	}
	return nil
}
