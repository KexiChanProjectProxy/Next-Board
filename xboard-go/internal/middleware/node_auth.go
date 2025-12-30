package middleware

import (
	"net/http"
	"strconv"

	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/config"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/repository"

	"github.com/gin-gonic/gin"
)

func NodeAuthMiddleware(cfg *config.NodeConfig, nodeRepo repository.NodeRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			token = c.PostForm("token")
		}

		if token != cfg.ServerToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		nodeIDStr := c.Query("node_id")
		if nodeIDStr == "" {
			nodeIDStr = c.PostForm("node_id")
		}

		if nodeIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "node_id is required",
			})
			c.Abort()
			return
		}

		nodeID, err := strconv.ParseUint(nodeIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid node_id",
			})
			c.Abort()
			return
		}

		node, err := nodeRepo.FindByID(nodeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Server does not exist",
			})
			c.Abort()
			return
		}

		nodeType := c.Query("node_type")
		if nodeType == "" {
			nodeType = c.PostForm("node_type")
		}

		// Normalize node type (like Xboard does)
		normalizedType := normalizeNodeType(nodeType)
		if normalizedType != "" && node.NodeType != normalizedType {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Node type mismatch",
			})
			c.Abort()
			return
		}

		c.Set("node_id", nodeID)
		c.Set("node_info", node)
		c.Next()
	}
}

func normalizeNodeType(nodeType string) string {
	switch nodeType {
	case "v2ray":
		return "vmess"
	case "hysteria2":
		return "hysteria"
	default:
		return nodeType
	}
}
