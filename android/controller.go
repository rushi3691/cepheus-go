package android

import (
	"cephgo/database"
	"cephgo/utils"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func _sendWithJwtHelper(user *database.User, c *fiber.Ctx) error {
	claims := jwt.MapClaims{
		"uid":        user.Id,
		"ini":        utils.GetInitials(user.UserName),
		"registered": user.Registered,
		"exp":        time.Now().Add(time.Hour * 72).Unix(),
		// "grade":      user.Grade,
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return c.JSON(fiber.Map{
		"token": t,
		"user":  user,
	})
}

func CreateUserController(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*database.User)

	{
		user, err := database.DB_STRUCT.GetUserByUUID(userClaims.UserUuid)
		if err == nil {
			return _sendWithJwtHelper(user, c)
		}
	}

	user, err := database.DB_STRUCT.CreateUser(userClaims)
	if err != nil {
		// log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return _sendWithJwtHelper(user, c)
}

func RegisterUserController(c *fiber.Ctx) error {
	user := c.Locals("claims").(*database.User)

	if err := c.BodyParser(&user); err != nil {
		// log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	user, err := database.DB_STRUCT.RegisterUser(user)
	user.Registered = true

	if err != nil {
		// log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return _sendWithJwtHelper(user, c)
}

func CreateTeamController(c *fiber.Ctx) error {
	userClaims := c.Locals("claims").(*database.User)
	team := new(database.Team)

	if err := c.BodyParser(&team); err != nil {
		// log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	team, err := database.DB_STRUCT.CreateTeam(team, userClaims.Ini)
	if err != nil {
		// log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(team)
}

func RegEventController(c *fiber.Ctx) error {
	user := c.Locals("claims").(*database.User)
	body := new(database.RegEventReq)
	if err := c.BodyParser(&body); err != nil {
		// log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	err := database.DB_STRUCT.RegisterEventUser(body, int(user.Id))
	if err != nil {
		// log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "registration done",
	})
}

type EventReq struct {
	EventId int `json:"event_id"`
}

func RemEventController(c *fiber.Ctx) error {
	user := c.Locals("claims").(*database.User)
	req := new(EventReq)
	if err := c.BodyParser(&req); err != nil {
		// log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	database.DB_STRUCT.RemoveEventUser(user.Id, req.EventId)
	return c.JSON(fiber.Map{
		"message": "removed",
	})
}

// func Logout(c *fiber.Ctx) error {
// 	return c.ClearCookie()
// }
