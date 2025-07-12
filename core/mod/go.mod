module github.com/devlup-labs/Libr/core/mod

go 1.24.4

require (
	github.com/caarlos0/env/v10 v10.0.0
	github.com/devlup-labs/Libr/core/crypto v1.1.0
	github.com/gorilla/mux v1.8.1
	github.com/joho/godotenv v1.5.1
)

replace github.com/devlup-labs/Libr/core/crypto => ../crypto
