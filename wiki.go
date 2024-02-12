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

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func (p *Page) Save() error {
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"

	body, error := os.ReadFile(filename)

	if error != nil {
		return nil, error
	}

	return &Page{Title: title, Body: body}, nil
}

func handler(resp http.ResponseWriter, req *http.Request) {
	println("Incoming Request", req.URL.Path)
	fmt.Fprintf(resp, "Hi there, I love %s!", req.URL.Path[1:])
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

func newEditHandler(resp http.ResponseWriter, req *http.Request) {
	title, err := getTitle(resp, req)
	if err != nil {
		return
	}

	p, err := loadPage(title)

	if err != nil {
		p = &Page{Title: title}
	}

	renderTemplate(resp, "edit", p)
}

func newViewHandler(resp http.ResponseWriter, req *http.Request) {
	title, err := getTitle(resp, req)
	if err != nil {
		return
	}

	p, err := loadPage(title)

	if err != nil {
		http.Redirect(resp, req, "/edit/"+title, http.StatusFound)

	}

	renderTemplate(resp, "view", p)
}

func saveHandler(resp http.ResponseWriter, req *http.Request) {
	title, err := getTitle(resp, req)
	if err != nil {
		return
	}

	body := req.FormValue("body")

	P := Page{Title: title, Body: []byte(body)}

	err = P.Save()

	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(resp, req, "/view/"+title, http.StatusFound)
}

func renderTemplate(resp http.ResponseWriter, temp string, p *Page) {
	// t, err := template.ParseFiles(temp + ".html")

	// if err != nil {
	// 	http.Error(resp, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// err = t.Execute(resp, p)

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

func main() {
	// p1 := Page{Title: "hello", Body: []byte("Hello world, see you there")}

	// p1.Save()

	// p2, _ := loadPage("hello")

	// println(string(p2.Body))

	http.HandleFunc("/", handler)
	http.HandleFunc("/edit/", newEditHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/view/", newViewHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))

}