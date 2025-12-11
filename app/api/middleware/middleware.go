package middleware

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

// Chain is a middleware chain used for wrapping middlewares neatly
// example:
// stack := Chain(
//
//	Logging,
//	Auth,
//	CORS,
//
// )
// handler = stack(router)
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// BallAndChain is a middleware chain that ends with ball
// example:
// stack := BallAndChain(
//
//	Logging,
//	Auth,
//	CORS,
//
// )
// handler = stack(router)
// Logging will be called AFTER the ServeHTTP of router
func BallAndChain(ball Middleware, chain ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(chain) - 1; i >= 0; i-- {
			next = chain[i](next)
		}
		return ball(next)
	}
}
