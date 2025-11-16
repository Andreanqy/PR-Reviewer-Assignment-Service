package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

var user, password, host, database string = "postgres", "password", "localhost", "mydb"
var port int = 5432

func main() {
	ConnectDB(fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, database))
	defer Close()

	r := gin.Default()
	r.POST("/pullRequest/create", CreatePRHandler)
	r.Run(":8080")
}
