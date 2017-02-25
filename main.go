package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"path/filepath"

	"strings"

	"github.com/gorilla/sessions"
	"github.com/nu7hatch/gouuid"
)

var tpl *template.Template
var secretKey string

func init() {

	s, _ := uuid.NewV4()
	secretKey = s.String()
	tpl = template.Must(template.ParseGlob("./*.html"))
}

func main() {
	http.Handle("/assets/",
		http.StripPrefix("/assets",
			http.FileServer(http.Dir("./assets"))))

	http.HandleFunc("/", index)
	http.HandleFunc("/admin", admin)
	http.HandleFunc("/login", login)

	http.ListenAndServe(":8888", nil)
}

type IndexPage struct {
	Photos []string
}

func index(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "index.html", IndexPage{
		Photos: getPics(),
	})
	if err != nil {
		log.Println(err)
	}

	log.Println(r.URL.Path)
}

func getPics() []string {
	photos := make([]string, 0)
	filepath.Walk("assets/images", func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}

		path = strings.Replace(path, "\\", "/", -1)
		photos = append(photos, path)
		return nil
	})

	return photos
}

func login(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "blogSession")

	if r.Method == http.MethodPost {
		if r.FormValue("pass") == "sakamoto" && r.FormValue("mail") == "des@test.in" {
			session.Values["mail"] = r.FormValue("mail")
			session.Save(r, w)
			http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
		}
	}

	_, err := r.Cookie("blogSession")
	if err != nil {
		err := tpl.ExecuteTemplate(w, "login.html", nil)
		if err != nil {
			log.Println(err)
		}
	} else {
		http.Redirect(w, r, "/admin", 301)
	}

	log.Println(r.URL.Path)
}

func admin(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("blogSession")
	if err != nil {
		http.Redirect(w, r, "/login", 301)
	} else {

		if r.Method == http.MethodPost {
			uploadPage(w, r)
		}

		err := tpl.ExecuteTemplate(w, "admin.html", nil)
		if err != nil {
			log.Println(err)
		}
	}

	log.Println(r.URL.Path)
}

func uploadPage(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		// To recieve a file, for html its going to be input type="file" name="file"
		src, hdr, err := r.FormFile("blogFile")
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}
		defer src.Close()

		//writing file by creating one
		dst, err := os.Create("./assets/images/" + hdr.Filename)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}
		defer dst.Close()

		// copy the uploaded file
		_, err = io.Copy(dst, src)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
		} else {
			http.Redirect(w, r, "/", 301)
		}

		//fmt.Fprintf(w, `{"response":"Success"}`)
	}

	log.Println(r.URL.Path)
}

var store = sessions.NewCookieStore([]byte(secretKey))
