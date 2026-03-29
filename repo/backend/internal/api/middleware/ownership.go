package middleware

func OwnsResource(requesterID, targetID int64) bool {
	return requesterID == targetID
}
