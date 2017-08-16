package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//Issue struct
type Issue struct {
	URL           string `json:"url"`
	RepositoryURL string `json:"repository_url"`
	LabelsURL     string `json:"labels_url"`
	CommentsURL   string `json:"comments_url"`
	EventsURL     string `json:"events_url"`
	HTMLURL       string `json:"html_url"`
	ID            int    `json:"id"`
	Number        int    `json:"number"`
	Title         string `json:"title"`
	State         string `json:"state"`
	Comments      int    `json:"comments"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	Body          string `json:"body"`
}

//Config struct
type Config struct {
	Author       string `json:"author"`
	Repo         string `json:"repo"`
	PerPage      int    `json:"per_page"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

//APILimit struct
type APILimit struct {
	Message string `json:"message"`
	DocURL  string `json:"documentation_url"`
}

const (
	configFilename string = "config.json"
)

var config Config
var issuesDir string

func init() {
	parseConfig()
	issuesDir = config.Author + "_" + config.Repo + "_issues"
	if err := os.Mkdir(issuesDir, 0755); os.IsExist(err) {
		log.Printf("dir %s already exists\n", issuesDir)
	} else {
		log.Printf("create dir %s successfully\n", issuesDir)
	}
}

func main() {

	var issues []Issue
	var page = 1

	for {
		resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?page=%d&per_page=%d&state=all&client_id=%s&client_secret=%s", config.Author, config.Repo, page, config.PerPage, config.ClientID, config.ClientSecret))
		if err != nil {
			log.Println("http.get error: ", err)
		}
		defer resp.Body.Close()

		remaining := resp.Header["X-Ratelimit-Remaining"][0]
		t, err := strconv.ParseInt(resp.Header["X-Ratelimit-Reset"][0], 10, 64)
		if err != nil {
			log.Println("parse time error: ", err)
		}
		resett := time.Unix(t, 0)
		resetm := int(math.Ceil(resett.Sub(time.Now()).Minutes())) + 5
		if remaining == "0" {
			log.Printf("github API 已达到次数限制，将在%d分钟后自动重试. 或者将自己的clinet_id和client_secret填入到配置文件以增加API次数. 参见 https://developer.github.com/v3/#rate-limiting\n", resetm)
			time.Sleep(time.Minute * time.Duration(resetm))
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("http.get response.body error: ", err)
		}
		var iss []Issue
		if err := json.Unmarshal(body, &iss); err != nil {
			log.Println("parse response.body error: ", err)
		}
		if len(iss) == 0 {
			break
		}
		issues = append(issues, iss...)
		page++
		log.Printf("hit %d issues\n", len(iss))
	}
	if len(issues) == 0 {
		log.Println("no issues found.")
	}
	for index, issue := range issues {
		title := fmt.Sprintf("%s_%s_%s_#%d.html", issue.CreatedAt[:10], issue.Title, issue.State, issue.Number)
		log.Printf("hit %d: %s\n", index, title)
		resp, err := http.Get(issue.HTMLURL)
		if err != nil {
			log.Println("get http response body error: ", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("get http response body error: ", err)
		}
		if err := ioutil.WriteFile(filepath.Join(issuesDir, remove(title)), body, 0755); err != nil {
			log.Println("error: ", err)
		}
	}
	fmt.Println("\n\nPress Enter to continue...")
	fmt.Scanln()
}

//Parse config file
func parseConfig() {
	data, err := ioutil.ReadFile(configFilename)
	if err != nil {
		log.Fatalln("read config file error: ", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalln("parse config file error: ", err)
	}
}

func remove(s string) string {
	return strings.NewReplacer("/", "", "\\", "", ":", "", "*", "", "<", "", ">", "", "|", "", "\"", "", "?", "").Replace(s)
}
