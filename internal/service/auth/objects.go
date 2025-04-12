package auth

type UserInfo struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
	Email     string `json:"email" validate:"required,email"`
}

type RegisterRequest struct {
	UserInfo
	Password string `json:"password" validate:"required,min=8,max=100"`
}

type RegisterResponse struct {
	Token string `json:"token"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type LogoutRequest struct {
	AuthToken string `header:"Authorization"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

type RefreshTokenRequest struct {
	AuthToken string `header:"Authorization"`
}

type RefreshTokenResponse struct {
	Token string `json:"token"`
}

type MeRequest struct {
	AuthToken string `header:"Authorization"`
}

type MeResponse struct {
	UserInfo
}
