// The cache file of the technical module intends to provide a resource to
// hold some data to avoid calling the wiki api every reference
package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
)

// A constant that refers to Wikipedia's ID for the page
const heidenheimPageID = "1333570"

// A constant that refers to the specific API call for the page information
const wikiRefURL = "http://en.wikipedia.org/w/api.php?action=query&pageids=1333570&prop=extracts&explaintext=0&format=json"

// A warning message from the extract
type wikiWarningExtracts struct {
	Message string `json:"*"`
}

// A warning property
type wikiWarnings struct {
	Extracts wikiWarningExtracts `json:"extracts"`
}

// The Wikipedia Page's property
type wikiPage struct {
	Pageid  int    `json:"pageid"`
	Ns      int    `json:"ns"`
	Title   string `json:"title"`
	Extract string `json:"extract"`
}

// The Wikipedia Page is returned as a property that is the page ID
type wikiPages struct {
	HeidenheimPage wikiPage `json:"1333570"`
}

// The results of the wikipedia query
type wikiQuery struct {
	Pages wikiPages `json:"pages"`
}

// The full response from Wikipedia's API page
type wikiResponse struct {
	Batchcomplete string       `json:"batchcomplete"`
	Warnings      wikiWarnings `json:"warnings"`
	Query         wikiQuery    `json:"query"`
}

// A typedef for a section, which is a map from title to content
type section map[string]string

// Cache structure will hold all the information regarding Heidenheim from
// wikipedia
type Cache struct {
	Title         string
	Intro         string
	Sections      section
	SectionTitles []string
}

// Initialize the cache with default wikipedia information
func (cache *Cache) init() {
	cache.Sections = make(section)

	data := getWikiInformation(wikiRefURL)

	cache.extractWikiInformation(data)
}

// Get the wikipedia information from the MediaWiki API
func getWikiInformation(url string) wikiResponse {
	// Run the http get
	wikiInformation, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	// Close this when done
	defer wikiInformation.Body.Close()

	// Create a wikiResponse object to hold the decoded JSON information
	var wikiJSON wikiResponse

	// Decode wikiInfo
	err = json.NewDecoder(wikiInformation.Body).Decode(&wikiJSON)

	if err != nil {
		log.Fatal(err)
	}

	return wikiJSON
}

// Extract the wikipedia information from a large block of text
func (cache *Cache) extractWikiInformation(wikiInfo wikiResponse) {
	// Fortunately the title is returned by Wikipedia
	cache.Title = wikiInfo.Query.Pages.HeidenheimPage.Title

	// Create the regex for extracting section titles and content
	regex := regexp.MustCompile(`(?:(?:={2,})([\s\S]*?)={2,})?\n?([\s\S]*?)\n{2,}`)

	// Run the regex
	extractedSections := regex.FindAllSubmatch([]byte(wikiInfo.Query.Pages.HeidenheimPage.Extract), -1)

	// Update the cache intro
	cache.Intro = string(extractedSections[0][2])

	// For each section, create its title as the key, and the content as the value
	for i := 1; i < len(extractedSections)-1; i++ {
		sectionTitle := strings.TrimSpace(string(extractedSections[i][1]))
		sectionContent := string(extractedSections[i][2])

		cache.Sections[sectionTitle] = sectionContent

		// Discard all section titles that are empty, but append the rest to
		// an array of titles.
		if sectionTitle != "" {
			cache.SectionTitles = append(cache.SectionTitles, sectionTitle)
		}
	}
	return
}

// Return the content of a section
func (cache Cache) getValue(key string) string {
	return cache.Sections[key]
}

// Return the page title
func (cache Cache) getTitle() string {
	return cache.Title
}

// Return the introduction to the page
func (cache Cache) getIntro() string {
	return cache.Intro
}

// Return the array of sections
func (cache Cache) getSections() []string {
	return cache.SectionTitles
}

// Return a random section and its contents
func (cache Cache) getRandom() (string, string) {
	randomIndex := rand.Int() % (len(cache.SectionTitles) - 1)
	randomSectionTitle := cache.SectionTitles[randomIndex]
	randomSection := cache.Sections[randomSectionTitle]
	return randomSectionTitle, randomSection
}
