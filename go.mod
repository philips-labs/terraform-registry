module terraform-registry

go 1.14

require (
	github.com/ProtonMail/go-crypto v0.0.0-20220517143526-88bb52951d5b
	github.com/google/go-github/v32 v32.1.0
	github.com/labstack/echo/v4 v4.9.1
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace golang.org/x/crypto/openpgp v0.0.0-20210817164053-32db794688a5 => github.com/ProtonMail/go-crypto/openpgp v0.0.0-20220517143526-88bb52951d5b
