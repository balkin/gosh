package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/syndtr/goleveldb/leveldb"
	"net/http"
	"log"
	"fmt"
	"bytes"
	"encoding/json"
	"math"
)

type url_struct struct {
	Url string
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintln(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" integrity="sha384-1q8mTJOASx8j1Au+a5WDVnPi2lkFfwwEAa8hDDdjZlpLegxhjVME1fgjWPGmkzs7" crossorigin="anonymous">
<title>GOSH</title>
<style>.container {padding:50px}</style>
</head>
<body>
<nav class="navbar navbar-inverse navbar-fixed-top"><div class="navbar-header">
          <a class="navbar-brand" href="https://github.com/balkin/gosh">GOSH</a>
</div></nav>
<div class="container">
<h1>Powered by GOSH</h1>
<p>This website is using GOSH shortening engine, check out the <a href="https://github.com/balkin/gosh">Github project</a>.</p>
</div></body>
</html>`)
}

func Shorten(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var url string
	url = r.FormValue("url")
	if url == "" {
		url = r.URL.Query().Get("url")
	}
	if url == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	var link string
	for LastKey < math.MaxInt32 {
		LastKey++
		link = NumericToShort(LastKey)
		exists, err := DB.Has([]byte(link), nil)
		if !exists && err == nil {
			break
		}
	}
	err := DB.Put([]byte(link), []byte(url), nil)
	log.Printf("Link: %d, %s to %s", LastKey, link, url)
	if err != nil {
		http.Error(w, "DB Error", http.StatusBadGateway)
		return
	}
	fmt.Fprintln(w, "OK")
}

func Expand(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	link := params.ByName("link")
	data, err := DB.Get([]byte (link), nil)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	us := url_struct{Url:string(data)}
	json.NewEncoder(w).Encode(us)
	url := string(data)
	log.Printf("Redirection: %s to %s", link, url)
	http.Redirect(w, r, url, 302)
}

func JsonLink(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	link := params.ByName("link")
	data, err := DB.Get([]byte (link), nil)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	url := string(data)
	fmt.Fprintf(w, `{"url": "%s"}`, url)
}

func PutLink(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	link := params.ByName("link")
	decoder := json.NewDecoder(r.Body)
	var us url_struct
	err := decoder.Decode(&us)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	err = DB.Put([]byte(link), []byte(us.Url), nil)
	log.Printf("Link: %s to %s", link, us.Url)
	if err != nil {
		http.Error(w, "DB Error", http.StatusBadGateway)
		return
	}
	fmt.Fprintln(w, `{"ok": "true"}`)
}

var LastKey int = 0
var DB *leveldb.DB = nil

var ShortChars = []byte{'-',
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
	'k', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u',
	'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E',
	'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'P', 'Q',
	'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '0',
	'1', '2', '3', '4', '5', '6', '7', '8', '9',
}
var MaxChars = len(ShortChars) - 1

func ShortToNumeric(str string) int {
	v, m := 0, 1
	for i := len(str) - 1; i >= 0; i-- {
		p := bytes.IndexByte(ShortChars, str[i])
		v += p * m
		m *= MaxChars
	}
	return v
}

func NumericToShort(i int) string {
	if i < MaxChars {
		return string(ShortChars[i])
	}
	s := ""
	for i > MaxChars {
		r := i % MaxChars
		i = (i - r) / MaxChars
		s = string(ShortChars[r]) + s
	}
	return string(ShortChars[i]) + s
}

func main() {
	log.Println("Starting GOSH")
	db, err := leveldb.OpenFile("leveldb", nil)
	if err != nil {
		log.Fatal("Error: ", err)
		return
	}
	DB = db
	defer db.Close()
	iter := db.NewIterator(nil, nil)
	iter.Last()
	LastKey = ShortToNumeric(string(iter.Key()))
	iter.Release()
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/short.go", Shorten)
	router.PUT("/:link", PutLink)
	router.GET("/:link", Expand)
	router.GET("/:link/json", JsonLink)
	err = http.ListenAndServe(":9000", router)
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
