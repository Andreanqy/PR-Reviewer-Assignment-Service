package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Структура PullRequest
type PR struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers,omitempty"`
}

func CreatePR(pr *PR) error {
	tx, err := Pool.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Сохранение PR и получение его id
	var prID int
	err = tx.QueryRow(
		"INSERT INTO PullRequests (pull_request_name, author_id, status) VALUES ($1, $2, $3) RETURNING id",
		pr.PullRequestName, pr.AuthorID, pr.Status,
	).Scan(&prID)
	if err != nil {
		return err
	}
	pr.PullRequestID = strconv.Itoa(prID)

	// Чтение доступных ревьюверов
	rows, err := tx.Query(`
        SELECT id FROM Users
        WHERE team = (SELECT team FROM Users WHERE id = $1)
        AND id != $1
        LIMIT 2
    `, pr.AuthorID)
	if err != nil {
		return err
	}

	reviewerIDs := []string{}

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			rows.Close()
			return err
		}
		reviewerIDs = append(reviewerIDs, reviewerID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	// Вставки в assigned_reviewers
	for _, reviewerID := range reviewerIDs {
		_, err = tx.Exec(
			"INSERT INTO PullRequestReviewers (pr_id, reviewer_id) VALUES ($1, $2)",
			pr.PullRequestID, reviewerID,
		)
		if err != nil {
			return err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return tx.Commit()
}

// Создание PullRequest
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

	c.JSON(http.StatusCreated, pr)
}
