package main

import (
	"fmt"
	"sso/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Print(cfg)
}
