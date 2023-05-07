package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/exp/slices"
)

type Plannable struct {
	ContextType     string `json:"context_type"`
	CourseID        int    `json:"course_id"`
	PlannableID     int    `json:"plannable_id"`
	PlannerOverride struct {
		ID             int       `json:"id"`
		PlannableType  string    `json:"plannable_type"`
		PlannableID    int       `json:"plannable_id"`
		UserID         int       `json:"user_id"`
		WorkflowState  string    `json:"workflow_state"`
		MarkedComplete bool      `json:"marked_complete"`
		DeletedAt      time.Time `json:"deleted_at"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		Dismissed      bool      `json:"dismissed"`
		AssignmentID   int       `json:"assignme5nt_id"`
	} `json:"planner_override"`
	PlannableType string      `json:"plannable_type"`
	NewActivity   bool        `json:"new_activity"`
	Submissions   interface{} `json:"submissions"`
	PlannableDate time.Time   `json:"plannable_date"`
	Plannable     struct {
		ID          int       `json:"id"`
		Title       string    `json:"title"`
		UnreadCount int       `json:"unread_count"`
		ReadState   string    `json:"read_state"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	} `json:"plannable"`
	HTMLURL      string `json:"html_url"`
	ContextName  string `json:"context_name"`
	ContextImage string `json:"context_image"`
}

type Params struct {
	key   string
	value string
}

type Response struct {
	plannables   []Plannable
	url_response map[string]string
}

func main() {
	// Get Cavnas Instructure API key from the env.
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file.")
	}
	CANVAS_ACCESS_TOKEN := os.Getenv("CANVAS_ACCESS_TOKEN")

	// Build planner URL for query.
	queryParams := [4]Params{
		{"start_date", ""},
		{"filter", ""},
		{"order", "desc"},
		{"per_page", "10"},
	}
	url := "https://uc.instructure.com/api/v1/planner/items?access_token="
	url += CANVAS_ACCESS_TOKEN
	for _, param := range queryParams {
		url += "&" + param.key + "=" + param.value
	}

	numPlannables := 25
	targetPlannableTypes := []string{"assignment", "quiz"}
	plannables := []Plannable{}
	i := 0
	for i < numPlannables {
		// Get plannables.
		response := GetPlannablesByType(url, targetPlannableTypes)
		plannables = append(plannables, response.plannables...)
		// Change the URL to the "next" URL from the return headers.
		url = response.url_response["next"]
		// Set i to the number of planables.
		i = len(plannables)
	}
	plannables = plannables[0:numPlannables]

	for _, plannable := range plannables {
		fmt.Println(plannable.Plannable.Title)
	}
}

func GetPlannables(url string) Response {
	// Create GET request.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	// Execute GET request.
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	// Close response body.
	defer res.Body.Close()
	// io.ReadAll from the response body.
	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		fmt.Print(err.Error())
	}
	// JSON unmarshal into a list of plannables.
	var plannables []Plannable
	err = json.Unmarshal(body, &plannables)
	if err != nil {
		panic(err.Error())
	}

	// Get the "Link" attributes from the response header.
	links := ParseLinkHeader(res.Header.Get("Link"))

	// Return response.
	return Response{plannables, links}
}

func ParseLinkHeader(linkHeader string) map[string]string {
	// We need to grab and append the access token from the OS again,
	// as the links returned in the header from the request
	// do not include the token return.
	CANVAS_ACCESS_TOKEN := os.Getenv("CANVAS_ACCESS_TOKEN")

	links := make(map[string]string)
	// Split the Link header into individual links.
	linkParts := strings.Split(linkHeader, ",")
	for _, part := range linkParts {
		// Split each link into URL and relationship type.
		urlAndRel := strings.Split(part, ";")
		// Extract the URL and relationship type.
		url := strings.Trim(urlAndRel[0], " <>")
		rel := strings.Trim(urlAndRel[1], " ")
		// Remove the surrounding quotes from the relationship type.
		rel = strings.Trim(rel, "rel=\"")
		rel = strings.Trim(rel, "\"")
		// Append the access token to the end of each URL, for future requests.
		url += "&access_token=" + CANVAS_ACCESS_TOKEN
		// Add the URL and relationship type to the map.
		links[rel] = url
	}

	return links
}

func GetPlannablesByType(url string, plannableType []string) Response {
	response := GetPlannables(url)
	typedPlannables := []Plannable{}
	// Add plannable to new typedPlannables slice if type matches.
	for _, plannable := range response.plannables {
		if slices.Contains(plannableType, plannable.PlannableType) {
			typedPlannables = append(typedPlannables, plannable)
		}
	}
	// Return modified list of plannables and the original url response.
	return Response{typedPlannables, response.url_response}
}
