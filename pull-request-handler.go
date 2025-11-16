package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Структура PullRequest
type PR struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	AuthorID  int    `json:"author_id"`
	Status    string `json:"status"`
	Reviewers []int  `json:"reviewers,omitempty"`
}

func CreatePR(pr *PR) error {
	tx, err := Pool.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Сохраняем PR и получаем его id
	err = tx.QueryRow(
		"INSERT INTO PullRequests (title, author_id, status) VALUES ($1, $2, $3) RETURNING id",
		pr.Title, pr.AuthorID, pr.Status,
	).Scan(&pr.ID)
	if err != nil {
		return err
	}

	// Сначала читаем ревьюеров
	rows, err := tx.Query(`
        SELECT id FROM Users
        WHERE team = (SELECT team FROM Users WHERE id = $1)
        AND id != $1
        LIMIT 2
    `, pr.AuthorID)
	if err != nil {
		return err
	}

	reviewerIDs := []int{}

	for rows.Next() {
		var reviewerID int
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
	rows.Close() // <-- обязательно закрываем перед вставкой!

	// Теперь делаем вставки в reviewers
	for _, reviewerID := range reviewerIDs {
		_, err = tx.Exec(
			"INSERT INTO PullRequestReviewers (pr_id, reviewer_id) VALUES ($1, $2)",
			pr.ID, reviewerID,
		)
		if err != nil {
			return err
		}
		pr.Reviewers = append(pr.Reviewers, reviewerID)
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

	c.JSON(http.StatusOK, pr)
}
