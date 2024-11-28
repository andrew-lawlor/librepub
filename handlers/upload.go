package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/andrew-lawlor/librepub/auth"
	"github.com/andrew-lawlor/librepub/config"
	controllers "github.com/andrew-lawlor/librepub/controller"
	"github.com/andrew-lawlor/librepub/models"
	"github.com/h2non/filetype"
)

func UploadDB(w http.ResponseWriter, r *http.Request) {
	var pageData = controllers.NewPageData(r, config.UploadTitle, config.UploadDesc, "")
	RenderTemplate(r, w, "upload_db.html", pageData)
}

func UploadDBPost(w http.ResponseWriter, r *http.Request) {
	var MAX_UPLOAD_SIZE int64 = 1024 * 20000 // ~20 MB limit for now.
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		models.WriteLog(models.LogError, err.Error())
		http.Error(w, "File size limited to ~20MB", http.StatusBadRequest)
		return
	}
	file, fileHeader, err := r.FormFile("koboDB")
	if err != nil {
		models.WriteLog(models.LogError, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	validateFile, _, _ := r.FormFile("koboDB")
	// Validate file type.
	buf := make([]byte, 512)
	_, err = validateFile.Read(buf)
	if err != nil {
		models.WriteLog(models.LogError, err.Error())
		http.Error(w, "File corrupted or otherwise unreadable.", http.StatusBadRequest)
		return
	}
	defer validateFile.Close()
	kind, _ := filetype.Match(buf)
	if kind == filetype.Unknown {
		http.Error(w, "Unknown file type", http.StatusBadRequest)
		return
	}
	if kind.MIME.Value != "application/vnd.sqlite3" {
		http.Error(w, "Unsupported file type.", http.StatusBadRequest)
		return
	}
	defer file.Close()
	// Create the uploads folder if it doesn't already exist.
	err = os.MkdirAll("./uploads/", os.ModePerm)
	if err != nil {
		models.WriteLog(models.LogError, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Create a new file in the uploads directory
	var fileName = fmt.Sprintf("%d%s%s", time.Now().UnixNano(), auth.GetLoggedInUserName(r), filepath.Ext(fileHeader.Filename))
	var filePath = "./uploads/" + fileName
	dst, err := os.Create(filePath)
	if err != nil {
		models.WriteLog(models.LogError, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Copy the uploaded file to the filesystem at the specified destination.
	_, err = io.Copy(dst, file)
	if err != nil {
		models.WriteLog(models.LogError, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bookmarks, err := models.GetKoboBookmarks(filePath)
	if err != nil {
		var pageData = controllers.NewPageData(r, config.UploadTitle, config.UploadDesc, "")
		pageData.ErrorMsgs = append(pageData.ErrorMsgs, "Your database is corrupted, invalid, or simply incompatible with this web app.")
		RenderTemplate(r, w, "upload_db.html", pageData)
		return
	}
	// DB is good and we have rows, process bookmarks.
	var userID = auth.GetLoggedInUserID(r)
	for _, v := range bookmarks {
		// Skip if annotation exists with unique kobo id.
		if models.AnnotationExists(v.BookmarkID, userID) {
			continue
		}
		// Next, create a book for this annotation if it doesn't exist. Use volume id from Kobo.
		var bookID = models.BookExists(v.VolumeID, userID)
		if bookID == -1 {
			bookID, err = models.CreateBook(v.VolumeID, userID, v.BookTitle)
			if err != nil {
				models.WriteLog(models.LogError, "Failed to create book. Skipping annotation.")
				continue
			}
		}
		// Next, create an annotation and associate it to bookID.
		_, err := models.CreateAnnotation(v.BookmarkID, bookID, v.Text, v.Annotation, v.AnnotationType, userID)
		if err != nil {
			models.WriteLog(models.LogError, "Failed to create annotation. Skipping.")
			continue
		}
	}
	// Extract vocab builder data if desired.
	var exportVocab = r.FormValue("exportVocab")
	if exportVocab == "on" {
		fmt.Println("Exporting vocab.")
		vocab, _ := models.ExportVocab(filePath)
		models.ImportVocab(vocab, userID)
	}
	// Clean up temp file - moved to asynch goroutine.
	// os.Remove(filePath)
	// Redirect to library page after successful upload.
	auth.AddSuccessMessage(r, w, "Kobo database uploaded successfully!")
	http.Redirect(w, r, "/library", http.StatusSeeOther)
}
