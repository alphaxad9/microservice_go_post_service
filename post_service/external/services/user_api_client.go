// github.com/alphaxad9/my-go-backend/post_service/external/services/user_api_client.go
package services

import (
	"fmt"

	"github.com/alphaxad9/my-go-backend/post_service/external"

	"github.com/google/uuid"
)

type UserAPIClient struct {
	httpClient *HTTPClient
	baseURL    string
}

func NewUserAPIClient(httpClient *HTTPClient, baseURL string) *UserAPIClient {
	return &UserAPIClient{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (c *UserAPIClient) GetUserByID(userID uuid.UUID) (*external.UserView, error) {
	url := fmt.Sprintf("%s/users/users/%s/", c.baseURL, userID.String())

	var response map[string]interface{}

	err := c.httpClient.Get(url, &response)
	if err != nil {
		return nil, err
	}

	// handle both:
	// { "user": { ... } }
	// or direct object

	var userData map[string]interface{}

	if u, ok := response["user"].(map[string]interface{}); ok {
		userData = u
	} else {
		userData = response
	}

	userIDStr, okID := userData["user_id"].(string)
	username, okUsername := userData["username"].(string)

	if !okID || !okUsername {
		return nil, fmt.Errorf("invalid user payload: %+v", userData)
	}

	parsedID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	firstName, _ := userData["first_name"].(string)
	lastName, _ := userData["last_name"].(string)

	return &external.UserView{
		UserID:    parsedID,
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}
