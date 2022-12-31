package routes

import (
	"cephgo/android"
	"cephgo/database"

	"cephgo/googleauth"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	Uid        int64
	Ini        *string
	Registered bool
	jwt.RegisteredClaims
}

func IsRegistered(c *fiber.Ctx) error {
	user := c.Locals("claims").(*database.User)
	if user.Registered {
		return c.Next()
	}
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"message": "user is not registered",
	})
}

func JwtMiddleware(c *fiber.Ctx) error {

	claims := &Claims{}

	_, err := jwt.ParseWithClaims(c.Cookies("SSIDCP"), claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "invalid token",
		})
	}
	c.Locals("claims", &database.User{
		Id:         claims.Uid,
		Ini:        claims.Ini,
		Registered: claims.Registered,
	})
	return c.Next()
}

func SetupRoutes(app *fiber.App) {

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "hello"})
	})

	and := app.Group("/apiM1")

	and.Post("/login", googleauth.GoogleAuthMiddleware, android.CreateUserController)
	and.Post("/register", JwtMiddleware, android.RegisterUserController)
	and.Post("/createteam", JwtMiddleware, IsRegistered, android.CreateTeamController)
	and.Post("/regevent", JwtMiddleware, IsRegistered, android.RegEventController)
	and.Delete("/remevent", JwtMiddleware, IsRegistered, android.RemEventController)

}
