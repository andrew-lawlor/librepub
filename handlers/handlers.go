package handlers

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/andrew-lawlor/librepub/auth"
	controllers "github.com/andrew-lawlor/librepub/controller"
	"github.com/andrew-lawlor/librepub/models"
)

var templates *template.Template

func RenderTemplate[T any](r *http.Request, w http.ResponseWriter, templateName string, pageData controllers.PageData[T]) {
	// If none provided, get messages from session to display to user.
	if len(pageData.ErrorMsgs) <= 0 {
		pageData.ErrorMsgs = auth.GetErrorMessages(r, w)
	}
	if pageData.SuccessMsg == "" {
		pageData.SuccessMsg = auth.GetSuccessMessage(r, w)
	}
	// Render template, pass data, and write to writer.
	execErr := templates.ExecuteTemplate(w, templateName, pageData)
	if execErr != nil {
		models.WriteLog(models.LogError, execErr.Error())
	}
}

func ParseTemplates() {
	templ := template.New("")
	filepath.Walk("templates/", func(path string, info os.FileInfo, err error) error {
		// Parse each file that's not a dir.
		if !info.IsDir() {
			_, err := templ.ParseFiles(path)
			if err != nil {
				models.WriteLog(models.LogError, err.Error())
				return err
			}
		}
		// Not a file.
		return nil
	})
	templates = templ
}

func GetFormFieldValues(r *http.Request, fields []string) map[string]string {
	var formData = make(map[string]string)
	var i = 0
	for _, field := range fields {
		formData[field] = r.FormValue(field)
		i++
	}
	return formData
}
