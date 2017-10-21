package main

import (
	"text/template"
	"net/http"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var db *sql.DB
var err error

type Page struct {
	Id int
	Title string
}

type DbConfig struct {
	User string
	Password string
	DbName string
}
type Pages []*Page

func fatalError(err error)  {
	if err != nil {
		log.Fatal(err)
	}
}

func renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	t := template.Must(template.ParseFiles(
		"templates/layout.tmpl",
		"templates/" + templateName + ".tmpl",
	))
	t.ExecuteTemplate(w, "base", data)
}

func getPage(pageId string) Page {
	rows, err := db.Query("SELECT * FROM pages WHERE id = ?", pageId)
	fatalError(err)

	defer rows.Close()

	page := Page{}

	for rows.Next() {
		var id int
		var title string

		err = rows.Scan(&id, &title)
		fatalError(err)

		page.Id = id
		page.Title = title
	}

	return page
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM pages")
	fatalError(err)

	defer rows.Close()

	page := Page{}
	pages := []Page{}

	for rows.Next() {
		var id int
		var title string

		err = rows.Scan(&id, &title)
		fatalError(err)

		page.Id = id
		page.Title = title

		pages = append(pages, page)
	}

	renderTemplate(w, "list", pages)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	pageId := r.URL.Path[len("/view/"):]

	page := getPage(pageId)

	renderTemplate(w, "view", page)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	pageId := r.URL.Path[len("/edit/"):]

	if r.Method == "POST" {

		title := r.FormValue("title")
		stmt, err := db.Prepare("UPDATE pages SET title = ? WHERE id = ?")
		fatalError(err)

		stmt.Exec(title, pageId)

		log.Printf("Page updated. New title: %s\n", title)

		http.Redirect(w, r, "/", 302)
	} else {
		page := getPage(pageId)
		renderTemplate(w, "edit", page)
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		title := r.FormValue("title")

		stmt, err := db.Prepare("INSERT INTO pages (title) VALUES(?)")
		fatalError(err)

		inserted, err := stmt.Exec(title)
		fatalError(err)

		insertedId, err := inserted.LastInsertId()
		fatalError(err)

		log.Printf("New page inserted id: %d\n", insertedId)

		http.Redirect(w, r, "/", 302)
	} else {
		renderTemplate(w, "edit", Page{})
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	pageId := r.URL.Path[len("/delete/"):]
	if r.Method == "POST" {
		stmt, err := db.Prepare("DELETE FROM pages WHERE id=?")
		fatalError(err)

		deleted, err := stmt.Exec(pageId)

		affected, err := deleted.RowsAffected()

		log.Printf("Page deleted. Affected rows %d", affected)

		http.Redirect(w, r, "/", 302)
	} else {
		page := getPage(pageId)
		renderTemplate(w, "delete", page)
	}
}

func initDb() {
	dbConfig := DbConfig{}

	source, err := ioutil.ReadFile("resource/db.yml")
	fatalError(err)

	err = yaml.Unmarshal(source, &dbConfig)
	fatalError(err)

	db, err = sql.Open("mysql", dbConfig.User+":"+dbConfig.Password+"@/"+dbConfig.DbName)
	fatalError(err)
}

func main() {
	initDb()
	defer db.Close()

	http.HandleFunc("/", listHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/delete/", deleteHandler)
	http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}