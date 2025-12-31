package handler

import (
	"net/http"
	"strconv"

	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/models"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/repository"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	userRepo    repository.UserRepository
	nodeRepo    repository.NodeRepository
	planRepo    repository.PlanRepository
	labelRepo   repository.LabelRepository
	uuidRepo    repository.UUIDRepository
	authService service.AuthService
}

func NewAdminHandler(
	userRepo repository.UserRepository,
	nodeRepo repository.NodeRepository,
	planRepo repository.PlanRepository,
	labelRepo repository.LabelRepository,
	uuidRepo repository.UUIDRepository,
	authService service.AuthService,
) *AdminHandler {
	return &AdminHandler{
		userRepo:    userRepo,
		nodeRepo:    nodeRepo,
		planRepo:    planRepo,
		labelRepo:   labelRepo,
		uuidRepo:    uuidRepo,
		authService: authService,
	}
}

// User management

type CreateUserRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6"`
	Role     string  `json:"role" binding:"required,oneof=admin user"`
	PlanID   *uint64 `json:"plan_id"`
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	user, err := h.authService.Register(req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "USER_CREATION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	// Assign plan if provided
	if req.PlanID != nil {
		user.PlanID = req.PlanID
		h.userRepo.Update(user)
	}

	// Generate UUID for user
	userUUID := &models.UserUUID{
		UserID: user.ID,
		UUID:   uuid.New().String(),
	}
	h.uuidRepo.Create(userUUID)

	c.JSON(http.StatusCreated, gin.H{
		"user": user,
	})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	users, total, err := h.userRepo.List(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch users",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"total":  total,
			"page":   page,
			"limit":  limit,
			"pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid user ID",
			},
		})
		return
	}

	user, err := h.userRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

type UpdateUserRequest struct {
	Email  *string `json:"email" binding:"omitempty,email"`
	PlanID *uint64 `json:"plan_id"`
	Banned *bool   `json:"banned"`
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid user ID",
			},
		})
		return
	}

	user, err := h.userRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			},
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.PlanID != nil {
		user.PlanID = req.PlanID
	}
	if req.Banned != nil {
		user.Banned = *req.Banned
	}

	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "UPDATE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid user ID",
			},
		})
		return
	}

	if err := h.userRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DELETE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// Node management

type CreateNodeRequest struct {
	Name           string   `json:"name" binding:"required"`
	NodeType       string   `json:"node_type" binding:"required"`
	Host           string   `json:"host" binding:"required"`
	Port           uint     `json:"port" binding:"required"`
	ProtocolConfig string   `json:"protocol_config"`
	NodeMultiplier float64  `json:"node_multiplier"`
	LabelIDs       []uint64 `json:"label_ids"`
}

func (h *AdminHandler) CreateNode(c *gin.Context) {
	var req CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	if req.NodeMultiplier == 0 {
		req.NodeMultiplier = 1.0
	}

	node := &models.Node{
		Name:           req.Name,
		NodeType:       req.NodeType,
		Host:           req.Host,
		Port:           req.Port,
		ProtocolConfig: req.ProtocolConfig,
		NodeMultiplier: req.NodeMultiplier,
		Status:         "active",
	}

	if err := h.nodeRepo.Create(node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "NODE_CREATION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	// Add labels
	for _, labelID := range req.LabelIDs {
		h.nodeRepo.AddLabel(node.ID, labelID)
	}

	c.JSON(http.StatusCreated, gin.H{
		"node": node,
	})
}

func (h *AdminHandler) ListNodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	nodes, total, err := h.nodeRepo.List(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch nodes",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": nodes,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *AdminHandler) GetNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid node ID",
			},
		})
		return
	}

	node, err := h.nodeRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NODE_NOT_FOUND",
				"message": "Node not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"node": node,
	})
}

type UpdateNodeRequest struct {
	Name           *string  `json:"name"`
	NodeType       *string  `json:"node_type"`
	Host           *string  `json:"host"`
	Port           *uint    `json:"port"`
	ProtocolConfig *string  `json:"protocol_config"`
	NodeMultiplier *float64 `json:"node_multiplier"`
	Status         *string  `json:"status"`
	LabelIDs       []uint64 `json:"label_ids"`
}

func (h *AdminHandler) UpdateNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid node ID",
			},
		})
		return
	}

	node, err := h.nodeRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NODE_NOT_FOUND",
				"message": "Node not found",
			},
		})
		return
	}

	var req UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	if req.Name != nil {
		node.Name = *req.Name
	}
	if req.NodeType != nil {
		node.NodeType = *req.NodeType
	}
	if req.Host != nil {
		node.Host = *req.Host
	}
	if req.Port != nil {
		node.Port = *req.Port
	}
	if req.ProtocolConfig != nil {
		node.ProtocolConfig = *req.ProtocolConfig
	}
	if req.NodeMultiplier != nil {
		node.NodeMultiplier = *req.NodeMultiplier
	}
	if req.Status != nil {
		node.Status = *req.Status
	}

	if err := h.nodeRepo.Update(node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "UPDATE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	// Update labels if provided
	if req.LabelIDs != nil {
		// Remove all existing labels
		if node.Labels != nil {
			for _, label := range node.Labels {
				h.nodeRepo.RemoveLabel(node.ID, label.ID)
			}
		}
		// Add new labels
		for _, labelID := range req.LabelIDs {
			h.nodeRepo.AddLabel(node.ID, labelID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"node": node,
	})
}

func (h *AdminHandler) DeleteNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid node ID",
			},
		})
		return
	}

	if err := h.nodeRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DELETE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Node deleted successfully",
	})
}

// Plan management

type CreatePlanRequest struct {
	Name           string   `json:"name" binding:"required"`
	QuotaBytes     uint64   `json:"quota_bytes" binding:"required"`
	ResetPeriod    string   `json:"reset_period" binding:"required,oneof=none daily weekly monthly yearly"`
	BaseMultiplier float64  `json:"base_multiplier"`
	LabelIDs       []uint64 `json:"label_ids"`
}

func (h *AdminHandler) CreatePlan(c *gin.Context) {
	var req CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	if req.BaseMultiplier == 0 {
		req.BaseMultiplier = 1.0
	}

	plan := &models.Plan{
		Name:           req.Name,
		QuotaBytes:     req.QuotaBytes,
		ResetPeriod:    req.ResetPeriod,
		BaseMultiplier: req.BaseMultiplier,
	}

	if err := h.planRepo.Create(plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "PLAN_CREATION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	// Add labels
	for _, labelID := range req.LabelIDs {
		h.planRepo.AddLabel(plan.ID, labelID)
	}

	c.JSON(http.StatusCreated, gin.H{
		"plan": plan,
	})
}

func (h *AdminHandler) ListPlans(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	plans, total, err := h.planRepo.List(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch plans",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *AdminHandler) GetPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid plan ID",
			},
		})
		return
	}

	plan, err := h.planRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "PLAN_NOT_FOUND",
				"message": "Plan not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plan": plan,
	})
}

type UpdatePlanRequest struct {
	Name           *string  `json:"name"`
	QuotaBytes     *uint64  `json:"quota_bytes"`
	ResetPeriod    *string  `json:"reset_period"`
	BaseMultiplier *float64 `json:"base_multiplier"`
	LabelIDs       []uint64 `json:"label_ids"`
}

func (h *AdminHandler) UpdatePlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid plan ID",
			},
		})
		return
	}

	plan, err := h.planRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "PLAN_NOT_FOUND",
				"message": "Plan not found",
			},
		})
		return
	}

	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	if req.Name != nil {
		plan.Name = *req.Name
	}
	if req.QuotaBytes != nil {
		plan.QuotaBytes = *req.QuotaBytes
	}
	if req.ResetPeriod != nil {
		plan.ResetPeriod = *req.ResetPeriod
	}
	if req.BaseMultiplier != nil {
		plan.BaseMultiplier = *req.BaseMultiplier
	}

	if err := h.planRepo.Update(plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "UPDATE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	// Update labels if provided
	if req.LabelIDs != nil {
		// Remove all existing labels
		if plan.Labels != nil {
			for _, label := range plan.Labels {
				h.planRepo.RemoveLabel(plan.ID, label.ID)
			}
		}
		// Add new labels
		for _, labelID := range req.LabelIDs {
			h.planRepo.AddLabel(plan.ID, labelID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"plan": plan,
	})
}

func (h *AdminHandler) DeletePlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid plan ID",
			},
		})
		return
	}

	if err := h.planRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DELETE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plan deleted successfully",
	})
}

// Label management

type CreateLabelRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (h *AdminHandler) CreateLabel(c *gin.Context) {
	var req CreateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	label := &models.Label{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.labelRepo.Create(label); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "LABEL_CREATION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"label": label,
	})
}

func (h *AdminHandler) ListLabels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	labels, total, err := h.labelRepo.List(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch labels",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"labels": labels,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *AdminHandler) GetLabel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid label ID",
			},
		})
		return
	}

	label, err := h.labelRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "LABEL_NOT_FOUND",
				"message": "Label not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"label": label,
	})
}

type UpdateLabelRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (h *AdminHandler) UpdateLabel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid label ID",
			},
		})
		return
	}

	label, err := h.labelRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "LABEL_NOT_FOUND",
				"message": "Label not found",
			},
		})
		return
	}

	var req UpdateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	if req.Name != nil {
		label.Name = *req.Name
	}
	if req.Description != nil {
		label.Description = *req.Description
	}

	if err := h.labelRepo.Update(label); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "UPDATE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"label": label,
	})
}

func (h *AdminHandler) DeleteLabel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid label ID",
			},
		})
		return
	}

	if err := h.labelRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DELETE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Label deleted successfully",
	})
}
