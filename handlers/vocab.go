package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/andrew-lawlor/librepub/auth"
	"github.com/andrew-lawlor/librepub/config"
	controllers "github.com/andrew-lawlor/librepub/controller"
	"github.com/andrew-lawlor/librepub/models"
	"github.com/gorilla/mux"
)

func Vocab(w http.ResponseWriter, r *http.Request) {
	var page = r.URL.Query().Get("page")
	var search = r.URL.Query().Get("search")
	var partial = r.URL.Query().Get("partial")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 0
	}
	pageInt++
	var vocab = models.GetVocab(pageInt, search, auth.GetLoggedInUserID(r))
	var pageData = controllers.NewPageData(r, config.VocabTitle, config.VocabDesc, vocab)
	pageData.SearchParm = search
	// If no more data, let client know and get out.
	if len(vocab) <= 0 {
		pageData.SuccessMsg = "No vocab found."
		// Just show feedback to client if this is a partial operation.
		if partial == "1" {
			fmt.Fprint(w, "<tr><td colspan='4' class='centered-text success'>No more vocab.</td></tr>")
			return
		} else { // Load full page with feedback if this is a full page load.
			RenderTemplate(r, w, "vocab.html", pageData)
		}
		return
	}
	// Set page for pagination.
	pageData.PageNumber = pageInt
	// Render entire page or partial template, depending on partial query parm.
	var templateName = "vocab.html"
	if partial == "1" {
		templateName = "part_vocab.html"
	}
	RenderTemplate(r, w, templateName, pageData)
}

func ExportVocab(w http.ResponseWriter, r *http.Request) {
	var vocab = models.GetAllVocab(auth.GetLoggedInUserID(r))
	var fileName = auth.GetLoggedInUserName(r) + "_vocab_export.tsv"
	var filePath = config.ExportDir + "/" + fileName
	exportFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "File creation failed.", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(exportFile, printVocab(vocab))
	exportFile.Close()
	// Force a download with the content- disposition field, otherwise text file will just be rendered directly by the browser.
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	http.ServeFile(w, r, filePath)
}

func DeleteVocab(w http.ResponseWriter, r *http.Request) {
	var id = mux.Vars(r)["id"]
	idInt, err := strconv.Atoi(id)
	if err != nil {
		auth.AddErrorMessage(r, w, "Invalid ID.")
		http.Redirect(w, r, "/vocab", http.StatusSeeOther)
		return
	}
	models.DeleteVocab(idInt, auth.GetLoggedInUserID(r))
}

func printVocab(vocab []models.Vocab) string {
	var tsvString string
	for _, v := range vocab {
		tsvString += v.Word + "\t" + v.Definition + "\n"
	}
	return tsvString
}
