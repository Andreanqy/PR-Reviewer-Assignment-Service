package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

var user, password, database string = "postgres", "password", "mydb"

func main() {
	ConnectDB(fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable", user, password, database))
	defer Close()

	r := gin.Default()
	r.POST("/pr", CreatePRHandler)
	r.Run(":8080")
}
