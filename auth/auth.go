package auth

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/andrew-lawlor/librepub/models"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key        = []byte("")
	store      *sessions.CookieStore
	cookieName string
)

func init() {
	content, err := os.ReadFile("./config/sac.txt")
	if err != nil {
		log.Fatal(err.Error())
	}
	var lines = strings.Split(string(content), "\n")
	cookieName = lines[0]
	key = []byte(lines[1])
	store = sessions.NewCookieStore(key)
	// Set session options
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   1209600,              // 2 weeks
		HttpOnly: true,                 // Prevents JavaScript access
		SameSite: http.SameSiteLaxMode, // Lax is a good balance for most use cases
		Secure:   true,                 // Only serve cookie over HTTPS
	}
}

func IsLoggedIn(r *http.Request) bool {
	session, _ := store.Get(r, cookieName)
	loggedIn, ok := session.Values["authenticated"].(bool)
	if !ok || !loggedIn {
		return false
	}
	return true
}

func GetLoggedInUserID(r *http.Request) int {
	session, _ := store.Get(r, cookieName)
	userID, ok := session.Values["userID"].(int)
	if !ok {
		return -1
	}
	return userID
}

func GetLoggedInUserName(r *http.Request) string {
	session, _ := store.Get(r, cookieName)
	userName, ok := session.Values["userName"].(string)
	if !ok {
		return ""
	}
	return userName
}

func GetLoggedInUserDisplayName(r *http.Request) string {
	session, _ := store.Get(r, cookieName)
	displayName, ok := session.Values["displayName"].(string)
	if !ok {
		return ""
	}
	return displayName
}

func Login(r *http.Request, w http.ResponseWriter, userName string, password string) bool {
	var user, err = models.GetUser(userName)
	if err != nil {
		models.WriteLog(models.LogError, err.Error())
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		models.WriteLog(models.LogError, err.Error())
		return false
	}
	// Set user as authenticated
	session, err := store.Get(r, cookieName)
	if err != nil {
		models.WriteLog(models.LogError, err.Error())
		return false
	}

	// Set session key/vals.
	session.Values["authenticated"] = true
	session.Values["userName"] = user.UserName
	session.Values["displayName"] = user.DisplayName
	session.Values["userID"] = user.UserID
	session.Save(r, w)
	models.WriteLog(models.LogInfo, "User logged in: "+user.UserName)
	return true
}

func Logout(r *http.Request, w http.ResponseWriter) {
	// End users session.
	session, _ := store.Get(r, cookieName)
	session.Values["authenticated"] = false
	session.Values["userName"] = ""
	session.Values["displayName"] = ""
	session.Save(r, w)
}

func AddErrorMessage(r *http.Request, w http.ResponseWriter, msg string) {
	session, _ := store.Get(r, cookieName)
	errorMsgs, ok := session.Values["errorMsgs"].([]string)
	// If map isn't yet defined, make it!
	if !ok {
		errorMsgs = make([]string, 0)
	}
	errorMsgs = append(errorMsgs, msg)
	session.Values["errorMsgs"] = errorMsgs
	session.Save(r, w)
}

func GetErrorMessages(r *http.Request, w http.ResponseWriter) []string {
	session, _ := store.Get(r, cookieName)
	errorMsgs, ok := session.Values["errorMsgs"].([]string)
	// If map isn't yet defined, make it!
	if !ok {
		return make([]string, 1)
	}
	// Clear messages.
	session.Values["errorMsgs"] = make([]string, 0)
	session.Save(r, w)
	return errorMsgs
}

func AddSuccessMessage(r *http.Request, w http.ResponseWriter, msg string) {
	session, _ := store.Get(r, cookieName)
	session.Values["successMsg"] = msg
	session.Save(r, w)
}

func GetSuccessMessage(r *http.Request, w http.ResponseWriter) string {
	session, _ := store.Get(r, cookieName)
	successMsg, ok := session.Values["successMsg"].(string)
	if !ok {
		return ""
	}
	// Clear message after reading.
	session.Values["successMsg"] = ""
	session.Save(r, w)
	return successMsg
}

func HashPW(password string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(hashed)
}

func SetCSRFToken(r *http.Request, w http.ResponseWriter) (string, error) {
	token, err := generateRandomString(32)
	if err != nil {
		return "", err
	}
	session, _ := store.Get(r, cookieName)
	session.Values["csrfToken"] = token
	session.Save(r, w)
	return token, nil
}

func CheckCSRFToken(r *http.Request, formToken string) bool {
	session, _ := store.Get(r, cookieName)
	token, ok := session.Values["csrfToken"].(string)
	if !ok {
		return false
	}
	if token == formToken {
		return true
	} else {
		return false
	}
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely-generated random string.
func generateRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func IsRegistrationEnabled() bool {
	config, err := models.GetConfig(1)
	// Property doesn't exist.
	if err != nil {
		return false
	}
	// Registration disabled.
	if config.Value == "0" {
		return false
	}
	return true
}
