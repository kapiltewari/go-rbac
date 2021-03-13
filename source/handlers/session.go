package handlers

import (
	"database/sql"
	"fmt"
	"go-rbac/db/models"
	"go-rbac/source/dtos/request"
	"go-rbac/source/dtos/response"
	"go-rbac/source/utils"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// RegisterUser ...
func (h *Handler) RegisterUser(c *fiber.Ctx) error {
	var req request.RegistrationRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "")
	}

	//validate
	errors := utils.ValidateStruct(req)
	if errors != nil {
		return utils.SendValidationError(c, errors)
	}

	//check if user is already registered or not
	exists, err := models.Users(models.UserWhere.Email.EQ(req.Email)).Exists(c.Context(), h.DB)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}
	if exists {
		return utils.SendError(c, fiber.StatusConflict, "email already have been registered")
	}

	//hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	var user models.User

	//check if user is registering with admin email
	sudoEmails := []string{os.Getenv("SUDO_EMAIL"), os.Getenv("SUDO_EMAIL2")}
	for _, email := range sudoEmails {
		if email == req.Email {
			//set admin role
			user.RoleID = 4
			break
		} else {
			//set user role
			user.RoleID = 1
		}
	}
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email
	user.Phone = req.Phone
	user.Password = string(hashedPassword)

	//otp for email verification
	otp := utils.GenerateOTP()

	//send otp
	// sent := utils.SendOTP(user.Email, otp)
	// if !sent {
	// 	return utils.SendError(c, fiber.StatusServiceUnavailable, "")
	// }

	//insert user
	err = user.Insert(c.Context(), h.DB, boil.Infer())
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	//set verification code to redis key for 10 minutes
	emailVerificationCode := fmt.Sprintf("verification-code-email-%v", user.Email)
	if err := h.Redis.Set(c.Context(), emailVerificationCode, otp, time.Duration(10*time.Minute)).Err(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	return utils.SendResponse(c, fiber.StatusCreated, "user successfully registered")
}

//AccountActivation activates a new registered account
func (h *Handler) AccountActivation(c *fiber.Ctx) error {
	var req request.AccountActivationRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "")
	}

	//validate
	errors := utils.ValidateStruct(req)
	if errors != nil {
		return utils.SendValidationError(c, errors)
	}

	//fetch user
	user, err := models.Users(models.UserWhere.Email.EQ(req.Email)).One(c.Context(), h.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.SendError(c, fiber.StatusNotFound, "user does not exist")
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	//check if the user is active already or not
	if user.Active == true {
		return utils.SendError(c, fiber.StatusUnprocessableEntity, "user is active already")
	}

	//get the key
	emailVerificationCode := fmt.Sprintf("verification-code-email-%v", req.Email)
	code, err := h.Redis.Get(c.Context(), emailVerificationCode).Result()
	if err != nil {
		return utils.SendError(c, fiber.StatusUnprocessableEntity, "")
	}

	//check if redis code is matched with req.Code
	if code != req.Code {
		return utils.SendError(c, fiber.StatusUnprocessableEntity, "")
	}

	//update user status to active
	user.Active = true
	user.Update(c.Context(), h.DB, boil.Infer())

	//delete that key
	h.Redis.Del(c.Context(), fmt.Sprintf("verification-code-email-%v", req.Email))

	return utils.SendResponse(c, fiber.StatusOK, "user activated")
}

//LoginUser ...
func (h *Handler) LoginUser(c *fiber.Ctx) error {
	var req request.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "")
	}

	//validate
	errors := utils.ValidateStruct(req)
	if errors != nil {
		return utils.SendValidationError(c, errors)
	}

	//get user by email
	existingUser, err := models.Users(qm.Load(models.UserRels.Role), models.UserWhere.Email.EQ(req.Email)).One(c.Context(), h.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.SendError(c, fiber.StatusNotFound, "invalid login details")

		}
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	//compare user hashed password and entered password and return if any error
	if err := utils.MatchPassword(existingUser.Password, req.Password); err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "invalid login details")
	}

	//if password match, check if the user is active or not
	if existingUser.Active == false {
		return utils.SendError(c, fiber.StatusForbidden, "user is not active")
	}

	//else generate tokens
	accessToken, refreshToken, err := h.GenerateTokens(c, existingUser.R.Role.Name, strconv.FormatInt(existingUser.UserID, 10))
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	//set httpOnly cookies for browser clients
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  time.Now().Local().Add(time.Minute * 60),
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   3600, //1 hour
		SameSite: "strict",
		HTTPOnly: true,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Local().Add(time.Hour * 24 * 10),
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   864000, //10 days
		SameSite: "strict",
		HTTPOnly: true,
	})

	//for mobile clients
	return utils.SendResponse(c, fiber.StatusOK, &response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// AccountReverification user requests for a new activation code
func (h *Handler) AccountReverification(c *fiber.Ctx) error {
	var req request.AccountReverificationRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "")
	}

	//validate
	errors := utils.ValidateStruct(req)
	if errors != nil {
		return utils.SendValidationError(c, errors)
	}

	//check if user exists with given email
	user, err := models.Users(models.UserWhere.Email.EQ(req.Email)).One(c.Context(), h.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.SendError(c, fiber.StatusNotFound, "user does not exist")
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	//if user is active already
	if user.Active == true {
		return utils.SendError(c, fiber.StatusUnprocessableEntity, "user is active already")
	}

	//match phone numbers
	if req.Phone != user.Phone {
		return utils.SendError(c, fiber.StatusUnprocessableEntity, "")
	}

	//generate a new otp
	otp := utils.GenerateOTP()

	//send otp
	// sent := utils.SendOTP(req.Email, otp)
	// if !sent {
	// 	return utils.SendError(c, fiber.StatusServiceUnavailable, functionName, nil)
	// }

	//set verification code to redis key for 10 minutes
	emailVerificationCode := fmt.Sprintf("verification-code-email-%v", req.Email)
	if err := h.Redis.Set(c.Context(), emailVerificationCode, otp, time.Duration(10*time.Minute)).Err(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}
	return utils.SendResponse(c, fiber.StatusOK, "new verification code is sent")
}

//ForgotPassword ...
func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var req request.ForgotPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "")
	}

	//validate
	errors := utils.ValidateStruct(req)
	if errors != nil {
		return utils.SendValidationError(c, errors)
	}

	//fetch user
	user, err := models.Users(models.UserWhere.Email.EQ(req.Email)).One(c.Context(), h.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.SendError(c, fiber.StatusNotFound, "user does not exist")

		}
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	//generate a new password reset token
	passwordResetToken, err := utils.GeneratePasswordResetToken(user.Password, user.Email, user.UserID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	//create link
	passwordResetLink := fmt.Sprintf("%v/reset-password/%v/%v", os.Getenv("URL"), user.UserID, passwordResetToken)
	log.Println(passwordResetLink)

	//send link
	// sent := utils.SendPasswordResetLink(req.Email, passwordResetLink)
	// if !sent {
	// 	return utils.SendError(c, fiber.StatusServiceUnavailable, functionName, nil)
	// }

	return utils.SendResponse(c, fiber.StatusOK, "password reset link is sent")
}

//ResetPassword updates the user's password
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	//get params
	paramID := c.Params("id")
	passwordResetToken := c.Params("token")

	//request body
	var req request.ResetPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "")
	}

	//validate request body
	errors := utils.ValidateStruct(req)
	if errors != nil {
		return utils.SendValidationError(c, errors)
	}

	//convert to int
	userID, err := strconv.Atoi(paramID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "")
	}

	//fetch user
	user, err := models.Users(models.UserWhere.UserID.EQ(int64(userID))).One(c.Context(), h.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.SendError(c, fiber.StatusNotFound, "user does not exist")
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	//parse claims
	claims, err := utils.ValidatePasswordResetToken(user.Password, passwordResetToken)

	//match email
	if user.Email != claims.Email {
		return utils.SendError(c, fiber.StatusUnprocessableEntity, "")
	}

	//hash new password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	//update user password
	user.Password = string(hashedPassword)
	user.Update(c.Context(), h.DB, boil.Infer())

	return utils.SendResponse(c, fiber.StatusOK, "password successfully reset")
}

//LogoutUser ...
func (h *Handler) LogoutUser(c *fiber.Ctx) error {
	var refreshToken string

	//check for browser cookie
	refreshToken = c.Cookies("refresh_token")

	//if empty string then check for refresh token in request body
	if refreshToken == "" {
		var req request.RefreshTokensRequest

		if err := c.BodyParser(&req); err != nil {
			return utils.SendError(c, fiber.StatusBadRequest, "")
		}

		//validate
		errors := utils.ValidateStruct(req)
		if errors != nil {
			return utils.SendValidationError(c, errors)
		}

		//set refresh token
		refreshToken = req.RefreshToken
	}

	//validate refresh token
	claims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "")
	}

	//get redis key refresh-token-user-id-%s
	jti, err := h.Redis.Get(c.Context(), fmt.Sprintf("refresh-token-user-id-%s", claims.Subject)).Result()
	if err != nil {
		return utils.SendError(c, fiber.StatusUnprocessableEntity, "")
	}

	//match jti
	if jti != claims.Jti {
		return utils.SendError(c, fiber.StatusUnauthorized, "")
	}

	//delete httpOnly cookies
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Local(),
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   -1,
		SameSite: "strict",
		HTTPOnly: true,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Local(),
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   -1,
		SameSite: "strict",
		HTTPOnly: true,
	})

	//delete redis key
	h.Redis.Del(c.Context(), fmt.Sprintf("refresh-token-user-id-%s", claims.Subject))

	return utils.SendResponse(c, fiber.StatusNoContent, nil)
}
