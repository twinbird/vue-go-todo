package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"gopkg.in/boj/redistore.v1"
)

const (
	SessionName        = "vue-go-todo-app"
	LoginPageTemplate  = "login.tmpl.html"
	SignupPageTemplate = "signup.tmpl.html"
	AppPageTemplate    = "app.tmpl.html"
)

var (
	cookieSecretKey []byte
	sessionStore    *redistore.RediStore
	productionMode  bool
	db              *sql.DB
	viewTemplates   map[string]*template.Template
)

func main() {
	// get configuration
	port := os.Getenv("PORT")
	cookieSecretKey = []byte(os.Getenv("SECRET_KEY"))
	if os.Getenv("ENV") == "production" {
		productionMode = true
	}

	// initialize templates
	initTemplate()

	// connect to redis
	var err error
	sessionStore, err = redistore.NewRediStore(10, "tcp", "redis:6379", "", []byte(cookieSecretKey))
	if err != nil {
		log.Fatal(err)
	}
	defer sessionStore.Close()

	// connect to postgres
	db, err = sql.Open("postgres", "host=postgres user=root dbname=todo_app password=root sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// setting handler
	r := mux.NewRouter()

	r.HandleFunc("/", rootHandler).Methods("GET")
	r.HandleFunc("/login", loginGetHandler).Methods("GET")
	r.HandleFunc("/login", loginPostHandler).Methods("POST")
	r.HandleFunc("/logout", logoutPostHandler).Methods("POST")
	r.HandleFunc("/signup", signupGetHandler).Methods("GET")
	r.HandleFunc("/signup", signupPostHandler).Methods("POST")
	r.HandleFunc("/app", authRequiredHandler(appGetHandler)).Methods("GET")
	r.HandleFunc("/app/tasks", authRequiredHandler(taskGetHandler)).Methods("GET")
	r.HandleFunc("/app/tasks", authRequiredHandler(taskPostHandler)).Methods("POST")
	r.HandleFunc("/app/tasks/{id}", authRequiredHandler(taskPutHandler)).Methods("PUT")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	s := csrf.Protect(cookieSecretKey, csrf.Secure(productionMode))(r)
	log.Fatal(http.ListenAndServe(":"+port, s))
}

func parseTemplate(filename string) (*template.Template, error) {
	return template.New(filename).Delims("[[", "]]").ParseFiles(filepath.Join("templates", filename))
}

func initTemplate() {
	viewTemplates = make(map[string]*template.Template)
	viewTemplates[LoginPageTemplate] =
		template.Must(parseTemplate(LoginPageTemplate))
	viewTemplates[SignupPageTemplate] =
		template.Must(parseTemplate(SignupPageTemplate))
	viewTemplates[AppPageTemplate] =
		template.Must(parseTemplate(AppPageTemplate))
}

func executeTemplate(w http.ResponseWriter, name string, data interface{}) {
	var t *template.Template
	var err error

	if productionMode {
		t = viewTemplates[name]
	} else {
		t, err = parseTemplate(name)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func isLoggedIn(r *http.Request) (bool, error) {
	session, err := sessionStore.Get(r, SessionName)
	if err != nil {
		return false, err
	}
	if session.Values["authenticated"] == nil || session.Values["authenticated"].(bool) == false {
		return false, nil
	} else {
		return true, nil
	}
}

func redirectToAppIfLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	ok, err := isLoggedIn(r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	if ok {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return true
	}
	return false
}

func authRequiredHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, err := isLoggedIn(r)
		if err != nil {
			log.Println(err)
			http.Error(w, "unauthorized. You need login from '/login'", http.StatusUnauthorized)
			return
		}
		if !ok {
			http.Error(w, "unauthorized. You need login from '/login'", http.StatusUnauthorized)
			return
		}
		fn(w, r)
	}
}
