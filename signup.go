package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"

	//Go get both required packages listed below
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// Caches all the templates from the views folder int the value views
// The full path to the views folder
var views = template.Must(template.ParseGlob("*.html"))

// Used by context
type key int

const MyKey key = 0

// Error messages will be type struct for context to pass around
type loginerror struct {
	// And the message is stored as a string
	// Skip down to the next comment
	Err string
}

var db *sql.DB
var err error

func signupPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "signup.html")
		return
	}
	username := req.FormValue("username")
	password := req.FormValue("password")
	var user string
	err := db.QueryRow("SELECT username FROM users WHERE username=?", username).Scan(&user)
	switch {
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}
		_, err = db.Exec("INSERT INTO users(username, password) VALUES(?, ?)", username, hashedPassword)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}
		http.Redirect(res, req, "/login", 301)
		return
	case err != nil:
		http.Error(res, "Server error, unable to create your account.", 500)
		return
	default:
		http.Redirect(res, req, "/", 301)
	}
}

// This function gets called if a login error occurs
func login(res http.ResponseWriter, req *http.Request) {
	// grab the context value (the message)
	// le short for login error
	le, _ := req.Context().Value(MyKey).(loginerror)
	// send the user the login template with the error message
	views.ExecuteTemplate(res, "Login", loginerror{le.Err})
}

func loginPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "login.html")
		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")

	var databaseUsername string
	var databasePassword string

	err := db.QueryRow("SELECT username, password FROM users WHERE username=?", username).Scan(&databaseUsername, &databasePassword)

	// uh oh error! let's tell the user
	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password))
	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return
	}

	http.Redirect(res, req, "/blog", 301)
}

func homePage(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "index.html")
}

func blogPage(res http.ResponseWriter, req *http.Request) {

	if req.Method != "POST" {
		http.ServeFile(res, req, "blog.html")
		return
	}
}

func main() {
	db, err = sql.Open("mysql", "root:linda12345@/user")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	http.Handle("/asset/", http.StripPrefix("/asset/", http.FileServer(http.Dir("assets"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("css/images"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))

	http.HandleFunc("/signup", signupPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/blog", blogPage)
	http.HandleFunc("/", homePage)
	fmt.Println("Server Starting on http://localhost:5500")
	http.ListenAndServe(":5500", nil)
}
