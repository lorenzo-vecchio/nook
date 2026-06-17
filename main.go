package main

import "github.com/lorenzo-vecchio/nook/cmd"

var version = "dev"

func main() {
	cmd.Version = version
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
