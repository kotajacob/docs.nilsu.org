package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strconv"
	text "text/template"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// parseTemplateFiles reads a list of filenames to create a template
func parseTemplateFiles(filenames ...string) *template.Template {
	var files []string
	t := template.New("layout")
	for _, file := range filenames {
		files = append(files, path.Join(config.TemplateDir, fmt.Sprintf("%s.html", file)))
	}
	t = template.Must(t.ParseFiles(files...))
	return t
}

// writeHTML to a http.ResponseWriter based on a list of template files
func writeHTML(w http.ResponseWriter, data interface{}, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, path.Join(config.TemplateDir, fmt.Sprintf("%s.html", file)))
	}

	templates := template.Must(template.ParseFiles(files...))
	templates.ExecuteTemplate(w, "layout", data)
}

// writeHTMLUnsafe to a http.ResponseWriter based on a list of template files
func writeHTMLUnsafe(w http.ResponseWriter, data interface{}, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, path.Join(config.TemplateDir, fmt.Sprintf("%s.html", file)))
	}

	templates := text.Must(text.ParseFiles(files...))
	templates.ExecuteTemplate(w, "layout", data)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, nil, "baseof", "index")
}

// CreateHandler generates a new random document name, parses form data, write
// the document to disk, and sends a redirect to the view page.
func CreateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.NewRandom()
	if err != nil {
		log.Printf("error generating random UUID %v\n", err)
		http.Error(w, "500 - Error! Could not create document.", http.StatusInternalServerError)
		return
	}

	doc := new(Doc)
	name := base64.RawURLEncoding.EncodeToString(id[:])
	doc.Name = &name
	doc.Hash()
	err = r.ParseForm()
	if err != nil {
		log.Printf("failed parsing form data %v\n", err)
		http.Error(w, "500 - Error! Could not create document.", http.StatusInternalServerError)
		return
	}
	data := r.Form.Get("data")
	doc.Data = &data
	if err := doc.Write(); err != nil {
		log.Printf("failed writing new document %v\n", err)
		http.Error(w, "500 - Error! Could not create document.", http.StatusInternalServerError)
		return
	}
	log.Printf("created %v\n", name)
	http.Redirect(w, r, path.Join("/view", name), 302)
}

// ViewLatestHandler opens the latest version of a document and prints the data.
func ViewLatestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	log.Printf("%s requested\n", name)
	doc, err := ReadDocLatest(name)
	if err != nil {
		log.Printf("failed opening document %v\n", err)
		http.Error(w, "404 - Error! Document not found.", http.StatusNotFound)
		return
	}
	if err := doc.HTML(); err != nil {
		log.Printf("failed coverting document to html %v\n", err)
		http.Error(w, "404 - Error! Document not found.", http.StatusNotFound)
		return
	}
	writeHTMLUnsafe(w, doc, "baseof", "view")
}

// ViewTimeHandler a specific version of a document and prints the data.
func ViewTimeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	u, err := strconv.ParseInt(vars["time"], 10, 64)
	if err != nil {
		log.Printf("failed opening document %v\n", err)
		http.Error(w, "404 - Error! Document not found.", http.StatusNotFound)
		return
	}
	t := time.Unix(u, 0)
	log.Printf("%s requested at %s\n", name, t)
	doc, err := ReadDoc(name, t)
	if err != nil {
		log.Printf("failed opening document %v\n", err)
		http.Error(w, "404 - Error! Document not found.", http.StatusNotFound)
		return
	}
	if err := doc.HTML(); err != nil {
		log.Printf("failed coverting document to html %v\n", err)
		http.Error(w, "404 - Error! Document not found.", http.StatusNotFound)
		return
	}
	writeHTMLUnsafe(w, doc, "baseof", "view")
}

// ReadLatestHandler opens the latest version of a hash document and prints the data.
func ReadLatestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	log.Printf("%s requested\n", hash)
	doc, err := ReadHashLatest(hash)
	if err != nil {
		log.Printf("failed opening document %v\n", err)
		http.Error(w, "404 - Error! Document not found.", http.StatusNotFound)
		return
	}
	if err := doc.HTML(); err != nil {
		log.Printf("failed coverting document to html %v\n", err)
		http.Error(w, "404 - Error! Document not found.", http.StatusNotFound)
		return
	}
	writeHTMLUnsafe(w, doc, "baseof", "read")
}

// ReadTimeHandler opens a specific version of a hash document and prints the data.
func ReadTimeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	u, err := strconv.ParseInt(vars["time"], 10, 64)
	if err != nil {
		log.Printf("failed opening document %v\n", err)
		http.Error(w, "404 - Error! Document not found.", http.StatusNotFound)
		return
	}
	t := time.Unix(u, 0)
	log.Printf("%s requested at %s\n", hash, t)
	doc, err := ReadHash(hash, t)
	if err != nil {
		log.Printf("failed opening document %v\n", err)
		http.Error(w, "404 - Error! Document not found.", http.StatusNotFound)
		return
	}
	if err := doc.HTML(); err != nil {
		log.Printf("failed coverting document to html %v\n", err)
		http.Error(w, "404 - Error! Document not found.", http.StatusNotFound)
		return
	}
	writeHTMLUnsafe(w, doc, "baseof", "read")
}

// EditHandler opens a document from the name in the URL and inserts the
// data into a <textarea> field with an appropriate save button.
func EditHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	log.Printf("%s requested\n", name)
	doc, err := ReadDocLatest(name)
	if err != nil {
		log.Printf("failed opening document %v\n", err)
		http.Error(w, "404 - Error! Document not found.", http.StatusNotFound)
		return
	}
	writeHTML(w, doc, "baseof", "edit")
}

// SaveHandler saves a document and it's data by name.
func SaveHandler(w http.ResponseWriter, r *http.Request) {
	doc := new(Doc)
	err := r.ParseForm()
	if err != nil {
		log.Printf("failed parsing form data %v\n", err)
		http.Error(w, "500 - Error! Could not save document.", http.StatusInternalServerError)
		return
	}
	name := r.Form.Get("name")
	doc.Name = &name
	doc.Hash()
	data := r.Form.Get("data")
	doc.Data = &data
	if err := doc.Write(); err != nil {
		log.Printf("failed writing new document %v\n", err)
		http.Error(w, "500 - Error! Could not create document.", http.StatusInternalServerError)
		return
	}
	log.Printf("saved %v\n", *doc.Name)
	http.Redirect(w, r, path.Join("/view", *doc.Name), 302)
}
