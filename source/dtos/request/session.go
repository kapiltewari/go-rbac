package request

//RegistrationRequest for user registration
type RegistrationRequest struct {
	FirstName string `json:"first_name" validate:"required,alpha,min=3,max=32"`
	LastName  string `json:"last_name" validate:"required,alpha,min=3,max=32"`
	Email     string `json:"email" validate:"required,email,min=8,max=255"`
	Phone     string `json:"phone" validate:"required,min=10,max=10"`
	Password  string `json:"password" validate:"required,min=8"`
}

//LoginRequest for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,min=8,max=255"`
	Password string `json:"password" validate:"required"`
}

//AccountActivationRequest to activate a new registered account
type AccountActivationRequest struct {
	Code  string `json:"code" validate:"required"`
	Email string `json:"email" validate:"required,email,min=8,max=255"`
}

//NewActivationCodeRequest to request a new email verification code
type NewActivationCodeRequest struct {
	Email string `json:"email" validate:"required,email,min=8,max=255"`
	Phone string `json:"phone" validate:"required,min=10,max=10"`
}

//RefreshTokensRequest for refreshing tokens
type RefreshTokensRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

//ForgotPasswordRequest ...
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email,min=8,max=255"`
}

//ResetPasswordRequest ...
type ResetPasswordRequest struct {
	Password string `json:"password" validate:"required,min=8"`
}
