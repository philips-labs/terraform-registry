module terraform-registry

go 1.18

require (
	github.com/ProtonMail/go-crypto v1.1.6
	github.com/google/go-github/v32 v32.1.0
	github.com/labstack/echo/v4 v4.13.3
	golang.org/x/oauth2 v0.30.0
)

require (
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/time v0.8.0 // indirect
)

replace golang.org/x/crypto/openpgp v0.0.0-20210817164053-32db794688a5 => github.com/ProtonMail/go-crypto/openpgp v0.0.0-20220517143526-88bb52951d5b
