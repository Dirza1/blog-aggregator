package main

import (
	"fmt"

	"github.com/Dirza1/blog-aggregator/src/internal/config"
)

func main() {
	config := config.Read()
	config.SetUser("Jasper")
	newConfig := config.Read()
	fmt.Println(newConfig.User)
}
