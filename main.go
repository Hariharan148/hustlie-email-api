package main

import (
	"net/http"
	"fmt"
	"encoding/json"
	"log"
	"github.com/Hariharan148/hustlie-email-api/api/handler/sendotp"
	"github.com/tidwall/gjson"
)

type Request struct {
	Email string 
	Name  string 
}

type Response struct{
	MailStatus string `json:"mail_status"`
	DbStatus string `json:"db_status"`
}


func parseBody(r *http.Request)Request{
	var body Request
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Printf("error %v",err)
	}
	return body
}


func response(w http.ResponseWriter, r *http.Request){
	if r.URL.Path != "/otp" {
		http.Error(w, "404 not found", http.StatusNotFound)
	}

	if r.Method != "POST" {
		http.Error(w, "Method not supported", http.StatusNotFound)
	}

	body := parseBody(r)


	res, otp, err := sendotp.SendEmail(body.Name, body.Email, r)
	if err != nil {
		http.Error(w, "Error while sending mail", http.StatusInternalServerError)
	}
	
	data, err:= json.MarshalIndent(res, "", "    ")
	if err != nil {
		fmt.Printf("Error while marshaling payload: ", err)
	}

	strData := string(data[:])

	status := gjson.Get(strData, "Messages.0.Status")


	response := Response{
		MailStatus: status.String(),
		DbStatus: otp,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)

}

func main(){
	http.HandleFunc("/sendotp", response)
	fmt.Println("Server starting at port 8000...")
	log.Fatal(http.ListenAndServe(":8000", nil ))
}