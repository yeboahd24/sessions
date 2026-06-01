package handlers

import (
	"html/template"
	"net/http"
	"os"

	"music-session-app/internal/auth"
	"music-session-app/internal/store"

	"github.com/alexedwards/scs/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Users    *store.UserStore
	Sessions *scs.SessionManager
}

// loginView is the data passed to templates/login.html so we can re-render the
// styled page with inline feedback instead of dumping a plain-text error page.
type loginView struct {
	Error    string
	Notice   string
	Tab      string // which tab to open: "login" or "register"
	Username string // preserved so the user doesn't have to retype it
}

func renderLogin(w http.ResponseWriter, status int, v loginView) {
	if v.Tab == "" {
		v.Tab = "login"
	}
	tmpl := template.Must(template.ParseFiles("templates/login.html"))
	w.WriteHeader(status)
	tmpl.Execute(w, v)
}

func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	v := loginView{Tab: "login"}
	if r.URL.Query().Get("registered") == "1" {
		v.Notice = "Account created — please sign in."
	}
	renderLogin(w, http.StatusOK, v)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	user, err := h.Users.GetByUsername(r.Context(), username)
	if err != nil || user == nil {
		renderLogin(w, http.StatusUnauthorized, loginView{Tab: "login", Username: username, Error: "Invalid username or password."})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		renderLogin(w, http.StatusUnauthorized, loginView{Tab: "login", Username: username, Error: "Invalid username or password."})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username, os.Getenv("JWT_SECRET"))
	if err != nil {
		renderLogin(w, http.StatusInternalServerError, loginView{Tab: "login", Username: username, Error: "Something went wrong. Please try again."})
		return
	}

	h.Sessions.Put(r.Context(), "jwt", token)
	h.Sessions.Put(r.Context(), "username", user.ID)
	http.Redirect(w, r, "/player", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.Sessions.Destroy(r.Context())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		renderLogin(w, http.StatusInternalServerError, loginView{Tab: "register", Username: username, Error: "Something went wrong. Please try again."})
		return
	}

	_, err = h.Users.Create(r.Context(), username, string(hashedPassword))
	if err != nil {
		renderLogin(w, http.StatusConflict, loginView{Tab: "register", Username: username, Error: "That username is already taken."})
		return
	}

	http.Redirect(w, r, "/login?registered=1", http.StatusSeeOther)
}
