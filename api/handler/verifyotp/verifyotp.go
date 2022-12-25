package verifyotp

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/Hariharan148/hustlie-email-api/api/config/db"
)

type Request struct {
	Email string `json:"email"`
	Otp  string `json:"otp"`
}

type Response struct {
	Found  bool `json:"found"`
	Error  string `json:"error"`
}

func VerifyOTP(c *fiber.Ctx)error{

	var body = new(Request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse the body"})
	}

	r := db.RedisClient(0)
	defer r.Close()

	val, err := r.Get(db.Ctx, body.Email).Result()
	if err == redis.Nil{
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"found": false})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"found":false,
			"error":"cant connect to database",
		})
	}

	// verify

	var found bool

	if body.Otp == val{
		found = true
	}

	resp := Response{
		Found: found,
		Error: "",
	}


	return c.Status(fiber.StatusOK).JSON(resp)
}




