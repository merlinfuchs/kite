module github.com/merlinfuchs/kite/kite-service

go 1.21.5

require (
	github.com/bwmarrin/discordgo v0.27.1
	github.com/cyrusaf/ctxlog v1.2.0
	github.com/endobit/clog v0.4.0
	github.com/go-playground/validator/v10 v10.16.0
	github.com/gofiber/fiber/v2 v2.52.0
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/jmoiron/sqlx v1.3.5
	github.com/jolestar/go-commons-pool/v2 v2.1.2
	github.com/knadh/koanf/parsers/toml v0.1.0
	github.com/knadh/koanf/providers/file v0.1.0
	github.com/knadh/koanf/providers/rawbytes v0.1.0
	github.com/knadh/koanf/v2 v2.0.1
	github.com/lib/pq v1.10.9
	github.com/merlinfuchs/kite/go-types v0.0.0-00010101000000-000000000000
	github.com/oklog/ulid/v2 v2.1.0
	github.com/pelletier/go-toml/v2 v2.1.1
	github.com/sqlc-dev/pqtype v0.3.0
	github.com/tetratelabs/wazero v1.5.0
	github.com/urfave/cli/v2 v2.27.0
	gopkg.in/guregu/null.v4 v4.0.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/net v0.18.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/merlinfuchs/kite/go-types => ../go-types
