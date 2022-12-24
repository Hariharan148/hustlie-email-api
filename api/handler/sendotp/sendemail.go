package sendotp

import (
	"github.com/Hariharan148/hustlie-email-api/api/config/emailApi"
	"github.com/Hariharan148/hustlie-email-api/api/config/db"
	"github.com/Hariharan148/hustlie-email-api/api/config/helpers"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
	"github.com/go-redis/redis/v8"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"net"
	// "strings"
	"strconv"
	"os"
	"time"

)

var ips net.IP
var ip string



// func getIP(r *http.Request) (string, error) {

//     //Get IP from the X-REAL-IP header
	
//     ip = r.Header.Get("X-REAL-IP")
	
//     netIP := net.ParseIP(ip)
//     if netIP != nil {
// 		fmt.Println("1",ip)
//         return ip, nil
//     }

//     //Get IP from X-FORWARDED-FOR header
//     ips := r.Header.Get("X-FORWARDED-FOR")
	
//     splitIps := strings.Split(ips, ",")
//     for _, ip := range splitIps {
//         netIP := net.ParseIP(ip)
//         if netIP != nil {
// 			fmt.Println("2",ips)
//             return ip, nil
//         }
//     }

//     //Get IP from RemoteAddr
	
//     ip, _, err := net.SplitHostPort(r.RemoteAddr)
//     if err != nil {
//         return "", err
//     }
//     netIP = net.ParseIP(ip)
//     if netIP != nil {
// 		fmt.Println(r.RemoteAddr)
//         return ip, nil
//     }
//     return "", fmt.Errorf("No valid ip found")
// }

var r1 = db.RedisClient(1)



func rateLimiter(r *http.Request)(string, error){

	defer r1.Close() 


	ips, _ = helpers.LocalIP()
	fmt.Println(ips)


	val, err := r1.Get(db.Ctx, ip).Result()
	if err != nil && err != redis.Nil{
		log.Printf("Error cant connect to db: ", err)
		return "", err

	} else if err == redis.Nil{
		err = r1.Set(db.Ctx, ip, os.Getenv("API_LIMIT"), 24 *3600 *time.Second).Err()
		if err != nil{
			log.Printf("Error cant connect to db: ", err)
			return "", err
		}
	} else {
		val, err = r1.Get(db.Ctx, ip).Result()
		valInt, err := strconv.Atoi(val)
		if err != nil {
			log.Printf("Error while converting to int: ", err)
			return "", err
		}
		if valInt <= 0 {
			return "true", nil
		}
		fmt.Println("lmit", valInt )
	
		
	}
	
	return "", nil
}


func SendEmail(name string, email string, r *http.Request)( *mailjet.ResultsV31, string, error){

	limit, err:= rateLimiter(r)
	if err != nil {
		return &mailjet.ResultsV31{}, "Error while rate limiting", err
	}

	if limit != "" {
		return &mailjet.ResultsV31{}, "Too many login attempts! You are restricted for next 24hrs", nil
	}

	client := emailApi.Client()

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

	r1.Decr(db.Ctx, ip)

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




