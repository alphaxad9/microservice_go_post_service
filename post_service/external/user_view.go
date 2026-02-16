// my-go-backend/post_service/external/user_view.go
package external

import "github.com/google/uuid"

type UserView struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
}

func (u UserView) FullName() *string {
	full := u.FirstName + " " + u.LastName
	if full == " " {
		return nil
	}
	return &full
}
