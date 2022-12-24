package emailApi

import (
	"os"
	"github.com/joho/godotenv"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
)


func Client()(*mailjet.Client){
	godotenv.Load(".env")

	mailjetClient := mailjet.NewMailjetClient(os.Getenv("MJ_APIKEY_PUBLIC"), os.Getenv("MJ_APIKEY_PRIVATE"))
	
	return mailjetClient
}