package main

import (
	"ismelen/ermc/internal/cli"
	"log"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
