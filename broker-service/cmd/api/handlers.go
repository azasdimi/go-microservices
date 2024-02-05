package main

import (
	"broker/event"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {

	payload := jsonResponse{
		Error:   false,
		Message: "Hit the Broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)

	// out, _ := json.MarshalIndent(payload, "", "\t")
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader((http.StatusAccepted))
	// w.Write(out)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.Authenticate(w, requestPayload.Auth)
	case "log":
		// app.LogItem(w, requestPayload.Log)
		// app.LogEventViaRabbit(w, requestPayload.Log)
		app.LogItemViaRPC(w, requestPayload.Log)
	case "mail":
		app.SendMail(w, requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("Unknown Action"))
	}
}

func (app *Config) Authenticate(w http.ResponseWriter, a AuthPayload) {
	//Create json to send to the auth microservice

	jsonData, _ := json.MarshalIndent(a, "", "\t")

	//Call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	//Make sure we get the correct status code

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("Invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("Error calling auth service"))
		return
	}

	//Create variable we'll read from
	var jsonFromService jsonResponse

	//Decode json from the authentication service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) LogItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceUrl := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(w, err)
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, err)
	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) SendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, _ := json.MarshalIndent(msg, "", "\t")

	//call the mail service
	mailServiceUrl := "http://mailer-service/send"

	//post to mail service
	request, err := http.NewRequest("POST", mailServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
	}

	defer response.Body.Close()

	//make sure we get back the right status code

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling mail service"))
	}

	//Send back payload
	var payload jsonResponse

	payload.Error = false
	payload.Message = "message sent to " + msg.To

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) LogEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.PushToQueue(l.Name, l.Data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = "logged via RabbitMQ"

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) PushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")

	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) LogItemViaRPC(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	rpcPayload := RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}

	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: result,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}
