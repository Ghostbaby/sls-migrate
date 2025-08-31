package handler

import (
	"net/http"
	"strconv"

	"github.com/Ghostbaby/sls-migrate/internal/models"
	"github.com/Ghostbaby/sls-migrate/internal/service"
	"github.com/gin-gonic/gin"
)

// AlertHandler Alert 处理器
type AlertHandler struct {
	alertService service.AlertService
}

// NewAlertHandler 创建新的 AlertHandler 实例
func NewAlertHandler(alertService service.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

// CreateAlert 创建 Alert
// @Summary 创建 Alert
// @Description 创建新的 Alert 记录
// @Tags Alert
// @Accept json
// @Produce json
// @Param alert body models.Alert true "Alert 信息"
// @Success 201 {object} models.Alert
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /alerts [post]
func (h *AlertHandler) CreateAlert(c *gin.Context) {
	var alert models.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	if err := h.alertService.CreateAlert(c.Request.Context(), &alert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create alert",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, alert)
}

// GetAlertByID 根据 ID 获取 Alert
// @Summary 根据 ID 获取 Alert
// @Description 根据 ID 获取 Alert 详细信息
// @Tags Alert
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} models.Alert
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /alerts/{id} [get]
func (h *AlertHandler) GetAlertByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid alert ID",
			"message": "ID must be a valid integer",
		})
		return
	}

	alert, err := h.alertService.GetAlertByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Alert not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// GetAlertByName 根据名称获取 Alert
// @Summary 根据名称获取 Alert
// @Description 根据名称获取 Alert 详细信息
// @Tags Alert
// @Accept json
// @Produce json
// @Param name path string true "Alert 名称"
// @Success 200 {object} models.Alert
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /alerts/name/{name} [get]
func (h *AlertHandler) GetAlertByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid alert name",
			"message": "Name cannot be empty",
		})
		return
	}

	alert, err := h.alertService.GetAlertByName(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Alert not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// UpdateAlert 更新 Alert
// @Summary 更新 Alert
// @Description 更新 Alert 信息
// @Tags Alert
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Param alert body models.Alert true "Alert 更新信息"
// @Success 200 {object} models.Alert
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /alerts/{id} [put]
func (h *AlertHandler) UpdateAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid alert ID",
			"message": "ID must be a valid integer",
		})
		return
	}

	var alert models.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	alert.ID = uint(id)
	if err := h.alertService.UpdateAlert(c.Request.Context(), &alert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update alert",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// DeleteAlert 删除 Alert
// @Summary 删除 Alert
// @Description 根据 ID 删除 Alert
// @Tags Alert
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /alerts/{id} [delete]
func (h *AlertHandler) DeleteAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid alert ID",
			"message": "ID must be a valid integer",
		})
		return
	}

	if err := h.alertService.DeleteAlert(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete alert",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert deleted successfully",
	})
}

// ListAlerts 获取 Alert 列表
// @Summary 获取 Alert 列表
// @Description 分页获取 Alert 列表
// @Tags Alert
// @Accept json
// @Produce json
// @Param page query int false "页码 (默认: 1)"
// @Param page_size query int false "每页大小 (默认: 20, 最大: 100)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /alerts [get]
func (h *AlertHandler) ListAlerts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	alerts, total, err := h.alertService.ListAlerts(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get alerts",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": alerts,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// ListAlertsByStatus 根据状态获取 Alert 列表
// @Summary 根据状态获取 Alert 列表
// @Description 根据状态分页获取 Alert 列表
// @Tags Alert
// @Accept json
// @Produce json
// @Param status query string true "Alert 状态 (ENABLED/DISABLED)"
// @Param page query int false "页码 (默认: 1)"
// @Param page_size query int false "每页大小 (默认: 20, 最大: 100)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /alerts/status/{status} [get]
func (h *AlertHandler) ListAlertsByStatus(c *gin.Context) {
	status := c.Param("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	alerts, total, err := h.alertService.ListAlertsByStatus(c.Request.Context(), status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get alerts",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": alerts,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}
