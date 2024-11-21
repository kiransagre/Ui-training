package fixlets

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Fixlet struct {
	SiteID                string
	FixletID              string
	Name                  string
	Criticality           string
	RelevantComputerCount string
}

type ResFixlet struct {
	SiteID                int
	FixletID              int
	Name                  string
	Criticality           string
	RelevantComputerCount int
	SourceReleaseDate     string
	CriticalityVal        int
}

type Aggrgate struct {
	CriticalityAggregate [][]interface{}
	MonthAggregate       [][]interface{}
}

func GetFixlets(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetFixlets")
	file, err := os.Open("fixlets.csv")

	if err != nil {
		fmt.Println("error on opening file")
		io.WriteString(w, "error while reading file : "+err.Error())
		return
	}

	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()

	if err != nil {
		fmt.Println("filed to return records")
		io.WriteString(w, "Failed to read file")
		return
	}

	var fixltes []Fixlet

	for _, record := range records {
		var fixlet = Fixlet{
			SiteID:                record[0],
			FixletID:              record[1],
			Name:                  record[2],
			Criticality:           record[3],
			RelevantComputerCount: record[4],
		}

		fixltes = append(fixltes, fixlet)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fixltes)
}

func Fixletsfromserver(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Fixletsfromserver")
	queryStr := r.URL.Query().Get("criticality")
	client := http.Client{}
	reqBody, err := json.Marshal(map[string]string{
		"query": "select JSON_OBJECT('SiteID', siteID, 'FixletID', id, 'Name', Title, 'Criticality', severity, 'RelevantComputerCount', count_computers(Relevant), 'SourceReleaseDate', SourceReleaseDate) from FIXLETS where siteID = 2 and Title like 'MS24-%'",
	})

	if err != nil {
		fmt.Println("Error while marshel", err.Error())
		return
	}

	req, err := http.NewRequest("POST", "https://bigfix-server.sbx0228.play.hclsofy.dev/api/query", bytes.NewBuffer(reqBody))

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	req.SetBasicAuth("admin", "NcweJrthQZdx58r")
	req.Header.Set("Content-Type", "application/json+sql")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var fixltes []ResFixlet
	err = json.Unmarshal(body, &fixltes)
	if err != nil {
		fmt.Println("error in unmarshan")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	var resFixlets []ResFixlet
	for _, record := range fixltes {
		var val int
		switch record.Criticality {
		case "Critical":
			val = 1
		case "Important":
			val = 2
		case "Moderate":
			val = 3
		case "Medium":
			val = 4
		case "Low":
			val = 5
		case "Unspecified":
			val = 6
		default:
			val = 0
		}

		if queryStr != "" {
			if queryStr == record.Criticality {
				var fixlet = ResFixlet{
					SiteID:                record.SiteID,
					FixletID:              record.FixletID,
					Name:                  record.Name,
					CriticalityVal:        val,
					Criticality:           record.Criticality,
					RelevantComputerCount: record.RelevantComputerCount,
					SourceReleaseDate:     record.SourceReleaseDate,
				}
				resFixlets = append(resFixlets, fixlet)
			}
		} else {
			var fixlet = ResFixlet{
				SiteID:                record.SiteID,
				FixletID:              record.FixletID,
				Name:                  record.Name,
				CriticalityVal:        val,
				Criticality:           record.Criticality,
				RelevantComputerCount: record.RelevantComputerCount,
				SourceReleaseDate:     record.SourceReleaseDate,
			}
			resFixlets = append(resFixlets, fixlet)
		}
	}

	json.NewEncoder(w).Encode(resFixlets)

}

func Fixletstats(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Fixletstats")
	client := http.Client{}
	reqBody, err := json.Marshal(map[string]string{
		"query": "select JSON_OBJECT('SiteID', siteID, 'FixletID', id, 'Name', Title, 'Criticality', severity, 'RelevantComputerCount', count_computers(Relevant), 'SourceReleaseDate', SourceReleaseDate) from FIXLETS where siteID = 2 and Title like 'MS24-%'",
	})

	if err != nil {
		fmt.Println("Error while marshel", err.Error())
		return
	}

	req, err := http.NewRequest("POST", "https://bigfix-server.sbx0228.play.hclsofy.dev/api/query", bytes.NewBuffer(reqBody))

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	req.SetBasicAuth("admin", "NcweJrthQZdx58r")
	req.Header.Set("Content-Type", "application/json+sql")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var fixltes []ResFixlet
	err = json.Unmarshal(body, &fixltes)
	if err != nil {
		fmt.Println("error in unmarshan")
		return
	}

	criticalCount := map[string]int{}
	sourceReleaseDateCount := map[string]int{}
	for _, record := range fixltes {
		criticalCount[record.Criticality]++

		formattedDate, err := time.Parse("2006-01-02", record.SourceReleaseDate)

		if err != nil {
			fmt.Println("error while parsing date")
			return
		}

		year, month := formattedDate.Year(), formattedDate.Month()
		yearMonth := fmt.Sprintf("%d %s", year, month)
		sourceReleaseDateCount[yearMonth]++

	}

	var criticalityAggregate [][]interface{}
	var sourceReleaseDate [][]interface{}

	for key, count := range criticalCount {
		criticalityAggregate = append(criticalityAggregate, []interface{}{key, count})
	}

	for key, count := range sourceReleaseDateCount {
		sourceReleaseDate = append(sourceReleaseDate, []interface{}{key, count})
	}

	var agg = Aggrgate{
		CriticalityAggregate: criticalityAggregate,
		MonthAggregate:       sourceReleaseDate,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(agg)
}
