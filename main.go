package main

import (
	"fmt"

	"github.com/max-programming/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}

	cfg.SetUser("usman")

	cfg, err = config.Read()
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg)
}
