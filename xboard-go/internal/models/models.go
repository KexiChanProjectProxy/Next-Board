package models

import (
	"time"
)

type User struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Email             string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash      string     `gorm:"not null" json:"-"`
	Role              string     `gorm:"type:enum('admin','user');default:'user'" json:"role"`
	PlanID            *uint64    `gorm:"index" json:"plan_id"`
	Plan              *Plan      `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
	TelegramChatID    *int64     `gorm:"index" json:"telegram_chat_id"`
	TelegramLinkedAt  *time.Time `json:"telegram_linked_at"`
	Banned            bool       `gorm:"default:false" json:"banned"`
	Balance           int        `gorm:"default:0" json:"balance"`                     // Balance in cents
	Discount          *int       `json:"discount"`                                     // Discount percentage
	CommissionType    int        `gorm:"default:0" json:"commission_type"`             // 0: system 1: period 2: onetime
	CommissionRate    *int       `json:"commission_rate"`                              // Commission rate percentage
	CommissionBalance int        `gorm:"default:0" json:"commission_balance"`          // Commission balance in cents
	Token             *string    `gorm:"index;size:32" json:"token,omitempty"`         // User API token
	LastLoginAt       *time.Time `gorm:"index" json:"last_login_at"`                   // Last login timestamp
	LastLoginIP       *string    `gorm:"size:45" json:"last_login_ip"`                 // Last login IP address
	Remarks           *string    `gorm:"type:text" json:"remarks,omitempty"`           // Admin remarks
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Label struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null;size:100" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Plan struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string    `gorm:"uniqueIndex;not null;size:100" json:"name"`
	QuotaBytes     uint64    `gorm:"default:0" json:"quota_bytes"`
	ResetPeriod    string    `gorm:"type:enum('none','daily','weekly','monthly','yearly');default:'monthly'" json:"reset_period"`
	BaseMultiplier float64   `gorm:"type:decimal(10,4);default:1.0" json:"base_multiplier"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Labels         []Label   `gorm:"many2many:plan_labels" json:"labels,omitempty"`
}

type PlanLabel struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	PlanID    uint64    `gorm:"index;not null" json:"plan_id"`
	LabelID   uint64    `gorm:"index;not null" json:"label_id"`
	CreatedAt time.Time `json:"created_at"`
}

type PlanLabelMultiplier struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	PlanID     uint64    `gorm:"index;not null" json:"plan_id"`
	LabelID    uint64    `gorm:"index;not null" json:"label_id"`
	Multiplier float64   `gorm:"type:decimal(10,4);default:1.0" json:"multiplier"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Node struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string     `gorm:"index;not null;size:100" json:"name"`
	NodeType       string     `gorm:"not null;size:50" json:"node_type"`
	Host           string     `gorm:"not null" json:"host"`
	Port           uint       `gorm:"not null" json:"port"`
	ProtocolConfig string     `gorm:"type:json" json:"protocol_config"`
	NodeMultiplier float64    `gorm:"type:decimal(10,4);default:1.0" json:"node_multiplier"`
	Status         string     `gorm:"type:enum('active','inactive','maintenance');default:'active'" json:"status"`
	LastSeenAt     *time.Time `gorm:"index" json:"last_seen_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Labels         []Label    `gorm:"many2many:node_labels" json:"labels,omitempty"`
}

type NodeLabel struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	NodeID    uint64    `gorm:"uniqueIndex:idx_node_label,priority:1;not null" json:"node_id"`
	LabelID   uint64    `gorm:"uniqueIndex:idx_node_label,priority:2;not null" json:"label_id"`
	CreatedAt time.Time `json:"created_at"`
}

type UsagePeriod struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint64    `gorm:"index;not null" json:"user_id"`
	PlanID          uint64    `gorm:"not null" json:"plan_id"`
	PeriodStart     time.Time `gorm:"index;not null" json:"period_start"`
	PeriodEnd       time.Time `gorm:"index;not null" json:"period_end"`
	RealBytesUp     uint64    `gorm:"default:0" json:"real_bytes_up"`
	RealBytesDown   uint64    `gorm:"default:0" json:"real_bytes_down"`
	BillableBytesUp uint64    `gorm:"default:0" json:"billable_bytes_up"`
	BillableBytesDown uint64  `gorm:"default:0" json:"billable_bytes_down"`
	IsCurrent       bool      `gorm:"index;default:true" json:"is_current"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type NodeUsage struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID            uint64    `gorm:"index;not null" json:"user_id"`
	NodeID            uint64    `gorm:"index;not null" json:"node_id"`
	PeriodID          uint64    `gorm:"index;not null" json:"period_id"`
	RealBytesUp       uint64    `gorm:"default:0" json:"real_bytes_up"`
	RealBytesDown     uint64    `gorm:"default:0" json:"real_bytes_down"`
	BillableBytesUp   uint64    `gorm:"default:0" json:"billable_bytes_up"`
	BillableBytesDown uint64    `gorm:"default:0" json:"billable_bytes_down"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type TelegramThreshold struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint64     `gorm:"index;not null" json:"user_id"`
	ThresholdType   string     `gorm:"type:enum('percent','bytes_remaining');not null" json:"threshold_type"`
	ThresholdValue  float64    `gorm:"type:decimal(10,2);not null" json:"threshold_value"`
	Enabled         bool       `gorm:"default:true" json:"enabled"`
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type UserUUID struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"uniqueIndex;not null" json:"user_id"`
	UUID      string    `gorm:"uniqueIndex;not null;size:36" json:"uuid"`
	CreatedAt time.Time `json:"created_at"`
}

type OnlineUser struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"index;not null" json:"user_id"`
	NodeID     uint64    `gorm:"not null" json:"node_id"`
	IPAddress  string    `gorm:"size:45;not null" json:"ip_address"`
	LastSeenAt time.Time `gorm:"index;not null" json:"last_seen_at"`
}

type RefreshToken struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"index;not null" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// DTO for traffic reporting
type TrafficReport struct {
	UserID   uint64 `json:"user_id"`
	Upload   uint64 `json:"upload"`
	Download uint64 `json:"download"`
}

// DTO for user list response (node protocol)
type NodeUserDTO struct {
	ID          uint64 `json:"id"`
	UUID        string `json:"uuid"`
	SpeedLimit  uint64 `json:"speed_limit"`
	DeviceLimit uint   `json:"device_limit"`
}

// DTO for config response (node protocol)
type NodeConfigDTO struct {
	Protocol        string                 `json:"protocol"`
	ListenIP        string                 `json:"listen_ip"`
	ServerPort      uint                   `json:"server_port"`
	Network         string                 `json:"network,omitempty"`
	NetworkSettings map[string]interface{} `json:"networkSettings,omitempty"`
	TLS             int                    `json:"tls,omitempty"`
	BaseConfig      map[string]interface{} `json:"base_config"`
	Routes          []interface{}          `json:"routes,omitempty"`
}

// DTO for online users
type AliveIPMap map[uint64][]string

// DTO for device limit response
type DeviceLimitDTO struct {
	Alive map[uint64]uint `json:"alive"`
}
