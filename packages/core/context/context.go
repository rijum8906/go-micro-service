// Package context
package context

type ContextKey string

const (
	// Client info keys
	KeyDeviceID   ContextKey = "device_id"
	KeyUserAgent  ContextKey = "user_agent"
	KeyClientIP   ContextKey = "client_ip"
	KeyClientType ContextKey = "client_type"

	// Request tracking
	KeyRequestID ContextKey = "request_id"
	KeyTraceID   ContextKey = "trace_id"

	// Auth context
	KeyUserID ContextKey = "user_id"
)
