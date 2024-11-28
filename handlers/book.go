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

// GET.
func GetBooks(w http.ResponseWriter, r *http.Request) {
	var page = r.URL.Query().Get("page")
	var search = r.URL.Query().Get("search")
	var partial = r.URL.Query().Get("partial")
	var templateName = "library.html"
	if partial == "1" {
		templateName = "part_books.html"
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 0
	}
	pageInt++
	var books = models.GetBooks(pageInt, search, auth.GetLoggedInUserID(r))
	var pageData = controllers.NewPageData(r, config.LibraryTitle, config.LibraryDesc, books)
	pageData.SearchParm = search
	// If no more data, let client know.
	if len(books) <= 0 && partial == "1" {
		pageData.SuccessMsg = "That's all, folks."
		RenderTemplate(r, w, "part_validation_messages.html", pageData)
		return
	}
	// Set page for pagination.
	pageData.PageNumber = pageInt
	// Get session messages from redirects.
	var successMsg = auth.GetSuccessMessage(r, w)
	pageData.SuccessMsg = successMsg
	RenderTemplate(r, w, templateName, pageData)
}

func GetBook(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["id"]
	intID, err := strconv.Atoi(bookID)
	if err != nil {
		http.Error(w, "Book not found for user.", http.StatusNotFound)
		return
	}
	var book = models.GetBook(intID, auth.GetLoggedInUserID(r))
	var pageData = controllers.NewPageData(r, config.LibraryTitle, config.LibraryDesc, book)
	RenderTemplate(r, w, "book.html", pageData)
}

func ExportBook(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["id"]
	intID, err := strconv.Atoi(bookID)
	if err != nil {
		http.Error(w, "Book not found for user.", http.StatusNotFound)
		return
	}
	var book = models.GetBook(intID, auth.GetLoggedInUserID(r))
	var fileName = auth.GetLoggedInUserName(r) + "_single_export.md"
	var filePath = config.ExportDir + "/" + fileName
	exportFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "File creation failed.", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(exportFile, printBook(&book))
	exportFile.Close()
	// Force a download with the content- disposition field, otherwise text file will just be rendered directly by the browser.
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	http.ServeFile(w, r, filePath)
}

func ExportBooks(w http.ResponseWriter, r *http.Request) {
	// TODO: Allow option to get all books, without pagination.
	var books = models.GetAllBooks(auth.GetLoggedInUserID(r))
	var fileName = auth.GetLoggedInUserName(r) + "_multi_export.md"
	var filePath = config.ExportDir + "/" + fileName
	exportFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "File creation failed.", http.StatusInternalServerError)
		return
	}
	for _, v := range books {
		fmt.Fprint(exportFile, printBook(&v))
	}
	exportFile.Close()
	// Force a download with the content- disposition field, otherwise text file will just be rendered directly by the browser.
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	http.ServeFile(w, r, filePath)
}

func printBook(book *models.Book) string {
	var output = "# Annotations for " + book.Title + "\n"
	for _, v := range book.Annotations {
		if v.AnnotationType == "highlight" {
			output += "## Highlight" + "\n"
		} else {
			output += "## Note" + "\n"
		}
		output += v.Text + "\n"
		if v.AnnotationType == "note" {
			output += "\n" + "---" + "\n\n" + v.Note + "\n"
		}
	}
	return output
}
