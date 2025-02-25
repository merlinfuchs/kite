module github.com/kitecloud/kite/kite-service

go 1.22.3

require (
	github.com/cyrusaf/ctxlog v1.3.2
	github.com/dgraph-io/ristretto v0.1.1
	github.com/diamondburned/arikawa/v3 v3.4.0
	github.com/eko/gocache/lib/v4 v4.2.0
	github.com/eko/gocache/store/ristretto/v4 v4.2.2
	github.com/endobit/clog v0.4.0
	github.com/expr-lang/expr v1.16.9
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/go-playground/validator/v10 v10.21.0
	github.com/golang-migrate/migrate/v4 v4.17.1
	github.com/jackc/pgx/v5 v5.6.0
	github.com/knadh/koanf/parsers/toml v0.1.0
	github.com/knadh/koanf/providers/env v0.1.0
	github.com/knadh/koanf/providers/file v0.1.0
	github.com/knadh/koanf/providers/rawbytes v0.1.0
	github.com/knadh/koanf/v2 v2.1.1
	github.com/lib/pq v1.10.9
	github.com/matoous/go-nanoid/v2 v2.1.0
	github.com/merlinfuchs/kite/kite-web v0.0.0
	github.com/minio/minio-go/v7 v7.0.76
	github.com/pelletier/go-toml/v2 v2.2.2
	github.com/ravener/discord-oauth2 v0.0.0-20230514095040-ae65713199b3
	github.com/rs/cors v1.11.0
	github.com/sashabaranov/go-openai v1.35.7
	github.com/sethvargo/go-limiter v1.0.0
	github.com/stretchr/testify v1.9.0
	github.com/urfave/cli/v2 v2.27.2
	github.com/valyala/fasttemplate v1.2.2
	golang.org/x/oauth2 v0.18.0
	gopkg.in/guregu/null.v4 v4.0.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/NdoleStudio/lemonsqueezy-go v1.2.4 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/schema v1.4.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/merlinfuchs/go-next-static v0.0.0-20240912153955-d431fbda6f18 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.52.3 // indirect
	github.com/prometheus/procfs v0.13.0 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/xrash/smetrics v0.0.0-20240312152122-5f08fbb34913 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/crypto v0.29.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sync v0.9.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	golang.org/x/time v0.8.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/merlinfuchs/kite/kite-web v0.0.0 => ../kite-web

replace github.com/diamondburned/arikawa/v3 v3.4.0 => github.com/merlinfuchs/arikawa/v3 v3.4.1-0.20241201143250-f3d1e0416b7d
