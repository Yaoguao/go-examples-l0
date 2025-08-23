package main

import (
	"fmt"
	"wb-examples-l0/internal/config"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg)
}
