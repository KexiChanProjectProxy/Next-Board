package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/metrics"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/models"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/repository"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NodeHandler struct {
	nodeRepo      repository.NodeRepository
	userRepo      repository.UserRepository
	planRepo      repository.PlanRepository
	uuidRepo      repository.UUIDRepository
	onlineRepo    repository.OnlineUserRepository
	accountingSvc service.AccountingService
	logger        *zap.Logger
}

func NewNodeHandler(
	nodeRepo repository.NodeRepository,
	userRepo repository.UserRepository,
	planRepo repository.PlanRepository,
	uuidRepo repository.UUIDRepository,
	onlineRepo repository.OnlineUserRepository,
	accountingSvc service.AccountingService,
	logger *zap.Logger,
) *NodeHandler {
	return &NodeHandler{
		nodeRepo:      nodeRepo,
		userRepo:      userRepo,
		planRepo:      planRepo,
		uuidRepo:      uuidRepo,
		onlineRepo:    onlineRepo,
		accountingSvc: accountingSvc,
		logger:        logger,
	}
}

// GetConfig returns node configuration (GET /config)
func (h *NodeHandler) GetConfig(c *gin.Context) {
	node := c.MustGet("node_info").(*models.Node)

	config := models.NodeConfigDTO{
		Protocol:   node.NodeType,
		ListenIP:   "0.0.0.0",
		ServerPort: node.Port,
		BaseConfig: map[string]interface{}{
			"push_interval": 60,
			"pull_interval": 60,
		},
	}

	// Parse protocol-specific config if available
	if node.ProtocolConfig != "" {
		var protocolConfig map[string]interface{}
		if err := json.Unmarshal([]byte(node.ProtocolConfig), &protocolConfig); err == nil {
			for k, v := range protocolConfig {
				switch k {
				case "network", "tls", "host", "server_name":
					// Add these to top level
					switch k {
					case "network":
						config.Network = v.(string)
					case "tls":
						config.TLS = int(v.(float64))
					}
				}
			}
		}
	}

	// Calculate ETag
	data, _ := json.Marshal(config)
	etag := calculateETag(data)

	// Check If-None-Match header
	if c.GetHeader("If-None-Match") == fmt.Sprintf("\"%s\"", etag) {
		c.Status(http.StatusNotModified)
		return
	}

	c.Header("ETag", fmt.Sprintf("\"%s\"", etag))
	c.JSON(http.StatusOK, config)
}

// GetUsers returns list of users allowed on this node (GET /user)
func (h *NodeHandler) GetUsers(c *gin.Context) {
	nodeID := c.MustGet("node_id").(uint64)

	node, err := h.nodeRepo.FindByIDWithLabels(nodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Server does not exist",
		})
		return
	}

	// Get all users with plans that allow this node's labels
	users, err := h.getAllowedUsers(node)
	if err != nil {
		h.logger.Error("Failed to get allowed users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get users",
		})
		return
	}

	// Build response
	response := gin.H{
		"users": users,
	}

	// Calculate ETag
	data, _ := json.Marshal(response)
	etag := calculateETag(data)

	// Check If-None-Match header
	if c.GetHeader("If-None-Match") == fmt.Sprintf("\"%s\"", etag) {
		c.Status(http.StatusNotModified)
		return
	}

	c.Header("ETag", fmt.Sprintf("\"%s\"", etag))
	c.JSON(http.StatusOK, response)
}

// PushTraffic handles traffic reports from nodes (POST /push)
func (h *NodeHandler) PushTraffic(c *gin.Context) {
	nodeID := c.MustGet("node_id").(uint64)

	var rawData interface{}
	if err := c.ShouldBindJSON(&rawData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    422,
			"message": "Invalid data format",
		})
		return
	}

	reports := parseTrafficData(rawData)
	if len(reports) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "ok",
			"data":    true,
		})
		return
	}

	// Process traffic
	if err := h.accountingSvc.ProcessTrafficReport(nodeID, reports); err != nil {
		h.logger.Error("Failed to process traffic", zap.Error(err))
		metrics.AccountingErrorsTotal.Inc()
	}

	// Update metrics
	metrics.TrafficReportsTotal.WithLabelValues(strconv.FormatUint(nodeID, 10)).Inc()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    true,
	})
}

// PushAlive handles online user reports (POST /alive)
func (h *NodeHandler) PushAlive(c *gin.Context) {
	nodeID := c.MustGet("node_id").(uint64)

	var aliveData models.AliveIPMap
	if err := c.ShouldBindJSON(&aliveData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid data format",
		})
		return
	}

	// Process online users
	for userID, ips := range aliveData {
		for _, ipWithNode := range ips {
			// Parse IP (format: "IP_nodeIdentifier")
			// For simplicity, we'll just use the whole string as IP
			if err := h.onlineRepo.UpsertOnlineUser(userID, nodeID, ipWithNode); err != nil {
				h.logger.Error("Failed to upsert online user",
					zap.Uint64("user_id", userID),
					zap.String("ip", ipWithNode),
					zap.Error(err),
				)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": true,
	})
}

// GetAliveList returns device limit info (GET /alivelist)
func (h *NodeHandler) GetAliveList(c *gin.Context) {
	counts, err := h.onlineRepo.GetAllOnlineDeviceCounts()
	if err != nil {
		h.logger.Error("Failed to get online device counts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get device limits",
		})
		return
	}

	c.JSON(http.StatusOK, models.DeviceLimitDTO{
		Alive: counts,
	})
}

// PushStatus handles node load status reports (POST /status)
func (h *NodeHandler) PushStatus(c *gin.Context) {
	nodeID := c.MustGet("node_id").(uint64)

	var status struct {
		CPU  float64 `json:"cpu" binding:"required,min=0,max=100"`
		Mem  struct {
			Total uint64 `json:"total" binding:"required,min=0"`
			Used  uint64 `json:"used" binding:"required,min=0"`
		} `json:"mem" binding:"required"`
		Swap struct {
			Total uint64 `json:"total" binding:"min=0"`
			Used  uint64 `json:"used" binding:"min=0"`
		} `json:"swap"`
		Disk struct {
			Total uint64 `json:"total" binding:"min=0"`
			Used  uint64 `json:"used" binding:"min=0"`
		} `json:"disk"`
	}

	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    422,
			"message": "Invalid status data",
		})
		return
	}

	// Update node last seen
	if err := h.nodeRepo.UpdateLastSeen(nodeID); err != nil {
		h.logger.Error("Failed to update node last seen", zap.Error(err))
	}

	// In a real implementation, you'd cache this status data
	// For now, we just acknowledge receipt

	c.JSON(http.StatusOK, gin.H{
		"data":    true,
		"code":    0,
		"message": "success",
	})
}

// Helper functions

func (h *NodeHandler) getAllowedUsers(node *models.Node) ([]models.NodeUserDTO, error) {
	// Get node label IDs
	labelIDs := make([]uint64, len(node.Labels))
	for i, label := range node.Labels {
		labelIDs[i] = label.ID
	}

	// Find plans that allow any of these labels
	var plans []models.Plan
	if len(labelIDs) > 0 {
		// Load all plans and filter by matching labels
		allPlans, _, _ := h.planRepo.List(0, 10000)
		for _, plan := range allPlans {
			planLabels, _ := h.planRepo.GetLabels(plan.ID)
			for _, planLabel := range planLabels {
				for _, nodeLabel := range node.Labels {
					if planLabel.ID == nodeLabel.ID {
						plans = append(plans, plan)
						break
					}
				}
			}
		}
	}

	// Get plan IDs
	planIDs := make([]uint64, len(plans))
	for i, plan := range plans {
		planIDs[i] = plan.ID
	}

	// Find users with these plans
	// This is a simplified implementation - in production, use proper DB queries
	var users []models.User
	allUsers, _, _ := h.userRepo.List(0, 10000)
	for _, user := range allUsers {
		if user.PlanID != nil {
			for _, planID := range planIDs {
				if *user.PlanID == planID {
					users = append(users, user)
					break
				}
			}
		}
	}

	// Get UUIDs for users
	uuidMap, err := h.uuidRepo.GetAllUserUUIDs()
	if err != nil {
		return nil, err
	}

	// Build user DTOs
	var userDTOs []models.NodeUserDTO
	for _, user := range users {
		if user.Banned {
			continue
		}

		// Check if user has exceeded quota
		usage, err := h.accountingSvc.GetCurrentUsage(user.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			continue
		}

		if usage != nil && user.Plan != nil {
			totalUsage := usage.BillableBytesUp + usage.BillableBytesDown
			if totalUsage >= user.Plan.QuotaBytes {
				continue
			}
		}

		uuid, exists := uuidMap[user.ID]
		if !exists {
			continue
		}

		userDTOs = append(userDTOs, models.NodeUserDTO{
			ID:          user.ID,
			UUID:        uuid,
			SpeedLimit:  0, // TODO: Add speed limit support
			DeviceLimit: 0, // TODO: Add device limit support
		})
	}

	return userDTOs, nil
}

func parseTrafficData(raw interface{}) []models.TrafficReport {
	var reports []models.TrafficReport

	switch data := raw.(type) {
	case []interface{}:
		// Array format: [[user_id, [upload, download]], ...]
		for _, item := range data {
			itemArr, ok := item.([]interface{})
			if !ok || len(itemArr) != 2 {
				continue
			}

			userID, ok := itemArr[0].(float64)
			if !ok {
				continue
			}

			traffic, ok := itemArr[1].([]interface{})
			if !ok || len(traffic) != 2 {
				continue
			}

			upload, ok1 := traffic[0].(float64)
			download, ok2 := traffic[1].(float64)
			if !ok1 || !ok2 {
				continue
			}

			reports = append(reports, models.TrafficReport{
				UserID:   uint64(userID),
				Upload:   uint64(upload),
				Download: uint64(download),
			})
		}

	case map[string]interface{}:
		// Object format: {"user_id": [upload, download], ...}
		for userIDStr, trafficData := range data {
			userID, err := strconv.ParseUint(userIDStr, 10, 64)
			if err != nil {
				continue
			}

			traffic, ok := trafficData.([]interface{})
			if !ok || len(traffic) != 2 {
				continue
			}

			upload, ok1 := traffic[0].(float64)
			download, ok2 := traffic[1].(float64)
			if !ok1 || !ok2 {
				continue
			}

			reports = append(reports, models.TrafficReport{
				UserID:   userID,
				Upload:   uint64(upload),
				Download: uint64(download),
			})
		}
	}

	return reports
}

func calculateETag(data []byte) string {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:])
}
