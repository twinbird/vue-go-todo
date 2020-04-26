package main

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/soveran/redisurl"
	"gopkg.in/boj/redistore.v1"
)

const (
	SessionName             = "vue-go-todo-app"
	SessionAuthenticatedKey = "SessionAuthenticatedKey"
	SessionUserIdKey        = "SessionUserIdKey"

	LoginPageTemplate  = "login.tmpl.html"
	SignupPageTemplate = "signup.tmpl.html"
	AppPageTemplate    = "app.tmpl.html"

	ContextUserIdKey = "ContextUserIdKey"

	RedisMaxCon = 10
)

var (
	cookieSecretKey []byte
	sessionStore    *redistore.RediStore
	productionMode  bool
	db              *sql.DB
	viewTemplates   map[string]*template.Template
	postgresURL     string
	redisURL        string
)

func getRediStore(url string) (*redistore.RediStore, error) {
	if strings.HasPrefix(url, "redis://") {
		redisPool := redis.NewPool(func() (redis.Conn, error) {
			return redisurl.ConnectToURL(redisURL)
		}, RedisMaxCon)
		return redistore.NewRediStoreWithPool(redisPool, []byte(cookieSecretKey))
	} else {
		return redistore.NewRediStore(10, "tcp", redisURL, "", []byte(cookieSecretKey))
	}
}

func main() {
	// get configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	cookieSecretKey = []byte(os.Getenv("SECRET_KEY"))
	if os.Getenv("ENV") == "development" {
		productionMode = false
	}
	postgresURL = os.Getenv("DATABASE_URL")
	redisURL = os.Getenv("REDIS_URL")

	// initialize templates
	initTemplate()

	// connect to redis
	var err error
	sessionStore, err = getRediStore(redisURL)
	if err != nil {
		log.Fatal(err)
	}
	defer sessionStore.Close()

	// connect to postgres
	db, err = sql.Open("postgres", postgresURL)
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
	r.HandleFunc("/app/tasks", authRequiredHandler(taskDeleteHandler)).Methods("DELETE")
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

func getAuthInfo(r *http.Request) (bool, int, error) {
	session, err := sessionStore.Get(r, SessionName)
	if err != nil {
		return false, 0, err
	}
	if session.Values[SessionAuthenticatedKey] == nil || session.Values[SessionAuthenticatedKey].(bool) == false {
		return false, 0, nil
	}
	if session.Values[SessionUserIdKey] == nil {
		return false, 0, nil
	}
	user_id := session.Values[SessionUserIdKey].(int)
	return true, user_id, nil
}

func redirectToAppIfLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	ok, _, err := getAuthInfo(r)
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
		ok, uid, err := getAuthInfo(r)
		if err != nil {
			log.Println(err)
			http.Error(w, "unauthorized. You need login from '/login'", http.StatusUnauthorized)
			return
		}
		if !ok {
			http.Error(w, "unauthorized. You need login from '/login'", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ContextUserIdKey, uid)
		fn(w, r.WithContext(ctx))
	}
}
