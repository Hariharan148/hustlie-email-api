package helpers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Hariharan148/hustlie-email-api/api/config/emailApi"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
	// "github.com/tidwall/gjson"
	"fmt"
	"log"
	// "encoding/json"
)

type Request struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}


func SendEmail(c *fiber.Ctx) (error) {


	// RATE LIMITING 

	rd1 := db.RedisClient(1)
	defer rd1.Close()


	fmt.Println("entered rl")
	val, err := rd1.Get(db.Ctx, c.IP()).Result()

	if err == redis.Nil {
		_ = rd1.Set(db.Ctx, c.IP(), os.Getenv("API_LIMIT"), 24*3600*time.Second).Err()

	} else {
		val, _ = rd1.Get(database.Ctx, c.IP()).Result()
		valInt, _ := strconv.Atoi(val)

		if valInt <= 0 {
			limit, _ := rd1.TTL(database.Ctx, c.IP()).Result()

			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":              "Too many login attempts! You are restricted for next 24hrs",
				"x_rate_limit_reset": limit / time.Minute / 60,
			})
		}
	}
	fmt.Println("rate limited")


	// GENERATE OTP


	// Sending email to mailjet


	var body = new(Request)

	if err := c.BodyParser(&body); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse the body"})
		return err
	}

	otp := "123456"

	client := emailApi.Client()
	fmt.Println("sendmail")

	formatedMsg := fmt.Sprintf("<h3>Here is your OTP - " + otp + "</h3><br />Dear " + body.Name + ", Welcome to Huslie!")

	messagesInfo := []mailjet.InfoMessagesV31{
		mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: "harirevorhustle@gmail.com",
				Name:  "Hustlie",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: body.Email,
					Name:  body.Name,
				},
			},
			Subject:  "OTP - Team Hustlie",
			TextPart: "Here is your OTP - " + otp,
			HTMLPart: formatedMsg,
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}
	res, err := client.SendMailV31(&messages)
	if err != nil {
		log.Println("error while sending the email!")
		return err
	}


	//DECREMENT THE LIMIT 
	
	rd1.Decr(db.Ctx, c.IP())
	

	//SAVE IN REDIS



	
	return c.Status(fiber.StatusOK).JSON(res)

}