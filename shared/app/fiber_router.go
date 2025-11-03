package app

import (
	"github.com/gofiber/fiber/v2"
)

// Get registers GET route with Circuit Breaker + Retry
func RouterGet(router fiber.Router, path string, handlers ...fiber.Handler) {
	router.Get(path, WrapHandlers(path, handlers)...)
}

// Post registers POST route with Circuit Breaker + Retry
func RouterPost(router fiber.Router, path string, handlers ...fiber.Handler) {
	router.Post(path, WrapHandlers(path, handlers)...)
}

// Post registers POST route with Circuit Breaker + Retry
func RouterPut(router fiber.Router, path string, handlers ...fiber.Handler) {
	RouterRequest("PUT", router, path, WrapHandlers(path, handlers)...)
}

func RouterDelete(router fiber.Router, path string, handlers ...fiber.Handler) {
	RouterRequest("DELETE", router, path, WrapHandlers(path, handlers)...)
}

// Request supports any method (PUT, DELETE, etc) with Circuit Breaker + Retry
func RouterRequest(method string, router fiber.Router, path string, handlers ...fiber.Handler) {
	router.Add(method, path, WrapHandlers(path, handlers)...)
}

// GetWithOptions allows passing custom retry options
func RouterGetWithOptions(router fiber.Router, path string, opts RetryOptions, handlers ...fiber.Handler) {
	router.Get(path, WrapHandlers(path, handlers, opts)...)
}

// PostWithOptions allows passing custom retry options
func RouterPostWithOptions(router fiber.Router, path string, opts RetryOptions, handlers ...fiber.Handler) {
	router.Post(path, WrapHandlers(path, handlers, opts)...)
}

// RequestWithOptions allows passing custom retry options
func RouterRequestWithOptions(method string, router fiber.Router, path string, opts RetryOptions, handlers ...fiber.Handler) {
	router.Add(method, path, WrapHandlers(path, handlers, opts)...)
}
