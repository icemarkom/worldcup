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

	_ "embed"
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
	stageRoundOfSixteen = "Round of 16"
	stageQuarterFinal   = "Quarter-final"
	stageSemiFinal      = "Semi-final"

	stageThirdPlace = "Play-off for third place"
	stageFinal      = "Final"

	wsjURL  = "https://worldcupjson.net/matches/"
	flagURL = "https://countryflagsapi.com/svg/"

	indexHTML = "template_index.html"

	flagFIFA     = "FIFA"
	flagWorldCup = "Qatar"
)

type Country struct {
	ID    string `json:"country"`
	Name  string `json:"name"`
	Flag  string
	Goals int `json:"goals"`
}

type Match struct {
	HomeTeam Country `json:"home_team"`
	AwayTeam Country `json:"away_team"`
	Time     string  `json:"time"`
	Updated  string
}

var (
	//go:embed template_index.html
	htmlFS embed.FS

	//go:embed flags/*.png
	flagsFS embed.FS

	// God save the King and his many domains. Also FIFA.
	FlagExceptions = map[string]string{
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
	match.Updated = time.Now().Format("2006-01-02 15:04:05 MST")

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &match); err != nil {
		return nil, err
	}

	// Flag hack: FIFA and ISO don't agree on 3-letter codes; or whether they should be 3-letter.
	for _, t := range []*Country{&match.HomeTeam, &match.AwayTeam} {
		t.Flag = t.ID
		if f, ok := FlagExceptions[t.ID]; ok {
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

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// Silently ignore favicons.
	if r.URL.Path == "/favicon.ico" {
		return
	}
	// The only URL we serve is an integer number between 49 and 64 (Round of 16 onwards).
	mn, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/"))
	if err != nil || (mn < 49 || mn > 64) {
		log.Printf("Bad URL request: %v", err)
		http.NotFound(w, r)
		return
	}

	match, err := fetchMatch(mn)
	if err != nil {
		log.Printf("Error fetching match %d: %v", mn, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tpl := template.Must(template.ParseFS(htmlFS, indexHTML))

	if err := tpl.Execute(w, match); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func EntryFunc(port string) {
	// HTTP server
	http.HandleFunc("/", handleRoot)
	http.Handle("/flags/", http.FileServer(http.FS(flagsFS)))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
