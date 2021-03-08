package handlers

import (
	"database/sql"
	"fmt"
	"go-rbac/db/models"
	"go-rbac/source/dtos/request"
	"go-rbac/source/dtos/response"
	"go-rbac/source/utils"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// RegisterUser ...
func (h *Handler) RegisterUser(c *fiber.Ctx) error {
	functionName := "RegisterUser"
	var req request.RegistrationRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.LogAndSendError(c, fiber.StatusBadRequest, functionName, err)
	}

	//validate
	errors := utils.ValidateStruct(req)
	if errors != nil {
		return utils.SendValidationError(c, errors)
	}

	//check if user is already registered or not
	exists, err := models.Users(models.UserWhere.Email.EQ(req.Email)).Exists(c.Context(), h.DB)
	if err != nil {
		return utils.LogAndSendError(c, fiber.StatusInternalServerError, functionName, err)
	}
	if exists {
		return utils.LogAndSendError(c, fiber.StatusConflict, functionName, err)
	}

	//hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.LogAndSendError(c, fiber.StatusInternalServerError, functionName, err)
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
	// sent := utils.SendMail(user.Email, otp)
	// if !sent {
	// 	return utils.LogAndSendError(c, fiber.StatusServiceUnavailable, functionName, err)
	// }

	//insert user
	err = user.Insert(c.Context(), h.DB, boil.Infer())
	if err != nil {
		return utils.LogAndSendError(c, fiber.StatusInternalServerError, functionName, err)
	}

	//set verification code to redis key for 10 minutes
	emailVerificationCode := fmt.Sprintf("verification-code-email-%v", user.Email)
	if err := h.Redis.Set(c.Context(), emailVerificationCode, otp, time.Duration(10*time.Minute)).Err(); err != nil {
		return utils.LogAndSendError(c, fiber.StatusInternalServerError, functionName, err)
	}

	return utils.SendResponse(c, fiber.StatusCreated, "Created. Please verify your email. And then login.")
}

//ActivateAccount activates a new registered account
func (h *Handler) ActivateAccount(c *fiber.Ctx) error {
	functionName := "ActivateAccount"
	var req request.AccountActivationRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.LogAndSendError(c, fiber.StatusBadRequest, functionName, err)
	}

	//validate
	errors := utils.ValidateStruct(req)
	if errors != nil {
		return utils.SendValidationError(c, errors)
	}

	//get the key
	emailVerificationCode := fmt.Sprintf("verification-code-email-%v", req.Email)
	code, err := h.Redis.Get(c.Context(), emailVerificationCode).Result()
	if err != nil {
		return utils.LogAndSendError(c, fiber.StatusUnauthorized, functionName, err)
	}

	//check if redis code is matched with req.Code
	if code != req.Code {
		return utils.LogAndSendError(c, fiber.StatusUnauthorized, functionName, err)
	}

	//update user status to active
	user, err := models.Users(models.UserWhere.Email.EQ(req.Email)).One(c.Context(), h.DB)
	if err != nil {
		return utils.LogAndSendError(c, fiber.StatusInternalServerError, functionName, err)
	}
	user.Active = true
	user.Update(c.Context(), h.DB, boil.Infer())

	//delete that key
	h.Redis.Del(c.Context(), fmt.Sprintf("verification-code-email-%v", req.Email))

	return utils.SendResponse(c, fiber.StatusOK, nil)
}

//LoginUser ...
func (h *Handler) LoginUser(c *fiber.Ctx) error {
	functionName := "LoginUser"

	var req request.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.LogAndSendError(c, fiber.StatusBadRequest, functionName, err)
	}

	//validate
	errors := utils.ValidateStruct(req)
	if errors != nil {
		return utils.SendValidationError(c, errors)
	}

	//get user by email
	existingUser, err := models.Users(qm.Load(qm.Rels(models.UserRels.Role)), models.UserWhere.Email.EQ(req.Email)).One(c.Context(), h.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.LogAndSendError(c, fiber.StatusNotFound, functionName, err)

		}
		return utils.LogAndSendError(c, fiber.StatusInternalServerError, functionName, err)
	}

	//compare user hashed password and entered password and return if any error
	if err := utils.MatchPassword(existingUser.Password, req.Password); err != nil {
		return utils.LogAndSendError(c, fiber.StatusNotFound, functionName, err)
	}

	//if password match, check if the user is active or not
	if existingUser.Active == false {
		return utils.LogAndSendError(c, fiber.StatusForbidden, functionName, err)
	}

	//else generate tokens
	accessToken, refreshToken, err := h.GenerateTokens(c, existingUser.R.Role.Name, strconv.FormatInt(existingUser.UserID, 10))
	if err != nil {
		return utils.LogAndSendError(c, fiber.StatusInternalServerError, functionName, err)
	}

	//set httpOnly cookies
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

	return utils.SendResponse(c, fiber.StatusOK, &response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// GenerateNewCode user requests for a new verification code
// to activate their account, resetting password etc
func (h *Handler) GenerateNewCode(c *fiber.Ctx) error {
	functionName := "GenerateNewCode"

	var req request.GenerateNewCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.LogAndSendError(c, fiber.StatusBadRequest, functionName, err)
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
			return utils.LogAndSendError(c, fiber.StatusNotFound, functionName, err)
		}
		return utils.LogAndSendError(c, fiber.StatusInternalServerError, functionName, err)
	}

	//match phone numbers
	if req.Phone != user.Phone {
		return utils.LogAndSendError(c, fiber.StatusUnprocessableEntity, functionName, err)
	}

	//generate a new otp
	otp := utils.GenerateOTP()

	//send otp
	// sent := utils.SendMail(req.Email, otp)
	// if !sent {
	// 	return utils.LogAndSendError(c, fiber.StatusServiceUnavailable, functionName, nil)
	// }

	//set verification code to redis key for 10 minutes
	emailVerificationCode := fmt.Sprintf("verification-code-email-%v", req.Email)
	if err := h.Redis.Set(c.Context(), emailVerificationCode, otp, time.Duration(10*time.Minute)).Err(); err != nil {
		return utils.LogAndSendError(c, fiber.StatusInternalServerError, functionName, err)
	}
	return utils.SendResponse(c, fiber.StatusOK, "New one time code is sent.")
}

//ResetPassword updates the user's password
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	functionName := "ResetPassword"

	var req request.ResetPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.LogAndSendError(c, fiber.StatusBadRequest, functionName, err)
	}

	//validate
	errors := utils.ValidateStruct(req)
	if errors != nil {
		return utils.SendValidationError(c, errors)
	}

	//get the key
	emailVerificationCode := fmt.Sprintf("verification-code-email-%v", req.Email)
	code, err := h.Redis.Get(c.Context(), emailVerificationCode).Result()
	if err != nil {
		return utils.LogAndSendError(c, fiber.StatusUnauthorized, functionName, err)
	}

	//check if redis code is matched with req.Code
	if code != req.Code {
		return utils.LogAndSendError(c, fiber.StatusUnauthorized, functionName, err)
	}

	//fetch user
	user, err := models.Users(models.UserWhere.Email.EQ(req.Email)).One(c.Context(), h.DB)
	if err != nil {
		return utils.LogAndSendError(c, fiber.StatusInternalServerError, functionName, err)
	}

	//hash new password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.LogAndSendError(c, fiber.StatusInternalServerError, functionName, err)
	}
	user.Password = string(hashedPassword)
	user.Update(c.Context(), h.DB, boil.Infer())

	//delete that key
	h.Redis.Del(c.Context(), fmt.Sprintf("verification-code-email-%v", req.Email))

	return utils.SendResponse(c, fiber.StatusOK, nil)
}

//LogoutUser ...
func (h *Handler) LogoutUser(c *fiber.Ctx) error {
	functionName := "LogoutUser"
	var req request.RefreshTokensRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.LogAndSendError(c, fiber.StatusBadRequest, functionName, err)
	}

	//validate
	errors := utils.ValidateStruct(req)
	if errors != nil {
		return utils.SendValidationError(c, errors)
	}

	claims, err := ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return utils.LogAndSendError(c, fiber.StatusUnauthorized, functionName, err)
	}

	//get redis key refresh-token-user-id-%s
	jti, err := h.Redis.Get(c.Context(), fmt.Sprintf("refresh-token-user-id-%s", claims.Subject)).Result()
	if err != nil {
		return utils.LogAndSendError(c, fiber.StatusUnauthorized, functionName, err)
	}

	//check if refresh_token from request body is matched with redis key
	if jti != claims.Jti {
		return utils.LogAndSendError(c, fiber.StatusUnauthorized, functionName, err)
	}

	h.Redis.Del(c.Context(), fmt.Sprintf("refresh-token-user-id-%s", claims.Subject))

	return utils.SendResponse(c, fiber.StatusNoContent, nil)
}
