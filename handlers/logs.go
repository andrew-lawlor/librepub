package handlers

import (
	"net/http"
	"strconv"

	"github.com/andrew-lawlor/librepub/auth"
	"github.com/andrew-lawlor/librepub/config"
	controllers "github.com/andrew-lawlor/librepub/controller"
	"github.com/andrew-lawlor/librepub/models"
	"github.com/gorilla/mux"
)

func Logs(w http.ResponseWriter, r *http.Request) {
	var page = r.URL.Query().Get("page")
	var search = r.URL.Query().Get("search")
	var partial = r.URL.Query().Get("partial")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 0
	}
	pageInt++
	var logs = models.GetLogs(pageInt, search)
	var pageData = controllers.NewPageData(r, config.LogsTitle, config.LogsDesc, logs)
	pageData.SearchParm = search
	// If no more data, let client know and get out.
	if len(logs) <= 0 {
		pageData.SuccessMsg = "No logs found."
		// Just show feedback to client if this is a partial operation.
		if partial == "1" {
			RenderTemplate(r, w, "part_validation_messages.html", pageData)
		} else { // Load full page with feedback if this is a full page load.
			RenderTemplate(r, w, "logs.html", pageData)
		}
		return
	}
	// Set page for pagination.
	pageData.PageNumber = pageInt
	// Render entire page or partial template, depending on partial query parm.
	var templateName = "logs.html"
	if partial == "1" {
		templateName = "part_logs.html"
	}
	RenderTemplate(r, w, templateName, pageData)
}

func DeleteLog(w http.ResponseWriter, r *http.Request) {
	var id = mux.Vars(r)["id"]
	idInt, err := strconv.Atoi(id)
	if err != nil {
		auth.AddErrorMessage(r, w, "Invalid ID.")
		http.Redirect(w, r, "/logs", http.StatusSeeOther)
		return
	}
	models.DeleteLog(idInt)
}

func DeleteAllLogs(w http.ResponseWriter, r *http.Request) {
	var msg string
	if models.DeleteAllLogs() {
		msg = "Logs cleared."
	} else {
		msg = "Database error. Check "
		models.WriteLog(models.LogError, msg)
		return
	}
	var pageData = controllers.NewPageData(r, config.LogsTitle, config.LogsDesc, "")
	pageData.PageTitle = "Logs Cleared"
	pageData.SuccessMsg = "Logs deleted successfully!"
	RenderTemplate(r, w, "part_dialog_success.html", pageData)
}
