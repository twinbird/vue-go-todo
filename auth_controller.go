package main

import (
	"log"
	"net/http"

	"github.com/gorilla/csrf"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if redirectToAppIfLoggedIn(w, r) {
		return
	}
	http.Redirect(w, r, "/static/top.html", http.StatusSeeOther)
}

func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	if redirectToAppIfLoggedIn(w, r) {
		return
	}
	executeTemplate(w, LoginPageTemplate, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

func signupGetHandler(w http.ResponseWriter, r *http.Request) {
	if redirectToAppIfLoggedIn(w, r) {
		return
	}
	executeTemplate(w, SignupPageTemplate, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	if redirectToAppIfLoggedIn(w, r) {
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if ok, id := authenticate(email, password); ok {
		session, err := sessionStore.Get(r, SessionName)
		if err != nil {
			log.Printf("auth: session get %v\n", err)
			executeTemplate(w, LoginPageTemplate, map[string]interface{}{
				"email":          email,
				"message":        "エラーが発生しました。もう一度お試しください。",
				csrf.TemplateTag: csrf.TemplateField(r),
			})
			return
		}
		session.Values["authenticated"] = true
		session.Values["user_id"] = id
		err = session.Save(r, w)
		if err != nil {
			log.Printf("auth: session save %v\n", err)
			executeTemplate(w, LoginPageTemplate, map[string]interface{}{
				"email":          email,
				"message":        "エラーが発生しました。もう一度お試しください。",
				csrf.TemplateTag: csrf.TemplateField(r),
			})
			return
		}
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}

	executeTemplate(w, LoginPageTemplate, map[string]interface{}{
		"email":          email,
		"message":        "メールアドレスかパスワードに誤りがあります。",
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

func signupPostHandler(w http.ResponseWriter, r *http.Request) {
	if redirectToAppIfLoggedIn(w, r) {
		return
	}

	req := &SignupValidator{
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		PasswordConfirm: r.FormValue("passwordConfirm"),
	}

	if ok, messages := req.validate(); !ok {
		executeTemplate(w, SignupPageTemplate, map[string]interface{}{
			"email":          req.Email,
			"messages":       messages,
			csrf.TemplateTag: csrf.TemplateField(r),
		})
		return
	}

	var user_id int
	var err error
	if user_id, err = registUser(req.Email, req.Password); err != nil {
		log.Println(err)
		executeTemplate(w, SignupPageTemplate, map[string]interface{}{
			"email":          req.Email,
			"message":        "エラーが発生しました。時間をおいてもう一度やり直してください。",
			csrf.TemplateTag: csrf.TemplateField(r),
		})
		return
	}
	session, err := sessionStore.Get(r, SessionName)
	if err != nil {
		log.Println(err)
		executeTemplate(w, LoginPageTemplate, map[string]interface{}{
			"email":          req.Email,
			"message":        "登録が完了しました。ログインしてください。",
			csrf.TemplateTag: csrf.TemplateField(r),
		})
		return
	}
	session.Values["authenticated"] = true
	session.Values["user_id"] = user_id
	err = session.Save(r, w)
	if err != nil {
		log.Println(err)
		executeTemplate(w, LoginPageTemplate, map[string]interface{}{
			"email":          req.Email,
			"message":        "登録が完了しました。ログインしてください。",
			csrf.TemplateTag: csrf.TemplateField(r),
		})
		return
	}
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

func logoutPostHandler(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, SessionName)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/static/top.html", http.StatusSeeOther)
}
