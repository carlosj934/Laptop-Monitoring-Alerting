package notifications

import (
	"github.com/carlosj934/laptop-dashboard-alerting/internal/alerts"
	"gopkg.in/toast.v1"
)

//Notifier sends windows toast notifications
type Notifier struct{}

// NewNotifier creates a new notifier
func NewNotifier() *Notifier {
	return &Notifier{}
}

//Send sends a Windows toast notification for an alert
func (n *Notifier) Send(alert alerts.AlertEvent) error {
	notification := toast.Notification {
		AppID: "Laptop Monitor",
		Title: n.getTitle(alert.AlertType),
		Message: alert.Message,
	}

	return notification.Push()
}

// getTitle returns a user-friendly title for the alert type
func (n *Notifier) getTitle(alertType alerts.AlertType) string {
	switch alertType {
	case alerts.AlertCPU:
		return "⚠️ High CPU Uage"
	case alerts.AlertMemoryOverall:
		return "⚠️ High MemoryUsage"
	case alerts.AlertMemoryProcess:
		return "⚠️ Process Memory Alert"
	case alerts.AlertDisk:
		return "⚠️ Low Disk Space"
	default:
		return "⚠️ System Alert"
	}
}
