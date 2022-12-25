package main

import (
	"log"
	"os"

	"github.com/Hariharan148/hustlie-email-api/api/handler/sendotp"
	"github.com/Hariharan148/hustlie-email-api/api/handler/verifyotp"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func setUpRoutes(app *fiber.App) {

	app.Post("/sendotp", sendotp.SendEmail )
	app.Post("/verify", verifyotp.VerifyOTP)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading environment variables ", err)
	}

	app := fiber.New()

	app.Use(logger.New())

	setUpRoutes(app)
	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}
