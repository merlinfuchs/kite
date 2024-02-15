module github.com/merlinfuchs/kite/examples/ping

go 1.21.5

require github.com/merlinfuchs/kite/kite-sdk-go v0.0.0

require (
	github.com/merlinfuchs/dismod v0.0.0-20240212142916-7150a62a3987
)

replace github.com/merlinfuchs/kite/kite-sdk-go v0.0.0 => ../../kite-sdk-go
