package sendotp

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Hariharan148/Url-Shortener-Go-Redis/api/database"
	"github.com/Hariharan148/hustlie-email-api/api/config/db"
	"github.com/Hariharan148/hustlie-email-api/api/config/emailApi"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
	"github.com/tidwall/gjson"
)

type Request struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Response struct {
	MailStatus      string        `json:"mail_status"`
	DbStatus        string        `json:"db_status"`
	XRateLimit      int           `json:"x_rate_limit"`
	XRateLimitReset time.Duration `json:"x_rate_limit_reset"`
}

func rateLimiter(c *fiber.Ctx) error {
	rd1 := db.RedisClient(1)
	defer rd1.Close()

	val, err := rd1.Get(db.Ctx, c.IP()).Result()

	if err == redis.Nil {
		_ = rd1.Set(db.Ctx, c.IP(), os.Getenv("API_LIMIT"), 24*3600*time.Second).Err()
	} else {
		// val, _ = rd1.Get(database.Ctx, c.IP()).Result()
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := rd1.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":              "Too many login attempts! You are restricted for next 24hrs",
				"x_rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}
	return nil
}

func SendEmail(c *fiber.Ctx) error {
	r1 := db.RedisClient(1)
	defer r1.Close()

	var body *Request

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse the body"})
	}

	rateLimiter(c)

	client := emailApi.Client()

	otp := otpGenerator()

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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error while sending the email!",
		})
	}

	data, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		fmt.Printf("Error while marshaling payload: ", err)
	}

	strData := string(data[:])

	msgStatus := gjson.Get(strData, "Messages.0.Status")

	resp := Response{
		MailStatus:      msgStatus.String(),
		DbStatus:        "",
		XRateLimit:      2,
		XRateLimitReset: 24,
	}

	r1.Decr(db.Ctx, c.IP())

	limit, _ := r1.Get(database.Ctx, c.IP()).Result()
	resp.XRateLimit, _ = strconv.Atoi(limit)

	limitReset, _ := r1.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = limitReset / time.Nanosecond / time.Minute

	return c.Status(fiber.StatusOK).JSON(res)
}

func otpGenerator() string {

	otp, err := rand.Prime(rand.Reader, 18)

	if err != nil {
		log.Printf("Error occured during generating otp", err)
	}
	strOtp := fmt.Sprintf("%v", otp)
	return strOtp

}
