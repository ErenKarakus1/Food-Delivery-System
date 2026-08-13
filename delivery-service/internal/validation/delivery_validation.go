package validation

func ValidateStatus(status string) bool {
	switch status {
	case "rejected":
		return true
	case "picked_by_courier":
		return true
	case "delivered":
		return true
	default:
		return false
	}
}

func ValidateStatusTransition(currentStatus string, updateStatus string) bool {
	switch updateStatus {
	case "rejected", "picked_by_courier":
		return currentStatus == "assigned"
	case "delivered":
		return currentStatus == "picked_by_courier"
	default:
		return false
	}
}
