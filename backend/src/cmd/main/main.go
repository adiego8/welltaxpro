package main

import (
	"context"
	"fmt"
	"io"
	"welltaxpro/src/cmd/server"

	"github.com/google/logger"
)

func main() {
	fmt.Printf("Starting WellTaxPro\n")

	// start Logger Settings
	logger.Init("WellTaxPro", true, false, io.Discard)
	ctx := context.Background()

	server.Run(ctx)
}
