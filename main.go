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
	r.POST("/pullRequest/merge", MergePRHandler)
	r.POST("/pullRequest/reassign", ReassignReviewerPRHandler)

	//r.POST("team/add", CreateTeamHandler)
	//r.POST("users/setIsActive", SetUserActiveHandler)

	r.Run(":8080")
}
