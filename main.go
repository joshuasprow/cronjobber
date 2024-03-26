package main

import (
	"os"

	"github.com/joshuasprow/cronjobber/cmd"
	"github.com/joshuasprow/cronjobber/pkg"
)

func main() {
	log := pkg.NewLogger()

	if err := cmd.Run(log); err != nil {
		log.Error("run", "error", err)
		os.Exit(1)
	}
}
