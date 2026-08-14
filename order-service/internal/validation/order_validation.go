package validation

func ValidateStatusRequestRestaurant(status string) bool {
	switch status {
	case "pending":
		return true
	case "accepted":
		return true
	case "rejected":
		return true
	case "preparing":
		return true
	case "ready_for_pickup":
		return true
	default:
		return false
	}
}

func ValidateRestaurantStatusTransition(currentStatus string, updateStatus string) bool {
	switch updateStatus {
	case "accepted", "rejected":
		return currentStatus == "pending"
	case "preparing":
		return currentStatus == "accepted"
	case "ready_for_pickup":
		return currentStatus == "preparing"
	default:
		return false
	}
}

func ValidateCourierStatusRequest(status string) bool {
	switch status {
	case "picked_by_courier":
		return true
	case "delivered":
		return true
	default:
		return false
	}
}

func ValidateCourierStatusTransition(currentStatus string, updateStatus string) bool {
	switch updateStatus {
	case "picked_by_courier":
		return currentStatus == "delivery_created"
	case "delivered":
		return currentStatus == "picked_by_courier"
	default:
		return false
	}
}
