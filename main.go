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
	"regexp"
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
	IndexContent  string
	HTMLBody      string
	ImagesDir     string
	Images        []Image
}

//Image struct
type Image struct {
	URL      string
	Name     string
	HasError bool
}

//Config struct
type Config struct {
	Owner           string `json:"owner"`
	Repo            string `json:"repo"`
	PerPage         int    `json:"per_page"`
	State           string `json:"state"`
	ClientID        string `json:"client_id"`
	ClientSecret    string `json:"client_secret"`
	IsArchiveImages bool   `json:"archive_images"`
}

//APILimit struct
type APILimit struct {
	Message string `json:"message"`
	DocURL  string `json:"documentation_url"`
}

const (
	configFilename string = "config.json"
	githubLink     string = "https://github.com"
	githubAPILink  string = "https://api.github.com"
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
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
		return
	}
	issuesDir = getIssuesDir()
	issues := getIssues()
	if len(issues) == 0 {
		log.Println("no issues found.")
	}
	var buff bytes.Buffer
	for index, issue := range issues {
		issue.GetTitle()
		log.Printf("get %d: %s\n", index, issue.Title)
		issue.GetIndexContent(index)
		buff.WriteString(issue.IndexContent)
		issue.GetHTMLBody()
		issue.GetImagesDir()
		issue.GetImages()
		issue.WriteToDisk()
	}
	generateIndexHTML(buff.String())
	fmt.Println("\n\nPress Enter to continue...")
	fmt.Scanln()
}

//GetTitle GetTitle
func (issue *Issue) GetTitle() {
	title := fmt.Sprintf("%s_%s_%s_#%d.html", issue.CreatedAt[:10], issue.Title, issue.State, issue.Number)
	title = removeBadChar(title)
	issue.Title = title
}

//GetIndexContent GetIndexContent
func (issue *Issue) GetIndexContent(index int) {
	issue.IndexContent = fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td><a href='%s' target='_blank'>#%d</td></tr>", index, issue.CreatedAt[:10], issue.Title, string(blackfriday.MarkdownBasic([]byte(issue.Body))), issue.State, urlEncode(issue.Title), issue.Number)
}

//GetHTMLBody GetHTMLBody
func (issue *Issue) GetHTMLBody() {
	_, body, _ := getHTTPResponse(issue.HTMLURL)
	content := string(body)
	for {
		if !strings.Contains(content, "items not shown") || strings.Contains(issue.Title, "items not shown") {
			break
		}

		needReplaced := content[strings.Index(content, "<include-fragment") : strings.Index(content, "</include-fragment>")+19]
		link := needReplaced[strings.Index(needReplaced, "data-url=")+10:]
		link = link[:strings.Index(link, ">")-1]
		_, body, _ := getHTTPResponse(githubLink + link)
		content = strings.Replace(content, needReplaced, string(body), -1)
	}
	issue.HTMLBody = content
}

//GetImages GetImages
func (issue *Issue) GetImages() {
	if !config.IsArchiveImages {
		return
	}
	reg := regexp.MustCompile(`(https?:\/\/)([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?\.(png|jpg)`)
	imgs := reg.FindAllString(issue.HTMLBody, -1)
	var images []Image
	for _, img := range imgs {
		var image Image
		image.URL = img
		image.Name = filepath.Base(img)
		images = append(images, image)
	}
	issue.Images = images
}

//GetIssuesDir GetIssuesDir
func getIssuesDir() string {
	issuesDir = config.Owner + "_" + config.Repo + "_issues"
	if err := os.Mkdir(issuesDir, 0755); os.IsExist(err) {
		log.Printf("dir %s already exists\n", issuesDir)
	} else {
		log.Printf("create dir %s successfully\n", issuesDir)
	}
	return issuesDir
}

//ReplaceByLocalImages ReplaceByLocalImages
func (issue *Issue) ReplaceByLocalImages(image Image) {
	if !config.IsArchiveImages {
		return
	}
	if !image.HasError {
		issue.HTMLBody = strings.Replace(issue.HTMLBody, image.URL, fmt.Sprintf("%s/%s/%s", "images", urlEncode(strings.TrimRight(issue.Title, ".html")), image.Name), -1)
	}
}

//WriteToDisk WriteToDisk
func (issue *Issue) WriteToDisk() {
	if config.IsArchiveImages {
		for _, img := range issue.Images {
			img.WriteToDisk(issue.ImagesDir)
			issue.ReplaceByLocalImages(img)
		}
	}

	if err := ioutil.WriteFile(filepath.Join(issuesDir, issue.Title), []byte(issue.HTMLBody), 0755); err != nil {
		log.Println("write issue to disk error: ", err)
	}
}

//WriteToDisk WriteToDisk
func (image *Image) WriteToDisk(imagesDir string) {
	_, body, err := getHTTPResponse(image.URL)
	if err != nil {
		image.HasError = true
	} else {
		if err := ioutil.WriteFile(filepath.Join(imagesDir, image.Name), body, 0755); err != nil {
			log.Println("write issue to disk error: ", err)
		}
	}
}

//GetImagesDir GetImagesDir
func (issue *Issue) GetImagesDir() {
	if !config.IsArchiveImages {
		return
	}
	imagesDir := filepath.Join(issuesDir, "images", strings.TrimRight(issue.Title, ".html"))
	if err := os.MkdirAll(imagesDir, 0755); os.IsExist(err) {
		// log.Printf("dir %s already exists\n", imagesDir)
	} else {
		// log.Printf("create dir %s successfully\n", imagesDir)
	}
	issue.ImagesDir = imagesDir
}
func getIssues() []Issue {
	var issues []Issue
	var page = 1
	for {
		header, body, err := getHTTPResponse(fmt.Sprintf("%s/repos/%s/%s/issues?page=%d&per_page=%d&state=%s&client_id=%s&client_secret=%s", githubAPILink, config.Owner, config.Repo, page, config.PerPage, config.State, config.ClientID, config.ClientSecret))
		if err != nil {
			log.Fatalln("get issues error", err)
		}
		remaining := header["X-Ratelimit-Remaining"][0]
		reset := header["X-Ratelimit-Reset"][0]
		t, err := strconv.ParseInt(reset, 10, 64)
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

		var iss []Issue
		if err := json.Unmarshal(body, &iss); err != nil {
			log.Println("parse response.body error: ", err)
		}
		if len(iss) == 0 {
			break
		}
		issues = append(issues, iss...)
		page++
		log.Printf("hit %d issues\n", len(issues))
	}
	return issues
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

func getHTTPResponse(link string) (http.Header, []byte, error) {
	resp, err := http.Get(link)
	if err != nil {
		log.Println("get http response error: ", err)
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("get http response body error: ", err)
		return nil, nil, err
	}
	return resp.Header, body, nil
}

func removeBadChar(s string) string {
	return strings.NewReplacer("/", "", "\\", "", ":", "", "*", "", "<", "", ">", "", "|", "", "\"", "", "?", "", "", "").Replace(s)
}

func urlEncode(s string) string {
	return strings.NewReplacer("!", "%21", "#", "%23", "$", "%24", "&", "%26", "'", "%27", "(", "%28", ")", "%29", "*", "%2A", "+", "%2B", ",", "%2C", "/", "%2F", ":", "%3A", ";", "%3B", "=", "%3D", "?", "%3F", "@", "%40", "[", "%5B", "]", "%5D").Replace(s)
}

func timer(t int) {
	for t1 := t - 1; t1 >= 0; t1-- {
		for t2 := 59; t2 >= 0; t2-- {
			fmt.Printf("\rautomatically retry after %2dm%2ds", t1, t2)
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
		owner           string
		repo            string
		perPage         int
		state           string
		clientID        string
		clientSecret    string
		isArchiveImages bool
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
	flag.BoolVar(&isArchiveImages, "a", config.IsArchiveImages, "save images to local")
	flag.BoolVar(&isArchiveImages, "archive_images", config.IsArchiveImages, "save images to local")
	flag.Set("logtostderr", "true")
	flag.Parse()

	config.Owner = owner
	config.Repo = repo
	config.PerPage = perPage
	config.State = state
	config.ClientID = clientID
	config.ClientSecret = clientSecret
	config.IsArchiveImages = isArchiveImages
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
    -a, --archive_images     save images to local
		`
