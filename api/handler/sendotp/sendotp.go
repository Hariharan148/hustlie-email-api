
package sendotp

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"os"
	"crypto/rand"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/Hariharan148/hustlie-email-api/api/config/emailApi"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
	"github.com/Hariharan148/hustlie-email-api/api/config/db"
)



type Request struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}


type Response struct {
	MailStatus      *mailjet.ResultsV31   `json:"mail_status"`
	DbStatus        string        		  `json:"db_status"`
	XRateLimit      int           		  `json:"x_rate_limit"`
	XRateLimitReset time.Duration		  `json:"x_rate_limit_reset`	 
}


var worklist = make(chan string)
var valInt int
var limit time.Duration


func otpGenerator() {

	otp, err := rand.Prime(rand.Reader, 18)

	if err != nil {
		log.Printf("Error occured during generating otp", err)
	}
	strOtp := fmt.Sprintf("%v", otp)
	worklist <- strOtp

}



func SendEmail(c *fiber.Ctx) (error) {


	// RATE LIMITING 

	rd1 := db.RedisClient(1)
	defer rd1.Close()

	// rd1.Incr(db.Ctx, c.IP())


	fmt.Println("entered r")
	val, err := rd1.Get(db.Ctx, c.IP()).Result()

	if err == redis.Nil {
		_ = rd1.Set(db.Ctx, c.IP(), os.Getenv("API_LIMIT"), 10*time.Second).Err()

	} else if err == nil{
		val, _ = rd1.Get(db.Ctx, c.IP()).Result()
		valInt, _ = strconv.Atoi(val)

		fmt.Println("hi")
		fmt.Println(valInt)
		if valInt <= 0 {
			limit, _ = rd1.TTL(db.Ctx, c.IP()).Result()

			remTime := strconv.Itoa(int(limit/time.Minute/60))

			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":              "Too many login attempts! You are restricted for next " + remTime +"hrs",
				"x_rate_limit_reset": limit / time.Minute / 60  ,
			})
		}
	} else {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":"Cannot connect to the database :(" ,	})
	}
	fmt.Println("rate limited")


	// GENERATE OTP

	go otpGenerator()


	// Sending email to mailjet

	var body = new(Request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse the body"})
	}


	client := emailApi.Client()
	fmt.Println("sendmail")

	otp := <- worklist

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


	//DECREMENT THE LIMIT 
	
	rd1.Decr(db.Ctx, c.IP())
	

	//SAVE IN REDIS

	r := db.RedisClient(0)
	defer r.Close()

	value, err := r.Get(db.Ctx, body.Email).Result()
	fmt.Println(err)
	if err != nil && err != redis.Nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"cant connect to database"})
	} else {
		value = "success"
	}


	err = r.Set(db.Ctx, body.Email, otp, 30*60*time.Second).Err()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"cant connect to database"})
	} else {
		value = "success"
	}

	

	messages := mailjet.MessagesV31{Info: messagesInfo}
	res, err := client.SendMailV31(&messages)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot send the email"})
	}


	// SEND RESPONSE

	fmt.Println(value)
	resp := Response{
		MailStatus: res,
		DbStatus: value,
		XRateLimit: valInt,
		XRateLimitReset: limit / time.Minute / 60, 
	}

	
	return c.Status(fiber.StatusOK).JSON(resp)

}