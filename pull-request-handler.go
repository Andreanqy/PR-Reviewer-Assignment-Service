package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Добавление PullRequest в БД
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

// Реализация операции merge
func MergePRHandler(c *gin.Context) {
	var body struct {
		PullRequestID string `json:"pull_request_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prID, _ := strconv.Atoi(body.PullRequestID)

	// Проверяем статус
	var status string
	err := Pool.QueryRow("SELECT status FROM PullRequests WHERE id=$1", prID).Scan(&status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
		return
	}

	if status == "MERGED" {
		//c.JSON(http.StatusOK, gin.H{"message": "PR already merged"})
		c.JSON(http.StatusOK, gin.H{
			"pull_request_id": body.PullRequestID,
			"status":          "MERGED",
		})
		return
	}

	_, err = Pool.Exec("UPDATE PullRequests SET status='MERGED' WHERE id=$1", prID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем PR с ревьюверами
	rows, _ := Pool.Query("SELECT reviewer_id FROM PullRequestReviewers WHERE pr_id=$1", prID)
	for rows.Next() {
		var id int
		rows.Scan(&id)
	}

	c.JSON(http.StatusOK, gin.H{
		"pull_request_id": body.PullRequestID,
		"status":          "MERGED",
	})
}

// Переназначение ревьювера
func ReassignReviewerPRHandler(c *gin.Context) {
	var body struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prID, _ := strconv.Atoi(body.PullRequestID)
	oldID, _ := strconv.Atoi(body.OldUserID)

	// Проверяем, что PR существует и открыт
	var status string
	var authorID int
	err := Pool.QueryRow("SELECT status, author_id FROM PullRequests WHERE id=$1", prID).Scan(&status, &authorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
		return
	}
	if status == "MERGED" {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot reassign on merged PR"})
		return
	}

	// Проверяем, что старый ревьювер назначен
	var exists bool
	err = Pool.QueryRow("SELECT EXISTS(SELECT 1 FROM PullRequestReviewers WHERE pr_id=$1 AND reviewer_id=$2)", prID, oldID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusConflict, gin.H{"error": "reviewer is not assigned to this PR"})
		return
	}

	// Получаем команду старого ревьювера
	var team string
	err = Pool.QueryRow("SELECT team FROM Users WHERE id=$1", oldID).Scan(&team)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "old user not found"})
		return
	}

	// Находим первого доступного кандидата
	var newID int
	err = Pool.QueryRow(`
        SELECT id FROM Users 
        WHERE team=$1 AND is_active=true AND id!=$2 AND id!=$3 
        AND id NOT IN (SELECT reviewer_id FROM PullRequestReviewers WHERE pr_id=$4)
        LIMIT 1
    `, team, oldID, authorID, prID).Scan(&newID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "no available replacement candidate"})
		return
	}

	// Обновляем ревьювера
	_, err = Pool.Exec("UPDATE PullRequestReviewers SET reviewer_id=$1 WHERE pr_id=$2 AND reviewer_id=$3", newID, prID, oldID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем новый список ревьюверов
	rows, _ := Pool.Query("SELECT reviewer_id FROM PullRequestReviewers WHERE pr_id=$1", prID)
	var reviewers []string
	for rows.Next() {
		var id int
		rows.Scan(&id)
		reviewers = append(reviewers, strconv.Itoa(id))
	}
	rows.Close()

	c.JSON(http.StatusOK, gin.H{
		"pr": PR{
			PullRequestID:     body.PullRequestID,
			Status:            status,
			AssignedReviewers: reviewers,
		},
		"replaced_by": strconv.Itoa(newID),
	})
}
