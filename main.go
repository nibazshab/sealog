package main

import (
	"flag"
	"os"
)

func args() {
	repw := flag.Bool("reset-password", false, "Reset admin password")

	flag.Parse()

	if *repw {
		resetAdminPassword()
		os.Exit(0)
	}
}
