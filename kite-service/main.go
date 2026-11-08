package main

import (
	// The service image has no zoneinfo, and formatDate in the Run JavaScript
	// block takes time zones by name.
	_ "time/tzdata"

	"github.com/kitecloud/kite/kite-service/cmd"
)

func main() {
	cmd.Execute()
}
