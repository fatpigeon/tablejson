package main

import (
	"fmt"
)

func main() {
	conf := GetConfig()
	fmt.Println(Portal(conf.Url, conf.File, conf.Mode))
}
