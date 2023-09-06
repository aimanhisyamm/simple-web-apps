package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
)

const (
	ADDRESS = "localhost:8080"

	// regex for simple email validation (without domain restrict)
	// and regex for indonesian phone number
	EMAIL_REGEX        = `^[A-Za-z0-9+_.-]+@[a-zA-Z0-9.-]+$`
	PHONE_NUMBER_REGEX = `^(\+62|62)?[\s-]?0?8[1-9]{1}\d{1}[\s-]?\d{4}[\s-]?\d{2,5}$`

	// flagging for validate sensitive information
	FLAG_EMAIL        = "EMAIL"
	FLAG_PHONE_NUMBER = "PHONE_NUMBER"

	// status http message
	STATUS_OK                 = "OK"
	STATUS_BAD_REQUEST        = "Bad Request"
	STATUS_METHOD_NOT_ALLOWED = "Method Not Allowed"

	// activity log message
	INFO_USER_ACTIVITY_LOG_MESSAGE     = "user %s take action %s"
	INFO_HIGHEST_ACTIVITY_USER_MESSAGE = "user %s with total records %d action"
	WARNING_EMPTY_ACTION_MESSAGE       = "user %s doesn't take any action"
	WARNING_EMPTY_USERNAME_MESSAGE     = "empty username"
)

var (
	// variable for custom logger
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger

	// variable to count total logger each user
	loggerCount = map[string]int{}

	// variable to prevent race condition on highest activity API and loggerCount
	mutex = &sync.Mutex{}
)

// initiate custom logger and save it to file logs.txt
func init() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// struct definition for HTTP request and response body
type UserActivity struct {
	Username string `json:"username"`
	Action   string `json:"action,omitempty"`
}

type UserInfo struct {
	Name           string `json:"name"`
	PersonalEmail  string `json:"personal_email"`
	PersonalNumber string `json:"personal_number"`
	OfficeEmail    string `json:"office_email"`
	OfficeNumber   string `json:"office_Number"`
}

type Response struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Errors interface{} `json:"errors,omitempty"`
}

// hompage endpoint
func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	InfoLogger.Println("Endpoint Hit: homepage")
	json.NewEncoder(w).Encode(Response{
		Code:   http.StatusOK,
		Status: STATUS_OK,
		Data:   "Welcome to Aiman's simple web application :)",
	})
}

// handler function/endpoint for check user with highest activity based on its log activity
func highestActivity(w http.ResponseWriter, r *http.Request) {
	InfoLogger.Println("Endpoint Hit: highestActivity")
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "POST":
		var body []UserActivity
		var errors []string

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errors = append(errors, err.Error())
			ErrorLogger.Println(err.Error())
			json.NewEncoder(w).Encode(Response{
				Code:   http.StatusBadRequest,
				Status: STATUS_BAD_REQUEST,
				Errors: errors,
			})
			return
		}

		mutex.Lock()
		// increament and record activity log
		increamentUserActivityLog(body)

		username, totalActivity := getHighestActivityUser(loggerCount)
		result := fmt.Sprintf(INFO_HIGHEST_ACTIVITY_USER_MESSAGE, username, totalActivity)
		InfoLogger.Println(result)
		json.NewEncoder(w).Encode(Response{
			Code:   http.StatusOK,
			Status: STATUS_OK,
			Data:   result,
		})
		mutex.Unlock()
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		WarningLogger.Println(STATUS_METHOD_NOT_ALLOWED)
		json.NewEncoder(w).Encode(Response{
			Code:   http.StatusMethodNotAllowed,
			Status: STATUS_METHOD_NOT_ALLOWED,
		})
	}
}

// handler function/endpoint for censored any valid sensitive information from user
func userInfo(w http.ResponseWriter, r *http.Request) {
	InfoLogger.Println("Endpoint Hit: userInfo")
	w.Header().Set("Content-Type", "application/json")

	// to prevent different HTTP method
	switch r.Method {
	case "POST":
		var (
			body   UserInfo
			errors []string
		)

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errors = append(errors, err.Error())
			ErrorLogger.Println(err.Error())
			json.NewEncoder(w).Encode(Response{
				Code:   http.StatusBadRequest,
				Status: STATUS_BAD_REQUEST,
				Errors: errors,
			})
			return
		}

		result := UserInfo{
			Name:           body.Name,
			PersonalEmail:  censoredSensitiveInfo(body.PersonalEmail, FLAG_EMAIL),
			OfficeEmail:    censoredSensitiveInfo(body.OfficeEmail, FLAG_EMAIL),
			PersonalNumber: censoredSensitiveInfo(body.PersonalNumber, FLAG_PHONE_NUMBER),
			OfficeNumber:   censoredSensitiveInfo(body.OfficeNumber, FLAG_PHONE_NUMBER),
		}

		json.NewEncoder(w).Encode(Response{
			Code:   http.StatusOK,
			Status: STATUS_OK,
			Data:   &result,
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		WarningLogger.Println(STATUS_METHOD_NOT_ALLOWED)
		json.NewEncoder(w).Encode(Response{
			Code:   http.StatusMethodNotAllowed,
			Status: STATUS_METHOD_NOT_ALLOWED,
		})
	}
}

// function for handle HTTP requests
func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/highest-activity", highestActivity)
	http.HandleFunc("/user-info", userInfo)

	log.Println("Server started at", ADDRESS)
	log.Fatal(http.ListenAndServe(ADDRESS, nil))
}

// ===== main function =====

func main() {
	handleRequests()
}

// ===== logic function ====

// function for increament loggerCount and record user activity
func increamentUserActivityLog(data []UserActivity) {
	for _, val := range data {
		if val.Username != "" && val.Action != "" {
			loggerCount[val.Username] += 1
			InfoLogger.Printf(INFO_USER_ACTIVITY_LOG_MESSAGE, val.Username, val.Action)
		} else if val.Username == "" {
			WarningLogger.Println(WARNING_EMPTY_USERNAME_MESSAGE)
		} else if val.Action == "" {
			WarningLogger.Printf(WARNING_EMPTY_ACTION_MESSAGE, val.Username)
		}
	}
}

// function for get highest username and total activity user from logger count
func getHighestActivityUser(logCounter map[string]int) (username string, totalActivity int) {
	var (
		highestTotalActivity int
		highestUsername      string
	)

	for key, _ := range logCounter {
		if logCounter[key] > highestTotalActivity {
			highestTotalActivity = logCounter[key]
			highestUsername = key
		}
	}
	return highestUsername, highestTotalActivity
}

// function for validate regex email and phone number
func isEmailValid(email string) bool {
	emailRegex := regexp.MustCompile(EMAIL_REGEX)
	return emailRegex.MatchString(email)
}

func isPhoneNumberValid(phoneNumber string) bool {
	phoneNumberRegex := regexp.MustCompile(PHONE_NUMBER_REGEX)
	return phoneNumberRegex.MatchString(phoneNumber)
}

// function for change any valid email and phone number data to be censored
func censoredSensitiveInfo(data string, flag string) string {
	var result string = data
	switch flag {
	case FLAG_EMAIL:
		if isEmailValid(data) {
			result = "*CENSORED*"
		}
	case FLAG_PHONE_NUMBER:
		if isPhoneNumberValid(data) {
			result = "*CENSORED*"
		}
	}
	return result
}
