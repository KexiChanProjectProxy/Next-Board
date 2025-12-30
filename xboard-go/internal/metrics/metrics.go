package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	ActiveNodes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_nodes",
			Help: "Number of active nodes",
		},
	)

	TrafficReportsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "traffic_reports_total",
			Help: "Total number of traffic reports received",
		},
		[]string{"node_id"},
	)

	TelegramNotificationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_notifications_total",
			Help: "Total number of Telegram notifications sent",
		},
		[]string{"type"},
	)

	AccountingErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "accounting_errors_total",
			Help: "Total number of accounting errors",
		},
	)

	UserTrafficBytes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_traffic_bytes_total",
			Help: "Total traffic in bytes per user",
		},
		[]string{"user_id", "direction", "type"},
	)

	OnlineUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "online_users_total",
			Help: "Number of currently online users",
		},
	)
)

func RecordTraffic(userID uint64, upload, download, billableUp, billableDown uint64) {
	userIDStr := string(rune(userID))

	UserTrafficBytes.WithLabelValues(userIDStr, "up", "real").Add(float64(upload))
	UserTrafficBytes.WithLabelValues(userIDStr, "down", "real").Add(float64(download))
	UserTrafficBytes.WithLabelValues(userIDStr, "up", "billable").Add(float64(billableUp))
	UserTrafficBytes.WithLabelValues(userIDStr, "down", "billable").Add(float64(billableDown))
}
