package main

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/andrew-lawlor/librepub/config"
	controllers "github.com/andrew-lawlor/librepub/controller"
	"github.com/andrew-lawlor/librepub/handlers"
	"github.com/andrew-lawlor/librepub/models"
	"github.com/gorilla/mux"
)

func main() {
	// Instantiate router from gorilla web libs.
	r := mux.NewRouter()
	handlers.ParseTemplates()
	r.HandleFunc("/", home)
	r.HandleFunc("/my-words", myWords)
	r.HandleFunc("/legal", legal)
	// Upload Kobo DB routes.
	r.HandleFunc("/upload-db", handlers.AuthMiddleware(handlers.UploadDB)).Methods("GET")
	r.HandleFunc("/upload-db", handlers.AuthMiddleware(handlers.UploadDBPost)).Methods("POST")
	// Books/Library routes.
	r.HandleFunc("/library", handlers.AuthMiddleware(handlers.GetBooks)).Methods("GET")
	r.HandleFunc("/library/{id}", handlers.AuthMiddleware(handlers.GetBook)).Methods("GET")
	// Annotation export routes.
	r.HandleFunc("/library/{id}/export", handlers.AuthMiddleware(handlers.ExportBook)).Methods("GET")
	r.HandleFunc("/export-all", handlers.AuthMiddleware(handlers.ExportBooks)).Methods("GET")
	// Vocab routes.
	r.HandleFunc("/vocab", handlers.AuthMiddleware(handlers.Vocab)).Methods("GET")
	r.HandleFunc("/vocab/{id}/delete", handlers.AuthMiddleware(handlers.DeleteVocab)).Methods("DELETE")
	r.HandleFunc("/export-vocab", handlers.AuthMiddleware(handlers.ExportVocab)).Methods("GET")
	// Auth: Register.
	r.HandleFunc("/register", handlers.RegistrationMiddleware(handlers.RegisterGet)).Methods("GET")
	r.HandleFunc("/register-post", handlers.CaptchaMiddleware(handlers.RegisterPost)).Methods("POST")
	// Auth: Login.
	r.HandleFunc("/login", handlers.LoginGet).Methods("GET")
	r.HandleFunc("/login-post", handlers.LoginPost).Methods("POST")
	// Auth: Logout.
	r.HandleFunc("/logout", handlers.Logout).Methods("GET")
	// Set up static file serving.
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	// Cron job to clean up export files.
	go doEvery(1*time.Minute, cleanExports)
	// Start server.
	http.ListenAndServe(":3002", r)
}

// Define function for route.
func home(w http.ResponseWriter, r *http.Request) {
	var pageData = controllers.NewPageData(r, config.HomeTitle, config.HomeDesc, "")
	handlers.RenderTemplate(r, w, "index.html", pageData)
}

func myWords(w http.ResponseWriter, r *http.Request) {
	var pageData = controllers.NewPageData(r, "My Words", "Learn about Kobo's My Words feature.", "")
	handlers.RenderTemplate(r, w, "my_words.html", pageData)
}

func legal(w http.ResponseWriter, r *http.Request) {
	var pageData = controllers.NewPageData(r, "Legal Stuff", "Legal stuff, just in case.", "")
	handlers.RenderTemplate(r, w, "legal.html", pageData)
}

// Run function every tick, based on passed in duration.
func doEvery(d time.Duration, f func()) {
	for range time.Tick(d) {
		f()
	}
}

func cleanExports() {
	duration := 1 * time.Minute
	exports, err := getOldFiles(config.ExportDir, duration)
	if err != nil {
		models.WriteLog(models.LogError, err.Error())
		return
	}
	for _, file := range exports {
		os.Remove(file)
	}
	uploads, err := getOldFiles(config.UploadDir, duration)
	if err != nil {
		models.WriteLog(models.LogError, err.Error())
		return
	}
	for _, file := range uploads {
		os.Remove(file)
	}
}

func getOldFiles(dir string, duration time.Duration) ([]string, error) {
	var oldFiles []string
	// Get the current time
	now := time.Now()
	// Walk through the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		// Check the modification time
		if now.Sub(info.ModTime()) > duration {
			oldFiles = append(oldFiles, path)
		}
		return nil
	})
	return oldFiles, err
}
