package main

import (
	"net/http"
	"fmt"
	"encoding/json"
	"log"
	"github.com/Hariharan148/hustlie-email-api/api/handler"
	"github.com/tidwall/gjson"
)

type Request struct {
	Email string 
	Name  string 
}

type Response struct{
	Response string `json:"response"`
	OTP 	 string			`json:"otp"`
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


	res, otp, err := handler.SendEmail(body.Name, body.Email)
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
		Response: status.String(),
		OTP: otp,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)

}

func main(){
	http.HandleFunc("/otp", response)
	fmt.Println("Server starting at port 8000...")
	log.Fatal(http.ListenAndServe(":8000", nil ))
}