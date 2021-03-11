package middlewares

import (
	"go-rbac/source/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/o1egl/paseto"
)

//LoggedIn checks if user is logged in with given role or not
//if no role is given then all logged in users are allowed
func LoggedIn(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		//read authorization header from request
		authorizationHeader := c.Get("Authorization")

		//from cookies
		accessTokenFromCookies := c.Cookies("access_token")

		var claims paseto.JSONToken
		var err error

		//check if cookie present
		if accessTokenFromCookies == "" {
			//check length
			if len(authorizationHeader) < 7 {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "invalid header",
				})
			}

			//extract token
			tokenFromHeader := authorizationHeader[len("Bearer "):]
			claims, err = handlers.ValidateAccessToken(tokenFromHeader)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "invalid token",
				})
			}
		} else {
			//if cookie present
			claims, err = handlers.ValidateAccessToken(accessTokenFromCookies)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "invalid token",
				})
			}
		}

		//if roles length is greater than 0 then verify role
		if len(roles) > 0 {
			ok := checkRole(claims.Audience, roles)
			if !ok {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "unauthorized",
				})
			}
		}

		//set claims
		c.Request().Header.Set("user", claims.Subject)
		c.Request().Header.Set("role", claims.Audience)
		return c.Next()
	}
}

func checkRole(audience string, roles []string) bool {
	for _, role := range roles {
		if role == audience {
			return true
		}
	}
	return false
}
