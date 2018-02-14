package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/docopt/docopt-go"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/russross/blackfriday"
	l "github.com/gomeet/gomeet-tools-markdown-server/http_log"
	"github.com/gomeet/gomeet-tools-markdown-server/utils/assets"
)

const version = `gomeet-tools-markdown-server 1.0`
const usage = `Usage: gomeet-tools-markdown-server [-v] [--root=DIR] [ADDR]

Options:
  -h --help        Show this screen.
     --version     Show version.
  -v --verbose     Show more information.
     --root=DIR    Document root. [Default: .]
`

var (
	verbose  bool
	httpAddr string
	rootDir  string
)

func init() {
	opts, _ := docopt.Parse(usage, nil, true, version, false)

	log.SetFlags(0)

	var err error

	verbose = opts["--verbose"].(bool)

	if opts["ADDR"] == nil {
		freePort, err := getFreePort()
		if err == nil {
			opts["ADDR"] = fmt.Sprintf("localhost:%d", freePort)
		} else {
			opts["ADDR"] = "127.0.0.1:8080"
		}
	}
	httpAddr = opts["ADDR"].(string)

	rootDir, err = filepath.Abs(opts["--root"].(string))
	if err != nil {
		log.Fatalf("Fatal: invalid document root %q: %v", opts["--root"], err)
	}
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func main() {
	log.Printf("httpAddr=%v rootDir=%v", httpAddr, rootDir)

	http.Handle(
		"/assets/",
		l.Log(
			http.StripPrefix(
				"/assets/",
				http.FileServer(
					&assetfs.AssetFS{
						Asset:     assets.Asset,
						AssetDir:  assets.AssetDir,
						AssetInfo: assets.AssetInfo,
						Prefix:    "assets",
					},
				),
			),
		),
	)
	http.Handle("/markdown/", l.Log(http.StripPrefix("/markdown/", http.HandlerFunc(markdown))))
	http.Handle("/favicon.ico", l.Log(http.HandlerFunc(favicon)))
	http.Handle("/", l.Log(http.HandlerFunc(index)))

	log.Printf("starting server at %v", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}

func favicon(w http.ResponseWriter, r *http.Request) {
	// TODO: serve favicon.ico
}

func renderFile(w http.ResponseWriter, file string, data interface{}) {
	contents, err := assets.Asset(file)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	t, err := template.
		New(file).
		Parse(string(contents))
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	matches, err := filepath.Glob(filepath.Join(rootDir, "*.md"))
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	files := []string{}
	for _, m := range matches {
		dir, _ := filepath.Rel(rootDir, m)
		files = append(files, dir)
	}

	renderFile(w, "assets/index.html", files)
}

type Markdown struct {
	Filename string
	Content  string
}

func markdown(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadFile(filepath.Join(rootDir, r.URL.Path))
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	flags := blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	extensions := blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_FOOTNOTES |
		blackfriday.EXTENSION_HEADER_IDS |
		blackfriday.EXTENSION_TITLEBLOCK |
		blackfriday.EXTENSION_AUTO_HEADER_IDS
	renderer := blackfriday.HtmlRenderer(flags, "", "")

	markdown := &Markdown{
		Filename: r.URL.Path,
		Content:  string(blackfriday.Markdown(b, renderer, extensions)),
	}

	renderFile(w, "assets/markdown.html", markdown)
}
