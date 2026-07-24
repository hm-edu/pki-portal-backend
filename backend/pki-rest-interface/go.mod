module github.com/hm-edu/pki-rest-interface

go 1.26.0

require (
	github.com/MicahParks/keyfunc/v3 v3.8.0
	github.com/getsentry/sentry-go/echo v0.48.0
	github.com/hm-edu/portal-apis v0.0.0-20260722062737-d43882e11746
	github.com/hm-edu/portal-common v0.0.0-20260613132347-a1589de7a36f
	github.com/labstack/echo/v5 v5.3.1
	github.com/labstack/gommon v0.5.0
	github.com/spf13/cobra v1.10.2
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.69.0
	go.uber.org/zap v1.28.0
	google.golang.org/grpc v1.82.1
)

require (
	github.com/MicahParks/jwkset v0.11.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.15 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.3 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/prometheus/client_golang v1.24.1 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260723215102-3fe39f3c1018 // indirect
)

require (
	github.com/TheZeroSlave/zapsentry v1.24.0
	github.com/getsentry/sentry-go v0.48.0
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/spf13/pflag v1.0.10 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/hm-edu/portal-common => ../common

replace github.com/johnbellone/grpc-middleware-sentry => github.com/janwytze/grpc-middleware-sentry v0.4.1-0.20260505112459-44a6e7845810
