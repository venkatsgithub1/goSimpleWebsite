package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"ps/web/test1/src/data"
	"strings"
	"time"
)

// Response hold user struct
type Response struct {
	user data.User
}

// Res variable to hold user data.
var Res Response

/*
// MyMux struct handles routes.
type MyMux struct{}

// overriding ServeHTTP method with struct MyMux.
// takes two parameters responseWriter and pointer to request.
func (m *MyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path {
	case "/":
		renderTemplate(w, "", "loginOrSignup")
		return
	case "/login":
		renderTemplate(w, "", "login")
		return
	case "/signup":
		renderTemplate(w, "", "signup")
		return
	case "/templates":
		fmt.Println("here")
		return
	}

	// if request's url path doesn't match any. Using not found to show 404.
	http.NotFound(w, r)
}*/

func redirectIfSessionIsActive(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("_cookie")

	if err == http.ErrNoCookie {
		return
	}

	if err != nil {
		panicErrs(err)
	}

	fmt.Println(sessionCookie.Value)

	session := data.Session{}

	validSession, err := session.CheckSession(sessionCookie.Value)

	panicErrs(err)

	if validSession {
		fmt.Println(session)
		Res.user.ID = session.UserID
		Res.user.GetUserByID()
		http.Redirect(w, r, "/successfulLogin", 302)
	}

}

// login function routes to login page.
func login(w http.ResponseWriter, r *http.Request) {
	// First check if there is a cookie, if there is a cookie,
	// check if cookie is _cookie.
	// If yes, check if value is a valid session id in database.
	fmt.Println("login func")
	redirectIfSessionIsActive(w, r)
	fmt.Println("--------------------")
	fmt.Println("Login")
	r.ParseForm()
	fmt.Println(r.Form)
	renderTemplate(w, "", "login")
	fmt.Println("--------------------")
}

// signup function routes to signup page.
func signup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("signup func")
	redirectIfSessionIsActive(w, r)
	fmt.Println("--------------------")
	fmt.Println("Signup")
	r.ParseForm()
	fmt.Println(r.Form)
	renderTemplate(w, "", "signup")
	fmt.Println("--------------------")
}

// welcome function shows welcome page.
func welcome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--------------------")
	fmt.Println("welcome")
	renderTemplate(w, "", "loginOrSignup")
	fmt.Println("--------------------")
}

// signupUser handler function takes form data from
// signup form and creates user in database.
func signupUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println(r.Form, "User signup data", len(r.Form))
	if len(r.Form) == 0 {
		http.Redirect(w, r, "/", 302)
		return
	}
	fmt.Println(r.PostFormValue("pwd"), "password")
	// get user data from form and create user in database.
	// concatenate both first and last names.
	firstName := r.PostFormValue("firstname")
	lastName := r.PostFormValue("lastname")
	userName := strings.ToTitle(lastName) + ", " + strings.ToTitle(firstName)
	user := data.User{Name: userName, Email: r.PostFormValue("email"), Password: r.PostFormValue("pwd")}
	if err := user.CreateUser(); err != nil {
		panicErrs(errors.New("Cannot create user"))
	}
	// create cookie and store user signin data in cookie.
	renderTemplate(w, "", "signupSuccessful")
}

func authenticate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.PostFormValue("email")
	password := r.PostFormValue("pwd")

	// 1=wrong email, 2=wrong password

	Message := ""

	indicator := Res.user.GetUserByEmail(email)

	if Res.user.Password == password {
		fmt.Println("signin successful")
		redirectIfSessionIsActive(w, r)
		session, err := Res.user.CreateSession()
		panicErrs(err)
		// insert session id into cookie.
		cookie := http.Cookie{
			Name:     "_cookie",
			Value:    session.UUID,
			HttpOnly: true,
		}

		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/successfulLogin", 302)
		// redirect to user page.
		// if cookie's value i.e. session is present in database
		// redirect to user page instead of login page even if /login is requested.
	} else {
		if indicator == 1 {
			Message = "wrong username"
			fmt.Println("wrong username")
		} else {
			Message = "wrong password"
			fmt.Println("wrong password")
		}
		renderTemplate(w, Message, "login")
	}

	fmt.Println(Message)
}

func signout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("_cookie")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", 302)
	}
	panicErrs(err)
	Res.user.DeleteSession(cookie.Value)
	cookie.Expires = time.Unix(0, 0)
	Res.user = data.User{}
	http.Redirect(w, r, "/", 302)
}

func signoutEverywhere(w http.ResponseWriter, r *http.Request) {
	Res.user.DeleteAllSessions()
	Res.user = data.User{}
	http.Redirect(w, r, "/", 302)
}

func successfulLogin(w http.ResponseWriter, r *http.Request) {
	if Res.user.Name == "" {
		http.Redirect(w, r, "/", 302)
	}
	renderTemplate(w, Res.user, "welcomeUserPage")
}

// renderTemplate function is generic to render html pages.
func renderTemplate(w http.ResponseWriter, data interface{}, htmlPage string) {
	tmpl, err := template.ParseFiles("./templates/html/" + htmlPage + ".html")

	checkErr(err)

	tmpl.Execute(w, data)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Println("index func")
	redirectIfSessionIsActive(w, r)
	renderTemplate(w, "", "loginOrSignup")
}

func main() {
	/*mux := &MyMux{}
	http.ListenAndServe(":8099", mux)*/
	// creating a new multiplexer.
	mux := http.NewServeMux()
	// using FileServer to serve static files.
	// using /templates as root.
	files := http.FileServer(http.Dir("./templates"))
	// using Handle function to serve static css / js files.
	mux.Handle("/static/", http.StripPrefix("/static/", files))
	// using handle func to register / handler.
	mux.HandleFunc("/", index)
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/signup", signup)
	mux.HandleFunc("/signupUser", signupUser)
	mux.HandleFunc("/authenticate", authenticate)
	mux.HandleFunc("/successfulLogin", successfulLogin)
	mux.HandleFunc("/signout", signout)
	mux.HandleFunc("/signoutEverywhere", signoutEverywhere)

	server := &http.Server{
		Addr:    ":8099",
		Handler: mux,
	}
	server.ListenAndServe()

}

func panicErrs(err error) {
	if err != nil {
		panic(err)
	}
}
