package main

import (
	"bytes"
	"encoding/json"
	"flag"
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

	"github.com/russross/blackfriday"
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
	Owner        string `json:"owner"`
	Repo         string `json:"repo"`
	PerPage      int    `json:"per_page"`
	State        string `json:"state"`
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
	usage()
	if config.PerPage > 100 || config.PerPage < 0 {
		config.PerPage = 100
	}
	if config.State != "all" && config.State != "open" && config.State != "closed" {
		config.State = "all"
	}
}

func main() {

	if config.Owner == "" || config.Repo == "" {
		fmt.Println(helpMessages)
		fmt.Println()
		return
	}

	issuesDir = config.Owner + "_" + config.Repo + "_issues"
	if err := os.Mkdir(issuesDir, 0755); os.IsExist(err) {
		log.Printf("dir %s already exists\n", issuesDir)
	} else {
		log.Printf("create dir %s successfully\n", issuesDir)
	}

	var issues []Issue
	var page = 1

	for {
		resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?page=%d&per_page=%d&state=%s&client_id=%s&client_secret=%s", config.Owner, config.Repo, page, config.PerPage, config.State, config.ClientID, config.ClientSecret))
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
			log.Printf("github API rate limit. The application will retry automatically after %d minutes. Or put in your clinet_id and client_secret to config file to increase the API rate. Refer to https://developer.github.com/v3/#rate-limiting\n\n", resetm)
			timer(resetm)
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
	var buff bytes.Buffer
	for index, issue := range issues {
		title := fmt.Sprintf("%s_%s_%s_#%d.html", issue.CreatedAt[:10], issue.Title, issue.State, issue.Number)
		title = remove(title)
		log.Printf("hit %d: %s\n", index, title)
		buff.WriteString(fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td><a href='%s' target='_blank'>#%d</td></tr>", index, issue.CreatedAt[:10], issue.Title, string(blackfriday.MarkdownBasic([]byte(issue.Body))), issue.State, urlEncode(title), issue.Number))
		resp, err := http.Get(issue.HTMLURL)
		if err != nil {
			log.Println("get http response body error: ", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("get http response body error: ", err)
		}
		if err := ioutil.WriteFile(filepath.Join(issuesDir, title), body, 0755); err != nil {
			log.Println("error: ", err)
		}
	}
	generateIndexHTML(buff.String())
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

func urlEncode(s string) string {
	return strings.NewReplacer("!", "%21", "#", "%23", "$", "%24", "&", "%26", "'", "%27", "(", "%28", ")", "%29", "*", "%2A", "+", "%2B", ",", "%2C", "/", "%2F", ":", "%3A", ";", "%3B", "=", "%3D", "?", "%3F", "@", "%40", "[", "%5B", "]", "%5D").Replace(s)
}

func timer(t int) {
	for t1 := t - 1; t1 >= 0; t1-- {
		for t2 := 59; t2 >= 0; t2-- {
			fmt.Printf("\rretry automatically after %2dm%2ds", t1, t2)
			time.Sleep(time.Second * 1)
		}
	}
}

func generateIndexHTML(c string) {
	html := `<!DOCTYPE html>
	<html>

	<head>
	    <meta charset="utf-8">
	    <meta name="viewport" content="width=device-width">
	    <title>Index</title>
	    <style>
	        .container {
	            margin: 0 auto;
	        }

	        table {
	            border-collapse: collapse;
	            border-spacing: 0;
	            margin: auto;
				table-layout:fixed;
				width: 90%;
	            max-width: 90%;
	        }

	        table,
	        th,
	        td {
	            border: 1px solid #ddd;
	            text-align: center;
				word-wrap:break-word;
				word-break:break-all;
	        }

	        th,
	        td {
	            padding-top: 10px;
	            padding-bottom: 10px;
	        }

			img {
				width: 50%;
			}
	    </style>
	</head>

	<body>
	    <div class="container">
	        <table>
	            <tr>
	                <th width="40px">Index</th>
	                <th width="100px">Created_At</th>
	                <th width="20%">Title</th>
	                <th width="80%">Body</th>
	                <th width="100px">State</th>
	                <th width="100px">#issue</th>
	            </tr>` + c + ` </table>
	 	    </div>
	 	</body>

	 	</html>`
	if err := ioutil.WriteFile(filepath.Join(issuesDir, "index.html"), []byte(html), 0755); err != nil {
		log.Println("generate index.html error: ", err)
	}
}

func usage() {
	flag.Usage = func() {
		fmt.Println(helpMessages)
	}
	var (
		owner        string
		repo         string
		perPage      int
		state        string
		clientID     string
		clientSecret string
	)

	flag.StringVar(&owner, "o", config.Owner, "github owner of repesitory")
	flag.StringVar(&owner, "owner", config.Owner, "github owner of repesitory")
	flag.StringVar(&repo, "r", config.Repo, "github repesitory")
	flag.StringVar(&repo, "repo", config.Repo, "github repesitory")
	flag.IntVar(&perPage, "p", config.PerPage, "pagination, page size up to 100")
	flag.IntVar(&perPage, "per_page", config.PerPage, "pagination, page size up to 100")
	flag.StringVar(&state, "s", config.State, "issues state (open, closed or all)")
	flag.StringVar(&state, "state", config.State, "issues state (open, closed or all)")
	flag.StringVar(&clientID, "ci", config.ClientID, "github OAuth application's client ID")
	flag.StringVar(&clientID, "client_id", config.ClientID, "github OAuth application's client ID")
	flag.StringVar(&clientSecret, "cs", config.ClientSecret, "github OAuth application's client Secret")
	flag.StringVar(&clientSecret, "client_secret", config.ClientSecret, "github OAuth application's client Secret")

	flag.Set("logtostderr", "true")
	flag.Parse()

	config.Owner = owner
	config.Repo = repo
	config.PerPage = perPage
	config.State = state
	config.ClientID = clientID
	config.ClientSecret = clientSecret
}

var helpMessages = `
Usage: export-github-issues [COMMANDS] [VARS]

SUPPORT COMMANDS:
    -h, --help               help messages

SUPPORT VARS:
    -o, --owner              github owner of repesitory
    -r, --repo               github repesitory
    -p, --per_page           pagination, page size up to 100
    -s, --state              issues state (open, closed or all)
    -ci, --client_id         github OAuth application's client ID
    -cs, --client_secret     github OAuth application's client Secret
		`
