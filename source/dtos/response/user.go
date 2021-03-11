package response

import "time"

//UserResponse ...
type UserResponse struct {
	UserID    int64         `json:"user_id"`
	FirstName string        `json:"first_name"`
	LastName  string        `json:"last_name"`
	Email     string        `json:"email,omitempty"`
	Phone     string        `json:"phone,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	Role      *RoleResponse `json:"role,omitempty"`
}
