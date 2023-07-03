package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type GoLinks struct {
	mu    sync.RWMutex
	links map[string]string
}

var goLinks GoLinks

func init() {
	goLinks.links = make(map[string]string)
	loadLinks()
}

func main() {
	http.HandleFunc("/", handleRedirect)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/remove", handleRemove)
	http.HandleFunc("/modify", handleModify)

	log.Println("Starting go link server...")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func loadLinks() {
	data, err := ioutil.ReadFile("links.json")
	if err != nil {
		log.Println("Error reading links.json:", err)
		return
	}

	goLinks.mu.Lock()
	defer goLinks.mu.Unlock()

	if err := json.Unmarshal(data, &goLinks.links); err != nil {
		log.Println("Error decoding links.json:", err)
		return
	}
}

func saveLinks() {
	goLinks.mu.RLock()
	data, err := json.Marshal(goLinks.links)
	goLinks.mu.RUnlock()

	if err != nil {
		log.Println("Error encoding links:", err)
		return
	}

	if err := ioutil.WriteFile("links.json", data, 0644); err != nil {
		log.Println("Error writing links.json:", err)
	}
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	goLinks.mu.RLock()
	dest, ok := goLinks.links[r.URL.Path]
	goLinks.mu.RUnlock()

	if ok {
		http.Redirect(w, r, dest, http.StatusFound)
	} else {
		http.NotFound(w, r)
	}
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	if key == "" || value == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	goLinks.mu.Lock()
	goLinks.links["/"+key] = value
	goLinks.mu.Unlock()

	saveLinks()

	http.Redirect(w, r, "/"+key, http.StatusFound)
}

func handleRemove(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if key == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	goLinks.mu.Lock()
	delete(goLinks.links, "/"+key)
	goLinks.mu.Unlock()

	saveLinks()

	http.Redirect(w, r, "/"+key, http.StatusFound)
}

func handleModify(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	newValue := r.URL.Query().Get("value")

	if key == "" || newValue == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	goLinks.mu.Lock()
	goLinks.links["/"+key] = newValue
	goLinks.mu.Unlock()

	saveLinks()

	http.Redirect(w, r, "/"+key, http.StatusFound)
}
