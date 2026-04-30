package dto

type UserInfo struct {
	UserID      string
	AccessToken string
	SessionID   string
	Role        string
}

type ClientInfo struct {
	DeviceID   string
	UserAgent  string
	IPAddress  string
	ClientType string
	APIVersion string
	SDKVersion string
	RequestID  string
	SessionID  string
	TraceID    string
	Locale     string
}
