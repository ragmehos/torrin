package main

import "github.com/torrin-app/torrin/shared/service"

func main() {
	service.RunHealth("scheduler", "8081")
}
