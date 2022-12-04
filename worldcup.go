package worldcup

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/*
	Match IDs used by WCJ:

		Round of 16:

			- NED-USA: 49
			- ARG-AUS: 50
			- FRA-POL: 51
			- ENG-SEN: 52
			- JAP-CRO: 53
			- BRA-KOR: 54
			- MOR-ESP: 55
			- POR-SUI: 56

		Quarter-final:
			- 53-54: 57
			- 49-50: 58
			- 55-56: 59
			- 51-52: 60

		Semi-final:
			- 57-58: 61
			- 59-60: 62

		Third place:
			- L61-L62: 63

		Final:
			- 61-62: 64
*/

const (
	stageFirstStage     = "First stage"
	stageRoundOfSixteen = "Round of 16"
	stageQuarterFinal   = "Quarter-final"
	stageSemiFinal      = "Semi-final"
	stageThirdPlace     = "Play-off for third place"
	stageFinal          = "Final"
	countryTBD          = "To Be Determined"
	timeFullTime        = "full-time"

	wsjURL  = "https://worldcupjson.net/matches/"
	flagURL = "https://countryflagsapi.com/svg/"

	indexHTML = "template_index.html"
	matchHTML = "template_match.html"

	flagFIFA     = "FIFA"
	flagWorldCup = "Qatar"

	paramNoSpoilers = "boring"
	localTZ         = "America/Los_Angeles"
)

// Country is information about a side in a match.
type Country struct {
	ID    string `json:"country"`
	Name  string `json:"name"`
	Flag  string
	Goals int `json:"goals"`
}

// Match is the information about a match.
type Match struct {
	ID         int     `json:"id"`
	Stage      string  `json:"stage_name"`
	HomeTeam   Country `json:"home_team"`
	AwayTeam   Country `json:"away_team"`
	Time       string  `json:"time"`
	Updated    string
	NoSpoilers bool
}

var (
	//go:embed *.html
	htmlFS embed.FS

	//go:embed flags/*.png
	flagsFS embed.FS

	// flagExceptions: God save the King and his many domains. Also FIFA.
	flagExceptions = map[string]string{
		"CRO": "HRV",
		"ENG": "GB-ENG",
		"NED": "NLD",
		"POR": "PRT",
		"SUI": "CHE",
	}
)

func fetchMatch(m int) (*Match, error) {
	var match Match

	resp, err := http.Get(fmt.Sprintf("%s/%d", wsjURL, m))
	if err != nil {
		return nil, err
	}
	tz, err := time.LoadLocation(localTZ)
	if err != nil {
		return nil, err
	}
	match.Updated = time.Now().In(tz).Format("2006-01-02 15:04:05 MST")

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &match); err != nil {
		return nil, err
	}

	// Flag hack: FIFA and ISO don't agree on 3-letter codes; or whether they should be 3-letter.
	for _, t := range []*Country{&match.HomeTeam, &match.AwayTeam} {
		t.Flag = t.ID
		if f, ok := flagExceptions[t.ID]; ok {
			t.Flag = f
		}
	}
	if match.HomeTeam.Name == "To Be Determined" {
		match.HomeTeam.Flag = flagFIFA
	}
	if match.AwayTeam.Name == "To Be Determined" {
		match.AwayTeam.Flag = flagWorldCup
	}
	return &match, nil
}

func fetchAllMatches() ([]Match, error) {
	var matches []Match

	resp, err := http.Get(fmt.Sprintf("%s", wsjURL))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &matches); err != nil {
		return nil, err
	}
	return matches, nil
}

func isTBD(c Country) string {
	if c.Name == countryTBD {
		return c.ID
	}
	return c.Name
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	matches, err := fetchAllMatches()
	if err != nil {
		log.Printf("Error fetching matches: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmap := template.FuncMap{
		"title16": func() string {
			return stageRoundOfSixteen
		},
		"titleQF": func() string {
			return stageQuarterFinal
		},
		"titleSF": func() string {
			return stageSemiFinal
		},
		"title3P": func() string {
			return stageThirdPlace
		},
		"titleF": func() string {
			return stageFinal
		},
		"noSpoilers": func() string {
			return paramNoSpoilers
		},
		"is16": func(m Match) bool {
			return m.Stage == stageRoundOfSixteen
		},
		"isQF": func(m Match) bool {
			return m.Stage == stageQuarterFinal
		},
		"isSF": func(m Match) bool {
			return m.Stage == stageSemiFinal
		},
		"is3P": func(m Match) bool {
			return m.Stage == stageThirdPlace
		},
		"isF": func(m Match) bool {
			return m.Stage == stageFinal
		},
		"isTBD": isTBD,
	}
	tpl := template.Must(template.New(indexHTML).Funcs(fmap).ParseFS(htmlFS, indexHTML))

	if err := tpl.Execute(w, matches); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func handleMatch(w http.ResponseWriter, r *http.Request, mn int) {
	match, err := fetchMatch(mn)
	if err != nil {
		log.Printf("Error fetching match %d: %v", mn, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	if _, ok := r.Form[paramNoSpoilers]; ok {
		match.NoSpoilers = true
	}

	fmap := template.FuncMap{
		"noSpoilers": func(m Match) bool {
			return match.NoSpoilers
		},
		"boringMatch": func(m Match) bool {
			return m.Time == timeFullTime && (m.HomeTeam.Goals+m.AwayTeam.Goals == 0)
		},
		"isTBD": isTBD,
	}

	tpl := template.Must(template.New(matchHTML).Funcs(fmap).ParseFS(htmlFS, matchHTML))

	if err := tpl.Execute(w, match); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Root should show us the list of matches.
	if r.URL.Path == "/" {
		handleRoot(w, r)
		return
	}

	// Silently ignore favicons.
	if r.URL.Path == "/favicon.ico" {
		return
	}

	// The only URL we serve is an integer number between 49 and 64 (Round of 16 onwards).
	mn, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/"))
	if err == nil && mn >= 49 && mn <= 64 {
		handleMatch(w, r, mn)
		return
	}

	// Everything else is 404d.
	log.Printf("Bad URL request: %v", err)
	http.NotFound(w, r)
	return
}

// EntryFunc is the entry function to the app; separated from the main to facilitate Cloud Function deployment.
func EntryFunc(port string) {
	// HTTP server
	http.HandleFunc("/", handleRequest)
	http.Handle("/flags/", http.FileServer(http.FS(flagsFS)))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
