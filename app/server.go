package app

import (
	"clearskies/app/database"
	"clearskies/app/model"
	"clearskies/app/session"
	"clearskies/app/utils"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/context"
	"github.com/jmoiron/sqlx"
)

func ip(r *http.Request) string {
	return r.Header.Get("X-FORWARDED-FOR")
}

var db *sqlx.DB

func init() {
	db = database.Db
}

func Serve() {
	router := routes()
	serveMux := http.NewServeMux()
	thumb := regexp.MustCompile(`^/thumbnails/[a-zA-Z0-9]{5}$`)
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !thumb.MatchString(r.RequestURI) {
			log.Print(ip(r), ": ", r.Method, " ", r.RequestURI)
		}
		s := session.Get(r)
		if _, ok := s.Values["SignedIn"]; !ok {
			s.Values["SignedIn"] = false
			s.Values["Username"] = ""
		}
		if _, ok := s.Values["Verified"]; !ok {
			s.Values["Verified"] = false
		}
		if _, ok := s.Values["Admin"]; !ok {
			s.Values["Admin"] = false
		}
		user := model.User{}
		db.Get(&user, "SELECT * FROM users WHERE username = $1", s.Values["Username"])
		context.Set(r, "csrf", string(utils.DeriveExpiryCode("CSRF", 0, utils.FromHex(user.Key))))
		s.Options.HttpOnly = true
		s.Save(r, w)
		router.ServeHTTP(w, r)
	})
	server := http.Server{
		Addr:    "127.0.0.1:9090",
		Handler: serveMux,
	}
	log.Print("Listening on port ", "9090", "...")
	server.ListenAndServe()
}
