package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
)

func main() {
	// Allow user to specify listen port on command line
	var port int
	flag.IntVar(&port, "port", 8080, "port to listen on")
	flag.Parse()

	// Create in-memory database and add a couple of test albums
	db := NewMemoryDatabase()
	db.AddAlbum(Album{ID: "a1", Title: "9th Symphony", Artist: "Beethoven", Price: 795})
	db.AddAlbum(Album{ID: "a2", Title: "Hey Jude", Artist: "The Beatles", Price: 2000})

	// Create server and wire up database
	server := NewServer(db, log.Default())

	log.Printf("listening on http://localhost:%d", port)
	http.ListenAndServe(":"+strconv.Itoa(port), server)
}
