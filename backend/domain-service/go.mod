module github.com/hm-edu/domain-service

go 1.18

require (
	github.com/arsmn/fiber-swagger/v2 v2.24.0
	github.com/getkin/kin-openapi v0.92.0
	github.com/gofiber/contrib/fiberzap v0.0.0-20220322073455-9caa62c84b3d
	github.com/gofiber/contrib/otelfiber v0.0.0-20220322073455-9caa62c84b3d
	github.com/hm-edu/domain-api v0.0.0-00010101000000-000000000000
	github.com/hm-edu/portal-common v0.0.0-20220323161424-5f2cec9d122b
	go.opentelemetry.io/otel v1.5.0 // indirect
	go.uber.org/zap v1.21.0
	google.golang.org/grpc v1.45.0
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/joho/godotenv v1.4.0
	github.com/klauspost/compress v1.15.1 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/swaggo/files v0.0.0-20210815190702-a29dd2bc99b2 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.34.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	go.opentelemetry.io/contrib v1.5.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.30.0 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/gofiber/fiber/v2 v2.30.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.10.1
	github.com/swaggo/swag v1.8.0
	go.opentelemetry.io/otel/exporters/jaeger v1.5.0 // indirect
	go.opentelemetry.io/otel/sdk v1.5.0 // indirect
	go.opentelemetry.io/otel/trace v1.5.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220328115105-d36c6a25d886 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.10 // indirect
	google.golang.org/genproto v0.0.0-20220323144105-ec3c684e5b14 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

require (
	entgo.io/ent v0.10.1
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/gofiber/jwt/v3 v3.2.9
	github.com/mattn/go-sqlite3 v1.14.12
	github.com/spf13/cobra v1.4.0
)

require gopkg.in/go-playground/assert.v1 v1.2.1 // indirect

require (
	ariga.io/atlas v0.3.7-0.20220303204946-787354f533c3 // indirect
	github.com/agext/levenshtein v1.2.1 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/go-openapi/inflect v0.19.0 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.4.1 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/hcl/v2 v2.10.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mitchellh/go-wordwrap v0.0.0-20150314170334-ad45545899c7 // indirect
	github.com/zclconf/go-cty v1.8.0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
)

replace github.com/hm-edu/portal-common => ../common

replace github.com/hm-edu/domain-api => ../domain-api/go
