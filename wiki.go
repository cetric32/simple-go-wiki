package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

var templatesDir = "templates/*.html"
var templates = template.Must(template.ParseGlob(templatesDir))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func (p *Page) Save() error {
	filename := "data/" + p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"

	body, error := os.ReadFile(filename)

	if error != nil {
		return nil, error
	}

	return &Page{Title: title, Body: body}, nil
}

func handler(resp http.ResponseWriter, req *http.Request) {
	println("Incoming Request", req.URL.Path)
	path := req.URL.Path[1:]

	if len(path) == 0 {
		http.Redirect(resp, req, "/view/FrontPage", http.StatusFound)
	} else {
		fmt.Fprintf(resp, "Hi there, I love %s!", path)
	}
}

func viewHandler(resp http.ResponseWriter, req *http.Request) {
	title := req.URL.Path[len("/view/"):]

	p, _ := loadPage(title)

	fmt.Fprintf(resp, "<h1>%s</h1><div>%s</div> <div><a href=\"%s\" >Edit Page<a/></div>", p.Title, p.Body, "/edit/"+p.Title)
}

func editHandler(resp http.ResponseWriter, req *http.Request) {
	title := req.URL.Path[len("/edit/"):]

	p, err := loadPage(title)

	if err != nil {
		p = &Page{Title: title}
	}

	fmt.Fprintf(resp, "<h1>Editing %s</h1>"+
		"<form action=\"/save/%s\" method=\"POST\">"+
		"<textarea name=\"body\">%s</textarea><br>"+
		"<input type=\"submit\" value=\"Save\">"+
		"</form>",
		p.Title, p.Title, p.Body)

}

func newEditHandler(resp http.ResponseWriter, req *http.Request, title string) {
	p, err := loadPage(title)

	if err != nil {
		p = &Page{Title: title}
	}

	renderTemplate(resp, "edit", p)
}

func newViewHandler(resp http.ResponseWriter, req *http.Request, title string) {
	p, err := loadPage(title)

	if title == "FrontPage" {
		renderTemplate(resp, "front", p)

		return
	}

	if err != nil {
		http.Redirect(resp, req, "/edit/"+title, http.StatusFound)

	}

	renderTemplate(resp, "view", p)
}

func saveHandler(resp http.ResponseWriter, req *http.Request, title string) {
	body := req.FormValue("body")

	P := Page{Title: title, Body: []byte(body)}

	err := P.Save()

	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(resp, req, "/view/"+title, http.StatusFound)
}

func renderTemplate(resp http.ResponseWriter, temp string, p *Page) {

	err := templates.ExecuteTemplate(resp, temp+".html", p)

	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
	}
}

func getTitle(resp http.ResponseWriter, req *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(req.URL.Path)

	if m == nil {
		http.NotFound(resp, req)

		return "", errors.New("invalid Page Title")
	}

	return m[2], nil // The title is the second subexpression.
}

func makeHandler(fn func(resp http.ResponseWriter, req *http.Request, title string)) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		println("Incoming Request", req.URL.Path)
		m := validPath.FindStringSubmatch(req.URL.Path)

		if m == nil {
			http.NotFound(resp, req)
			return
		}

		// call the respective handler here
		fn(resp, req, m[2])
	}
}

func createFolders() error {
	_, err := os.Stat("data")

	if os.IsNotExist(err) {
		println("Data folder does not exist,creating the folder now")
		err = os.Mkdir("data", 0755)
	} else {
		println("Data folder exists")
	}

	return err
}

func main() {
	error := createFolders()

	if error != nil {
		println("Application error", error)
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/edit/", makeHandler(newEditHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/view/", makeHandler(newViewHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
