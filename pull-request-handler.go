package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Структура PullRequest
type PR struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	AuthorID int    `json:"author_id"`
	Status   string `json:"status"`
}

// Создание PullRequest в БД
func CreatePR(pr *PR) error {
	err := Pool.QueryRow(
		"INSERT INTO PullRequests (title, author_id, status) VALUES ($1, $2, $3) RETURNING id",
		pr.Title, pr.AuthorID, pr.Status,
	).Scan(&pr.ID)
	return err
}

func CreatePRHandler(c *gin.Context) {
	var pr PR
	if err := c.ShouldBindJSON(&pr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pr.Status = "OPEN"

	if err := CreatePR(&pr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pr.ID)
}
