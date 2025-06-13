package main

import (
	"fmt"

	"github.com/Dirza1/blog-aggregator/internal/config"
)

func main() {

	test := config.Read()
	test.SetUser("Jasper")

	output := config.Read()
	fmt.Println(output.User)
	fmt.Println(output.Url)

}
