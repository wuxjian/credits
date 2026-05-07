package handlers

import (
	"net/http"

	"credits/database"
	"credits/models"

	"github.com/gin-gonic/gin"
)

// GetPoints returns the current total points
func GetPoints(c *gin.Context) {
	var totalPoints int
	err := database.DB.QueryRow("SELECT total_points FROM current_points WHERE id = 1").Scan(&totalPoints)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    models.PointsResponse{TotalPoints: totalPoints},
	})
}

// AdjustPoints allows parent to manually add or deduct points
func AdjustPoints(c *gin.Context) {
	var req models.PointAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "请填写完整信息"})
		return
	}

	if req.PointsChange == 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "积分变动值不能为0"})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	// Check if points would go negative
	var currentPoints int
	err = tx.QueryRow("SELECT total_points FROM current_points WHERE id = 1").Scan(&currentPoints)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	if currentPoints+req.PointsChange < 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "积分不足，无法扣除"})
		return
	}

	// Update points
	_, err = tx.Exec("UPDATE current_points SET total_points = total_points + ? WHERE id = 1", req.PointsChange)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	// Record history
	_, err = tx.Exec(
		"INSERT INTO point_history (points_change, reason, type) VALUES (?, ?, 'manual')",
		req.PointsChange, req.Reason)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "积分调整成功"})
}

// GetPointHistory returns all point history records
func GetPointHistory(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, points_change, reason, type, created_at FROM point_history ORDER BY id DESC LIMIT 100")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	defer rows.Close()

	var history []models.PointHistory
	for rows.Next() {
		var h models.PointHistory
		if err := rows.Scan(&h.ID, &h.PointsChange, &h.Reason, &h.Type, &h.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
			return
		}
		history = append(history, h)
	}

	if history == nil {
		history = []models.PointHistory{}
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: history})
}
