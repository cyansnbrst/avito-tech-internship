package models

// Auth request
type AuthRequest struct {
	Username string `json:"username" validate:"required,min=4,max=20"`
	Password string `json:"password" validate:"required,min=4,max=20"`
}

// Auth response
type AuthResponse struct {
	Token string `json:"token"`
}
