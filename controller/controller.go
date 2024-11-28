package controllers

import (
	"net/http"
	"regexp"
	"unicode/utf8"

	"github.com/andrew-lawlor/librepub/auth"
	"github.com/andrew-lawlor/librepub/config"
	"github.com/andrew-lawlor/librepub/models"
)

type PageData[T any] struct {
	PageTitle           string
	SuccessMsg          string
	ErrorMsgs           []string
	PageDescription     string
	PageNumber          int
	SearchParm          string
	LoggedIn            bool
	UserName            string
	UserDisplayName     string
	RegistrationEnabled bool
	CSRFToken           string
	CSRFFieldName       string
	Captcha             models.Question
	Data                T
}

func NewPageData[T any](r *http.Request, pageTitle string, pageDescription string, data T) PageData[T] {
	var pageData = PageData[T]{
		PageTitle:           pageTitle,
		SuccessMsg:          "",
		ErrorMsgs:           make([]string, 0),
		PageDescription:     pageDescription,
		LoggedIn:            auth.IsLoggedIn(r),
		UserName:            auth.GetLoggedInUserName(r),
		UserDisplayName:     auth.GetLoggedInUserDisplayName(r),
		RegistrationEnabled: auth.IsRegistrationEnabled(),
		CSRFToken:           "",
		CSRFFieldName:       config.CSRFFieldName,
		Data:                data,
	}
	return pageData
}

func ValidateUserRegistration(formData map[string]string) (bool, []string) {
	var errorMsgs = make([]string, 0)
	var result bool = true
	if !ValidateStringLength(formData["userName"], 1, 24) {
		result = false
		errorMsgs = append(errorMsgs, "Username must be between 1 and 24 characters in length.")
	}
	if !ValidateStringLength(formData["password"], 8, 32) {
		result = false
		errorMsgs = append(errorMsgs, "Password must be between 8 and 32 characters in length.")
	}
	if !ValidateStringLength(formData["displayName"], 1, 24) {
		result = false
		errorMsgs = append(errorMsgs, "Display Name must be between 1 and 24 characters in length.")
	}
	if !ValidateUserName(formData["userName"]) {
		result = false
		errorMsgs = append(errorMsgs, "Username can only contain alphanumeric characters, and the following special characters: _ - ")
	}
	if !ValidatePassword(formData["password"]) {
		result = false
		errorMsgs = append(errorMsgs, "Password can only contain alphanumeric characters, and the following special characters: _ - , .")
	}
	if !ValidateDisplayName(formData["displayName"]) {
		result = false
		errorMsgs = append(errorMsgs, "Display Name can only contain alphanumeric characters, single spaces, and the following special characters: _ -")
	}
	return result, errorMsgs
}

func ValidateUserName(str string) bool {
	r, _ := regexp.Compile("^[a-zA-Z0-9_-]*$")
	return r.MatchString(str)
}

func ValidatePassword(str string) bool {
	r, _ := regexp.Compile("^[a-zA-Z0-9.,-]*$")
	return r.MatchString(str)
}

func ValidateDisplayName(str string) bool {
	r, _ := regexp.Compile("^[a-zA-Z0-9_ -]*$")
	return r.MatchString(str)
}

func ValidateStringLength(inputStr string, minLength int, maxLength int) bool {
	if utf8.RuneCountInString(inputStr) < minLength || utf8.RuneCountInString(inputStr) > maxLength {
		return false
	} else {
		return true
	}
}
