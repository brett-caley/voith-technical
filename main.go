package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

var wikiCache Cache

// Just a fun test to see how we are doing that I left in
func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!")
}

// Return the intro
func handleIntro(w http.ResponseWriter, r *http.Request) {
	// Ensure JSON header is used
	w.Header().Set("Content-Type", "application/json")
	// Responsd with the intro
	json.NewEncoder(w).Encode(map[string]interface{}{"response": wikiCache.getIntro(), "status": http.StatusOK})
}

// Handle returning section content, queries: name, limit
func handleSection(w http.ResponseWriter, r *http.Request) {
	// Ensure JSON header is used
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	var response map[string]interface{} = make(map[string]interface{})

	// If we don't have vars, bad
	if vars == nil {
		response["status"] = http.StatusBadRequest
		json.NewEncoder(w).Encode(response)
		return
	}

	name, ok := vars["name"]

	// If we don't have a name, bad
	if ok == false {
		response["status"] = http.StatusBadRequest
		json.NewEncoder(w).Encode(response)
		return
	}

	// Again, if we don't have a name, bad
	if name == "" {
		response["status"] = http.StatusBadRequest
		json.NewEncoder(w).Encode(response)
		return
	}

	// Determine if there is a limit
	limit, ok := vars["limit"]

	// Gather the content
	sectionContent := wikiCache.getValue(name)

	// Only do the limit if it is specified
	if ok != false {
		limitInt, err := strconv.Atoi(limit)

		// If the limit is mangled, bad
		if err != nil || limitInt >= len(sectionContent) {
			response["response"] = "Invalid limit."
			response["status"] = http.StatusBadRequest
			json.NewEncoder(w).Encode(response)
			return
		}

		// Return sliced content
		if limitInt != 0 {
			sectionContent = sectionContent[:limitInt]
		}
	}

	response["response"] = sectionContent
	response["status"] = http.StatusOK
	json.NewEncoder(w).Encode(response)
}

// Handle the request for a list of  content
func handleList(w http.ResponseWriter, r *http.Request) {
	// Ensure JSON header is used
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	var response map[string]interface{} = make(map[string]interface{})

	// No vars, bad
	if vars == nil {
		response["status"] = http.StatusBadRequest
		json.NewEncoder(w).Encode(response)
		return
	}

	listType, ok := vars["type"]

	// No values, bad
	if ok == false {
		response["status"] = http.StatusBadRequest
		json.NewEncoder(w).Encode(response)
		return
	}

	// Can only handle sections at the moment
	if listType == "" || listType != "sections" {
		response["status"] = http.StatusBadRequest
		json.NewEncoder(w).Encode(response)
		return
	}

	// Retrieve sections
	sections := wikiCache.getSections()

	limit, ok := vars["limit"]

	// Handle limit similar to above
	if ok != false {
		limitInt, err := strconv.Atoi(limit)

		if err != nil || limitInt >= len(sections) {
			response["response"] = "Invalid limit."
			response["status"] = http.StatusBadRequest
			json.NewEncoder(w).Encode(response)
			return
		}
		// Set as a slice
		if limitInt != 0 {
			sections = sections[:limitInt]
		}
	}

	response["response"] = sections
	response["status"] = http.StatusOK

	json.NewEncoder(w).Encode(response)
}

// Handle the request for random section content
func handleRandom(w http.ResponseWriter, r *http.Request) {
	// Ensure JSON header is used
	w.Header().Set("Content-Type", "application/json")

	// Generate a map to put the details into
	var randomDetails map[string]string = make(map[string]string)

	// Gather the random content
	var title, content = wikiCache.getRandom()

	// Insert into the map
	randomDetails["title"] = title
	randomDetails["content"] = content

	// Send
	json.NewEncoder(w).
		Encode(map[string]interface{}{"response": randomDetails, "status": http.StatusOK})
}

func main() {
	// First step: initialize the data
	wikiCache.init()

	// Initialize the router
	router := mux.NewRouter().StrictSlash(true)

	// Good practice is to have a route to API versions
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	// Handle the home route
	subrouter.HandleFunc("/", handleHome)

	// Handle the intro route
	subrouter.HandleFunc("/intro", handleIntro)

	// Handle the section route
	subrouter.Path("/section").
		Queries("name", "{name}", "limit", "{limit}").
		HandlerFunc(handleSection)

	// Handle the lists route
	subrouter.Path("/lists").
		Queries("type", "{type}", "limit", "{limit}").
		HandlerFunc(handleList)

	// Handle the random route
	subrouter.HandleFunc("/random", handleRandom)

	// Start and serve
	log.Fatal(http.ListenAndServe(":8080", router))
}
