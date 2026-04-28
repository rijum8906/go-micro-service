module github.com/rijum8906/relay/services/graphql-gateway

go 1.26.1

tool github.com/99designs/gqlgen

require (
	github.com/99designs/gqlgen v0.17.88
	github.com/rijum8906/relay/packages/core v0.0.0-20260324082703-6286eea9b4f5
	github.com/rijum8906/relay/packages/pb v0.0.0-00010101000000-000000000000
	github.com/rs/cors v1.11.1
	github.com/vektah/gqlparser/v2 v2.5.32
	go.uber.org/zap v1.27.1
	google.golang.org/grpc v1.79.3
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dgryski/trifles v0.0.0-20240922021506-5ecb8eeff266 // indirect
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.49.0 // indirect
)

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/caarlos0/env/v11 v11.4.0
	github.com/go-playground/validator/v10 v10.30.1
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/joho/godotenv v1.5.1
	github.com/redis/go-redis/v9 v9.18.0
	github.com/sosodev/duration v1.4.0 // indirect
	github.com/urfave/cli/v3 v3.7.0 // indirect
	golang.org/x/mod v0.33.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/tools v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.11
)

replace github.com/rijum8906/relay => ../..

replace github.com/rijum8906/relay/packages/core => ../../packages/core

replace github.com/rijum8906/relay/packages/pb => ../../packages/pb
