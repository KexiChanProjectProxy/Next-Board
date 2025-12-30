package handler

import (
	"net/http"
	"time"

	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/repository"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo      repository.UserRepository
	nodeRepo      repository.NodeRepository
	planRepo      repository.PlanRepository
	accountingSvc service.AccountingService
	authService   service.AuthService
}

func NewUserHandler(
	userRepo repository.UserRepository,
	nodeRepo repository.NodeRepository,
	planRepo repository.PlanRepository,
	accountingSvc service.AccountingService,
	authService service.AuthService,
) *UserHandler {
	return &UserHandler{
		userRepo:      userRepo,
		nodeRepo:      nodeRepo,
		planRepo:      planRepo,
		accountingSvc: accountingSvc,
		authService:   authService,
	}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)

	user, err := h.userRepo.FindByID(userID)
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
		"user": gin.H{
			"id":                 user.ID,
			"email":              user.Email,
			"role":               user.Role,
			"plan_id":            user.PlanID,
			"telegram_chat_id":   user.TelegramChatID,
			"telegram_linked_at": user.TelegramLinkedAt,
			"created_at":         user.CreatedAt,
		},
	})
}

func (h *UserHandler) GetMyPlan(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			},
		})
		return
	}

	if user.PlanID == nil {
		c.JSON(http.StatusOK, gin.H{
			"plan": nil,
		})
		return
	}

	plan, err := h.planRepo.FindByIDWithLabels(*user.PlanID)
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

func (h *UserHandler) GetMyNodes(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			},
		})
		return
	}

	if user.PlanID == nil {
		c.JSON(http.StatusOK, gin.H{
			"nodes": []interface{}{},
		})
		return
	}

	plan, err := h.planRepo.FindByIDWithLabels(*user.PlanID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"nodes": []interface{}{},
		})
		return
	}

	// Get all nodes
	allNodes, err := h.nodeRepo.FindActiveNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch nodes",
			},
		})
		return
	}

	// Filter nodes that have at least one label matching the plan
	planLabelIDs := make(map[uint64]bool)
	for _, label := range plan.Labels {
		planLabelIDs[label.ID] = true
	}

	var allowedNodes []interface{}
	for _, node := range allNodes {
		hasMatchingLabel := false
		for _, label := range node.Labels {
			if planLabelIDs[label.ID] {
				hasMatchingLabel = true
				break
			}
		}

		if hasMatchingLabel {
			allowedNodes = append(allowedNodes, gin.H{
				"id":              node.ID,
				"name":            node.Name,
				"node_type":       node.NodeType,
				"host":            node.Host,
				"port":            node.Port,
				"node_multiplier": node.NodeMultiplier,
				"status":          node.Status,
				"labels":          node.Labels,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": allowedNodes,
	})
}

func (h *UserHandler) GetMyUsage(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)

	usage, err := h.accountingSvc.GetCurrentUsage(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"usage": gin.H{
				"real_bytes_up":       0,
				"real_bytes_down":     0,
				"billable_bytes_up":   0,
				"billable_bytes_down": 0,
				"period_start":        nil,
				"period_end":          nil,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"usage": gin.H{
			"real_bytes_up":       usage.RealBytesUp,
			"real_bytes_down":     usage.RealBytesDown,
			"billable_bytes_up":   usage.BillableBytesUp,
			"billable_bytes_down": usage.BillableBytesDown,
			"period_start":        usage.PeriodStart,
			"period_end":          usage.PeriodEnd,
		},
	})
}

func (h *UserHandler) GetMyUsageHistory(c *gin.Context) {
	_ = c.MustGet("user_id").(uint64) // userID will be used when Prometheus integration is implemented

	// Default to last 30 days
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	// TODO: Parse range parameter if provided
	// range := c.Query("range")

	c.JSON(http.StatusOK, gin.H{
		"message": "Prometheus integration not yet implemented",
		"note":    "This endpoint will query Prometheus for historical usage data",
		"params": gin.H{
			"start": startTime,
			"end":   endTime,
		},
	})
}

type TelegramLinkResponse struct {
	LinkToken string `json:"link_token"`
	ExpiresIn int    `json:"expires_in"`
}

func (h *UserHandler) GenerateTelegramLink(c *gin.Context) {
	token, err := h.authService.GenerateTelegramLinkToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to generate link token",
			},
		})
		return
	}

	// TODO: Store this token with user_id and expiration in cache/db

	c.JSON(http.StatusOK, gin.H{
		"link_token": token,
		"expires_in": 300, // 5 minutes
		"instructions": "Send this token to the bot using /link <token>",
	})
}
