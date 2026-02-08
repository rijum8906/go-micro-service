package jwt

func redisKey(jti string) string {
	return "scoped_action:" + jti
}

func sessionKey(sessionID string) string {
	return "auth:session:" + sessionID
}
