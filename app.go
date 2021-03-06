package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

// version
var version = "1.0.0"

var dataDir = "data/"

// Template Caching
var templates = template.Must(template.ParseFiles("views/view.html", "views/edit.html"))

// validation
var validPath = regexp.MustCompile("^/(view|edit|save|delete)/([a-zA-Z0-9]+)$")

// Page Title and Body
type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	if os.Mkdir(dataDir, 0777) != nil {
		fmt.Println(dataDir + " is already exist")
	}

	filename := p.Title + ".txt"
	err := ioutil.WriteFile(dataDir+filename, p.Body, 0600)
	if err != err {
		fmt.Println("Page not found")
		return err
	}

	return nil
}

func (p *Page) delete() error {
	err := os.Remove(dataDir + p.Title + ".txt")
	if err != nil {
		return err
	}
	return nil
}

func loadPage(title string) (*Page, error) {
	body, err := ioutil.ReadFile(dataDir + title + ".txt")
	if err != nil {
		return nil, err
	}
	page := &Page{Title: title, Body: body}
	return page, nil
}

// HTML Rendering
func renderTemplate(w http.ResponseWriter, tmpl string, page *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HTTP Handler
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		matched := validPath.FindStringSubmatch(r.URL.Path)
		if matched == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, matched[2])
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		fmt.Println(title + " is not found")
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
	}

	renderTemplate(w, "view", page)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		page = &Page{Title: title}
	}

	renderTemplate(w, "edit", page)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := []byte(r.FormValue("body"))
	page := &Page{Title: title, Body: body}
	err := page.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if page.Title == "FrontPage" {
		fmt.Println("FrontPageは削除できません")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	err = page.delete()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Println("DELETE SUCCESS")
	http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
	// version管理
	showVersion := false
	// -v が指定されたときに、showVersion変数がtrueにする
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()
	if showVersion {
		fmt.Println("version: ", version)
		return
	}

	// httpのルーテゥングとハンドラの設定
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/delete/", makeHandler(deleteHandler))
	http.ListenAndServe(":8080", nil)
}
