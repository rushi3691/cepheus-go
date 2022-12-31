package googleauth

import (
	"cephgo/database"
	"context"
	"os"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/idtoken"
)

type ReqToken struct {
	Token string `json:"idToken"`
}

var CLIENT_ID = os.Getenv("CLIENT_ID")

func GoogleAuthMiddleware(c *fiber.Ctx) error {
	token := new(ReqToken)
	if err := c.BodyParser(&token); err != nil {
		return c.JSON(fiber.Map{
			"message": "pass token",
		})
	}
	{
		payload, err := idtoken.Validate(context.Background(), token.Token, CLIENT_ID)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"message": "invalid ID Token",
			})
		}
		user := new(database.User)
		user.UserUuid = payload.Claims["sub"].(string)
		user.UserName = payload.Claims["name"].(string)
		user.Email = payload.Claims["email"].(string)
		c.Locals("user", user)
	}
	return c.Next()
}
