package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"credits/database"
	"credits/models"

	"github.com/gin-gonic/gin"
)

// GetRedeemItems returns all active redeem items
func GetRedeemItems(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, name, points_required, active FROM redeem_items WHERE active = 1 ORDER BY points_required ASC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	defer rows.Close()

	var items []models.RedeemItem
	for rows.Next() {
		var item models.RedeemItem
		if err := rows.Scan(&item.ID, &item.Name, &item.PointsRequired, &item.Active); err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
			return
		}
		items = append(items, item)
	}

	if items == nil {
		items = []models.RedeemItem{}
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: items})
}

// CreateRedeemItem adds a new redeem item
func CreateRedeemItem(c *gin.Context) {
	var req struct {
		Name           string `json:"name" binding:"required"`
		PointsRequired int    `json:"points_required" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "请填写完整信息"})
		return
	}

	if req.PointsRequired <= 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "所需积分必须大于0"})
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO redeem_items (name, points_required) VALUES (?, ?)",
		req.Name, req.PointsRequired)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "奖品添加成功", Data: map[string]int64{"id": id}})
}

// DeleteRedeemItem soft-deletes a redeem item
func DeleteRedeemItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "无效的奖品ID"})
		return
	}

	_, err = database.DB.Exec("UPDATE redeem_items SET active = 0 WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "奖品已删除"})
}

// RedeemItem exchanges points for a prize
func RedeemItem(c *gin.Context) {
	var req models.RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "请提供奖品ID"})
		return
	}

	// Get redeem item info
	var itemName string
	var pointsRequired int
	err := database.DB.QueryRow(
		"SELECT name, points_required FROM redeem_items WHERE id = ? AND active = 1",
		req.ItemID).Scan(&itemName, &pointsRequired)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: "奖品不存在或已下架"})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	// Check if enough points
	var currentPoints int
	err = tx.QueryRow("SELECT total_points FROM current_points WHERE id = 1").Scan(&currentPoints)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	if currentPoints < pointsRequired {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("积分不足！需要 %d 积分，当前只有 %d 积分", pointsRequired, currentPoints),
		})
		return
	}

	// Deduct points
	_, err = tx.Exec("UPDATE current_points SET total_points = total_points - ? WHERE id = 1", pointsRequired)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	// Record history
	reason := fmt.Sprintf("兑换奖品: %s", itemName)
	_, err = tx.Exec(
		"INSERT INTO point_history (points_change, reason, type) VALUES (?, ?, 'redeem')",
		-pointsRequired, reason)
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
		Message: fmt.Sprintf("兑换成功！消耗 %d 积分获得了「%s」", pointsRequired, itemName),
	})
}
