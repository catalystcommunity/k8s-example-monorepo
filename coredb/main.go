package coredb

import (
	"embed"
	"fmt"
)

//go:embed migrations/*.sql
var Migrations embed.FS

func main() {
	fmt.Println("hello world")
}
