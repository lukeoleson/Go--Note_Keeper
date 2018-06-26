package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/russross/blackfriday"
)

// Note - define whats in the database
type Note struct {
	ID        int
	Title     string
	Content   string
	Body      template.HTML
	UpdatedAt string
	CreatedAt string
}

// get notes list
func getNoteRows() []Note {

	// what we want to return from this function
	var notes []Note

	rows, err := globalDB.Query("SELECT id, title FROM notes ORDER BY updated_at DESC")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {

		// tmp place to stash our values before putting them into the return slice
		note := Note{}

		err := rows.Scan(&note.ID, &note.Title)
		if err != nil {
			log.Fatal(err)
		}

		// append note onto notes
		notes = append(notes, note)

	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return notes

}

// fetch note record
func getNote(noteID int) Note {

	note := Note{}

	err := globalDB.QueryRow("SELECT id, title, content FROM notes WHERE id = ?", noteID).Scan(&note.ID, &note.Title, &note.Content)
	if err != nil {
		panic(err)
	}

	return note

}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	// grab all notes from database
	notes := getNoteRows()

	// pass these to template
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, notes)
}

func newHandler(w http.ResponseWriter, r *http.Request) {

	// if the http request (r) says post (returned from form?), then
	if r.Method == "POST" {

		// Gets the title and content from form feilds of same name on the edit page
		title := r.FormValue("title")
		content := r.FormValue("content")

		// Prepares a statement to be execute after (could not pass variables directly in)
		stmt, err := globalDB.Prepare("INSERT INTO notes (title, content) VALUES (?, ?)")
		if err != nil {
			panic(err)
		}

		// executes sql statement prepared earlier passing in variables from form
		res, err := stmt.Exec(title, content)
		if err != nil {
			panic(err)
		}

		// grab new note id from database
		id, err := res.LastInsertId()
		if err != nil {
			panic(err)
		}

		// Redirect to index page
		redirectURL := fmt.Sprintf("/view/%d", id)
		http.Redirect(w, r, redirectURL, http.StatusFound)

	} else {
		t, _ := template.ParseFiles("templates/new.html")
		t.Execute(w, nil)
	}

}

func viewHandler(w http.ResponseWriter, r *http.Request) {

	// need to convert string to int
	//id := path.Base(r.URL.Path)

	// string to int
	id, _ := strconv.Atoi(path.Base(r.URL.Path))

	// grab single note
	note := getNote(id)

	note.Body = template.HTML(blackfriday.MarkdownCommon([]byte(note.Content)))
	//fmt.Println(string(tmp))

	t, _ := template.ParseFiles("templates/view.html")
	t.Execute(w, note)

}

func editHandler(w http.ResponseWriter, r *http.Request) {

	// are we submitting something to this form?
	if r.Method == "POST" {

		// get form values
		id, _ := strconv.Atoi(path.Base(r.URL.Path))
		title := r.FormValue("title")
		content := r.FormValue("content")

		// update database
		stmt, err := globalDB.Prepare("UPDATE notes set title = ?, content = ?, updated_at = datetime() WHERE id = ?")

		_, err = stmt.Exec(title, content, id)
		if err != nil {
			panic(err)
		}

		strconv.Itoa(id)

		http.Redirect(w, r, "/view/"+strconv.Itoa(id), http.StatusFound)

	} else {

		id, _ := strconv.Atoi(path.Base(r.URL.Path))

		note := getNote(id)

		t, _ := template.ParseFiles("templates/edit.html")
		t.Execute(w, note)

	}

}

func deleteHandler(w http.ResponseWriter, r *http.Request) {

	// get note id
	id, _ := strconv.Atoi(path.Base(r.URL.Path))

	// Delete note from database
	_, err := globalDB.Exec("DELETE FROM notes WHERE id=?", id)
	if err != nil {
		panic(err)
	}

	//	redirect to index page
	http.Redirect(w, r, "/", http.StatusFound)

}

// db connection handler
var globalDB *sql.DB

func main() {

	// open our notes database file
	// NOTE: we are not checking for an error here right now
	globalDB, _ = sql.Open("sqlite3", "./notes.sqlite3")
	// if error not nothing.. do something!
	// if err != nil {
	// 	panic(err)
	// }

	// good idea to close database after use
	// but not until program is done
	defer globalDB.Close()

	// serve static files from /static
	fs := http.FileServer(http.Dir("templates/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// plumb up our pages
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/new", newHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/delete/", deleteHandler)
	http.ListenAndServe(":8000", nil)

}
