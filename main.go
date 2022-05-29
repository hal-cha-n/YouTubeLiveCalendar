package main

import (
	"os"
)

func main() {
	aaa_key := get_developer_key()
}

func get_developer_key() string {
	aaa_key := os.Getenv("AAA")
	return aaa_key
}
