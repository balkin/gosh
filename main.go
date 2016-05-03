package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/syndtr/goleveldb/leveldb"
	"net/http"
	"log"
	"fmt"
	"bytes"
)

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
	fmt.Println(w, "SHORT")
}

func Expand(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	http.Redirect(w, r, "http://baron.su/", 302)
}

func Link(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	fmt.Println(w, "SHORT")
}

var LastKey = 0

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
		log.Printf("i=%d, char=%c, p=%d, mc=%d", i, str[i], p, MaxChars)
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
	defer db.Close()
	iter := db.NewIterator(nil, nil)
	iter.Last()
	LastKey = ShortToNumeric(string(iter.Key()))
	iter.Release()
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/short.go", Shorten)
	router.PUT("/:link", Link)
	router.GET("/:link", Expand)
	err = http.ListenAndServe(":9000", router)
	log.Fatal("Error: ", err)
}
