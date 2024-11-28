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

func SiteConfig(w http.ResponseWriter, r *http.Request) {
	var configs = models.GetConfigs()
	var pageData = controllers.NewPageData(r, config.ConfigTitle, "Configuration Page", configs)
	// Simple GET case.
	if r.Method == "GET" {
		RenderTemplate(r, w, "site_config.html", pageData)
		return
	}
}

func SiteConfigUpdate(w http.ResponseWriter, r *http.Request) {
	configID := mux.Vars(r)["id"]
	intID, err := strconv.Atoi(configID)
	if err != nil {
		auth.AddErrorMessage(r, w, "Invalid Config ID.")
		http.Redirect(w, r, "/site-config", http.StatusSeeOther)
		return
	}
	config, err := models.GetConfig(intID)
	if err != nil {
		auth.AddErrorMessage(r, w, "Config not found.")
		http.Redirect(w, r, "/site-config", http.StatusSeeOther)
		return
	}
	// Getting and validating form fields.
	fields := []string{"value"}
	var fieldValues = GetFormFieldValues(r, fields)
	// TODO: add validation based on type. Put in Controllers.
	config.Value = fieldValues["value"]
	var result = models.EditConfig(config)
	// Database error.
	if !result {
		auth.AddErrorMessage(r, w, "Update failed.")
		http.Redirect(w, r, "/site-config", http.StatusSeeOther)
		return
	}
	// Success.
	models.WriteLog(models.LogInfo, "Config Updated: "+config.Name)
	auth.AddSuccessMessage(r, w, "Config updated!")
	http.Redirect(w, r, "/site-config", http.StatusSeeOther)
}
