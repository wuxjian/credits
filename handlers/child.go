package handlers

import (
	"fmt"
	"net/http"
	"time"

	"credits/database"
	"credits/models"

	"github.com/gin-gonic/gin"
)

// GetChildTasks returns all non-deleted tasks with completion status for the child
func GetChildTasks(c *gin.Context) {
	today := time.Now().Format("2006-01-02")

	// Get all one-time tasks with completion status
	onceRows, err := database.DB.Query(`
		SELECT t.id, t.name, t.points, t.task_type,
			CASE WHEN cp.task_id IS NOT NULL THEN 1 ELSE 0 END as completed
		FROM tasks t
		LEFT JOIN child_progress cp ON t.id = cp.task_id
		WHERE t.task_type = 'once' AND t.deleted = 0
		ORDER BY t.id DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	defer onceRows.Close()

	var tasks []models.ChildTask
	for onceRows.Next() {
		var t models.ChildTask
		if err := onceRows.Scan(&t.ID, &t.Name, &t.Points, &t.TaskType, &t.Completed); err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
			return
		}
		tasks = append(tasks, t)
	}

	// Get all daily tasks with today's completion status
	dailyRows, err := database.DB.Query(`
		SELECT t.id, t.name, t.points, t.task_type,
			CASE WHEN dts.task_id IS NOT NULL THEN 1 ELSE 0 END as completed
		FROM tasks t
		LEFT JOIN daily_task_status dts ON t.id = dts.task_id AND dts.date = ? AND dts.completed = 1
		WHERE t.task_type = 'daily' AND t.deleted = 0
		ORDER BY t.id DESC
	`, today)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	defer dailyRows.Close()

	for dailyRows.Next() {
		var t models.ChildTask
		if err := dailyRows.Scan(&t.ID, &t.Name, &t.Points, &t.TaskType, &t.Completed); err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
			return
		}
		tasks = append(tasks, t)
	}

	if tasks == nil {
		tasks = []models.ChildTask{}
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: tasks})
}

// CompleteTask marks a task as completed and awards points
func CompleteTask(c *gin.Context) {
	var req models.CompleteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "请提供任务ID"})
		return
	}

	// Get task info
	var taskName string
	var taskPoints int
	var taskType string
	err := database.DB.QueryRow(
		"SELECT name, points, task_type FROM tasks WHERE id = ? AND deleted = 0",
		req.TaskID).Scan(&taskName, &taskPoints, &taskType)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: "任务不存在或已删除"})
		return
	}

	today := time.Now().Format("2006-01-02")

	// Check if already completed
	if taskType == "once" {
		var count int
		database.DB.QueryRow("SELECT COUNT(*) FROM child_progress WHERE task_id = ?", req.TaskID).Scan(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "该一次性任务已经完成过了"})
			return
		}
	} else {
		var count int
		database.DB.QueryRow(
			"SELECT COUNT(*) FROM daily_task_status WHERE task_id = ? AND date = ? AND completed = 1",
			req.TaskID, today).Scan(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "该周期性任务今天已经完成过了"})
			return
		}
	}

	// Begin transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	// Record completion
	if taskType == "once" {
		_, err = tx.Exec("INSERT INTO child_progress (task_id) VALUES (?)", req.TaskID)
	} else {
		_, err = tx.Exec(
			"INSERT OR REPLACE INTO daily_task_status (task_id, date, completed) VALUES (?, ?, 1)",
			req.TaskID, today)
	}
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	// Update points
	_, err = tx.Exec("UPDATE current_points SET total_points = total_points + ? WHERE id = 1", taskPoints)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	// Record point history
	reason := fmt.Sprintf("完成任务: %s", taskName)
	_, err = tx.Exec(
		"INSERT INTO point_history (points_change, reason, type) VALUES (?, ?, 'task')",
		taskPoints, reason)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("太棒了！完成「%s」，获得 %d 积分！", taskName, taskPoints),
	})
}
