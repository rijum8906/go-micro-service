// Package coredto
package coredto

type RequestMeta struct {
	DeviceId string `validate:"required"`
}

type BrowserInfo struct {
	UserAgent string `validate:"required"`
	IPAddr    string `validate:"required"`
}
