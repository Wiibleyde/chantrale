package dto

// MemberDTO represents a Discord guild member in API responses.
type MemberDTO struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}
