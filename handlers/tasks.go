package handlers

import (
	"net/http"
	"strconv"

	"credits/database"
	"credits/models"

	"github.com/gin-gonic/gin"
)

// GetTasks returns all non-deleted tasks (for parent view)
func GetTasks(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, name, points, task_type, deleted, created_at FROM tasks WHERE deleted = 0 ORDER BY id DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Name, &t.Points, &t.TaskType, &t.Deleted, &t.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
			return
		}
		tasks = append(tasks, t)
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: tasks})
}

// CreateTask creates a new task
func CreateTask(c *gin.Context) {
	var req models.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "请填写完整信息"})
		return
	}

	if req.TaskType != "once" && req.TaskType != "daily" {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "任务类型必须是 once 或 daily"})
		return
	}

	if req.Points <= 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "积分必须大于0"})
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO tasks (name, points, task_type) VALUES (?, ?, ?)",
		req.Name, req.Points, req.TaskType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "任务创建成功", Data: map[string]int64{"id": id}})
}

// DeleteTask soft-deletes a task
func DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "无效的任务ID"})
		return
	}

	_, err = database.DB.Exec("UPDATE tasks SET deleted = 1 WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "任务已删除"})
}
