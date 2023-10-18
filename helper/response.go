package helper

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type UserResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Points      int    `json:"points"`
}
