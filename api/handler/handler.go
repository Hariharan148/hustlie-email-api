package handler

import (
	"github.com/Hariharan148/hustlie-email-api/api/config"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
	"crypto/rand"
	"fmt"
	"log"

)


func SendEmail(name string, email string)( *mailjet.ResultsV31, string, error){

	client := config.Client()

	otp := otpGenerator()

	formatedMsg := fmt.Sprintf("<h3>Here is your OTP - " + otp + "</h3><br />Dear " + name +", Welcome to Huslie!")

	messagesInfo := []mailjet.InfoMessagesV31 {
      mailjet.InfoMessagesV31{
        From: &mailjet.RecipientV31{
          Email: "harirevorhustle@gmail.com",
          Name: "Hustlie",
        },
        To: &mailjet.RecipientsV31{
          mailjet.RecipientV31 {
            Email: email,
            Name: name,
          },
        }, 
        Subject: "OTP - Team Hustlie",
        TextPart: "Here is your OTP - " + otp,
        HTMLPart:  formatedMsg,
      },
	}


	messages := mailjet.MessagesV31{Info: messagesInfo }
	res, err := client.SendMailV31(&messages)
	if err != nil {
		return res, otp, err
	}

	return res, otp, nil
}



func otpGenerator() string {

	otp, err := rand.Prime(rand.Reader, 18)

	if err != nil {
		log.Printf("Error occured during generating otp", err)
	}
	strOtp := fmt.Sprintf("%v", otp)
	return strOtp

}




