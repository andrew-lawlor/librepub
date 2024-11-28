package handlers

import (
	"net/http"
	"strconv"

	"github.com/andrew-lawlor/librepub/auth"
	"github.com/andrew-lawlor/librepub/config"
	controllers "github.com/andrew-lawlor/librepub/controller"
	"github.com/andrew-lawlor/librepub/models"
)

func RegisterGet(w http.ResponseWriter, r *http.Request) {
	if auth.IsLoggedIn(r) {
		http.Redirect(w, r, "/index.html", http.StatusSeeOther)
		return
	}
	var pageData = controllers.NewPageData(r, config.RegisterTitle, "Registration page", "")
	pageData.Captcha, _ = models.GetRandomQuestion()
	token, err := auth.SetCSRFToken(r, w)
	if err != nil {
		RenderTemplate(r, w, "index.html", pageData)
		return
	}
	pageData.CSRFToken = token
	var templateToRender = "register.html"
	RenderTemplate(r, w, templateToRender, pageData)
}

func RegisterPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var csrfToken = r.FormValue(config.CSRFFieldName)
	if !auth.CheckCSRFToken(r, csrfToken) {
		auth.AddErrorMessage(r, w, "Invalid form token. Fool of a Took.")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	fields := []string{"userName", "password", "displayName"}
	formData := GetFormFieldValues(r, fields)
	result, errorMsgs := controllers.ValidateUserRegistration(formData)
	// Form validation failed. Redirect.
	if !result {
		for _, val := range errorMsgs {
			auth.AddErrorMessage(r, w, val)
		}
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	var res = models.NewUser(formData["userName"], formData["displayName"], auth.HashPW(formData["password"]))
	if res {
		auth.AddSuccessMessage(r, w, "Account registered successfully!")
		// Registration successful - send to login page.
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	} else {
		auth.AddErrorMessage(r, w, "Account registration failed!")
		http.Redirect(w, r, "/register", http.StatusInternalServerError)
		return
	}
}

func LoginGet(w http.ResponseWriter, r *http.Request) {
	if auth.IsLoggedIn(r) {
		http.Redirect(w, r, "/index.html", http.StatusSeeOther)
		return
	}
	var pageData = controllers.NewPageData(r, config.LoginTitle, "Login page", "")
	token, err := auth.SetCSRFToken(r, w)
	if err != nil {
		RenderTemplate(r, w, "index.html", pageData)
		return
	}
	pageData.CSRFToken = token
	var templateToRender = "login.html"
	RenderTemplate(r, w, templateToRender, pageData)
}

func LoginPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var csrfToken = r.FormValue(config.CSRFFieldName)
	if !auth.CheckCSRFToken(r, csrfToken) {
		auth.AddErrorMessage(r, w, "Invalid form token. Fool of a Took.")
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
		return
	}
	fields := []string{"userName", "password"}
	formData := GetFormFieldValues(r, fields)
	var userName = formData["userName"]
	var password = formData["password"]
	// Failed authentication.
	if !auth.Login(r, w, userName, password) {
		// Do error stuff.
		auth.AddErrorMessage(r, w, "Invalid username or password.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// Success!
	auth.AddSuccessMessage(r, w, "Welcome back, "+auth.GetLoggedInUserDisplayName(r)+"!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	auth.Logout(r, w)
	auth.AddSuccessMessage(r, w, "Goodbye, for now!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Profile(w http.ResponseWriter, r *http.Request) {
	var pageData = controllers.NewPageData(r, config.ProfileTitle, config.ProfileDesc, "")
	RenderTemplate(r, w, "profile.html", pageData)
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Passed auth.
		if auth.IsLoggedIn(r) {
			next.ServeHTTP(w, r)
			return
		}
		// Failed auth.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

func RegistrationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config, err := models.GetConfig(1)
		// Property doesn't exist. Panic and redirect.
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		// Registration disabled.
		if config.Value == "0" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		// Registration allowed.
		next.ServeHTTP(w, r)
	})
}

func CaptchaMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Process Captcha test.
		var humanCheck = r.FormValue("humanCheck")
		var humanCheckID = r.FormValue("humanCheckID")
		questionID, err := strconv.Atoi(humanCheckID)
		if err != nil {
			http.Error(w, "Bad request :)", http.StatusBadRequest)
			return
		}
		question, err := models.GetQuestion(questionID)
		if err != nil {
			http.Error(w, "Bad request :)", http.StatusBadRequest)
			return
		}
		if question.Answer != humanCheck {
			http.Error(w, "Incorrect humanity check answer.", http.StatusBadRequest)
			return
		}
		// Registration allowed.
		next.ServeHTTP(w, r)
	})
}
