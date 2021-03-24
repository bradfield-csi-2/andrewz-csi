package main

import (
	"fmt"
	"golang.org/x/net/html"
	//"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	//	"path/filepath"
	"strings"
	"time"
)

const (
	fileProtocolPrefix = "file:///"
	emptyString        = ""
	htmlExt            = ".html"
)

type mirrorJob struct {
	HttpsLink, Filepath string
}

//var createdFiles[string]bool
var seenRenderLinks map[string]bool
var seenFiles map[string]bool
var baseURL string
var baseFilePath string
var currentLink string
var linksToMirror []string
var linkToMirrorURLMap map[string]string
var httpsLinksToMirror []string
var mirrorJobQue []mirrorJob

func main() {
	//seenFiles = make(map[string]bool)
	start := time.Now()

	seenRenderLinks = make(map[string]bool)
	linkToMirrorURLMap = make(map[string]string)
	baseURL = "https://golang.org"
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Gtwd failed\n")
	} else {
		baseFilePath = dir + "/mirror"
		fmt.Printf("pwd: %s\n", dir)
	}
	seenRenderLinks[baseURL] = true
	//seenFiles["/index.html"] = true
	mirrorJobQue = append(mirrorJobQue, mirrorJob{HttpsLink: baseURL, Filepath: (baseFilePath + "/index.html")})
	//linksToMirror = append(linksToMirror, "")

	processLinks()
	t := time.Now()
	elapsed := t.Sub(start)

	fmt.Printf("Finished!! --  %v \n", elapsed)
}

func processLinks() {

	for len(mirrorJobQue) > 0 {
		lastIdx := len(mirrorJobQue) - 1
		currMirrorJob := mirrorJobQue[lastIdx]
		//currentLink = link
		mirrorJobQue = mirrorJobQue[:lastIdx]

		//ASUMMING that link has / or is blank
		//urlLink := baseURL + link
		httpsURLLink := currMirrorJob.HttpsLink
		filepath := currMirrorJob.Filepath

		if strings.Contains(httpsURLLink, "main.go") {
			fmt.Printf("EFF in the CHat")
		}

		doc := getAndParseHTML(httpsURLLink)
		updateLinks(doc, httpsURLLink)

		makeMirrorFileDirectories(filepath)

		createLocalMirror(doc, filepath)

	}

}

/*
func getFullHTTPLink(link, currentFullURLPath string) httpsLink {
	//TODO: check if necessary. may never call with this kind of link
	if strings.HasPrefix(strings.ToLower(link), "https") {
		return link
	}

	//TODO: check if necessary. may never call with this kind of link
	if strings.HasPrefix(strings.ToLower(link), "http") {
		if !strings.HasPrefix(link, "http") {
			log.Fatalf("unhandled non lower case http format - link: %s \n ", link)
		}

		fmt.Printf("Replacing http with https for link: %s \n", link)
		return strings.Replace(link, "http", "https", 1)
	}

	if strings.HasPrefix(link, "//") {
		return "https:" + link
	}

	if strings.HasPrefix(link, "/") {
		//TODO: base url logic
		//can only be plus base url path right??
		return baseURL + link
	}

	if path.Base(currentFullURLPath) != "/" {
		log.Fatalf("Cannot gee relative link of non directory url path: %s | cannot get http link for: %s\n", currentFullURLPath, link)
	}

	if link == ".." {
		return path.Dir(path.Dir(currentFullURLPath))
	}

	return path.Join(currentFullURLPath, link)
}
*/
func getAndParseHTML(urlLink string) (doc *html.Node) {
	resp, err := http.Get(urlLink)
	if err != nil {
		fmt.Println("in get and Parse")
		fmt.Println(urlLink)
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err = html.Parse(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sync challenge 2: %v\n", err)
		os.Exit(1)
	}

	return
}

func updateLinks(n *html.Node, currentPageURL string) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for idx, a := range n.Attr {
			if a.Key == "href" {
				if a.Val == "archive/" {
					fmt.Println("archive debug")
				}
				if mirrorLink, shouldReplaceLink := getMirrorLink(a.Val, currentPageURL); shouldReplaceLink {
					n.Attr[idx].Val = mirrorLink
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		updateLinks(c, currentPageURL)
	}
}

func getHTTPSURLLink(originalLink, currentPageURL string) (httpsURLLink string, shouldReplaceLink bool) {
	//httpsURLLink := getHTTPSURLLink(originalLink)

	if originalLink == "archive/" {
		fmt.Println("debug")
	}
	switch {
	case strings.HasPrefix(originalLink, "https"):
		shouldReplaceLink = false
	case strings.HasPrefix(originalLink, "//"):
		httpsURLLink = "https:" + originalLink
		shouldReplaceLink = true
	case strings.HasPrefix(originalLink, "/"):
		httpsURLLink = baseURL + originalLink
		shouldReplaceLink = true
	default:
		shouldReplaceLink = true
		httpsURLLink = currentPageURL
		if !strings.HasSuffix(httpsURLLink, "/") {
			httpsURLLink += "/"
		}
		httpsURLLink += originalLink
		//log.Fatalf("Something wrong in getHTTPSURLLin. original link: %s \n", originalLink)
	}
	if httpsURLLink == "https:/golang.org/doc" {
		fmt.Println("debug")
	}
	return
}

func constructMirrorLink(httpsURLLink string) (mirrorLink string) {

	if ext := path.Ext(httpsURLLink); ext != "" && !strings.Contains(ext, ".html") {
		//mirrorLink = httpsURLLink
		return httpsURLLink
	}

	resp, err := http.Get(httpsURLLink)
	//defer resp.Body.Close()
	if err != nil {
		fmt.Println(httpsURLLink)
		log.Fatalf("http.Get => %v", err.Error())
	}

	finalURL := resp.Request.URL.String()

	if strings.HasPrefix(strings.ToLower(finalURL), baseURL) {
		if !strings.HasPrefix(finalURL, baseURL) {
			log.Fatalf("Issue replacing url: %s using baseURL: %s || capitalization issue?", finalURL, baseURL)
		}

		mirrorLink = strings.Replace(finalURL, baseURL, baseFilePath, 1)

		hasJumpTag := false
		var jumpTag string
		if strings.Contains(mirrorLink, "#") {
			hasJumpTag = true
			if strings.Count(mirrorLink, "#") > 1 {
				log.Fatalf("Multiple jumptags in finalURL: %s \n", finalURL)
			}
			idx := strings.Index(mirrorLink, "#")
			jumpTag = mirrorLink[idx:]
			mirrorLink = mirrorLink[:idx]

			idx = strings.Index(httpsURLLink, "#")
			httpsURLLink = httpsURLLink[:idx]
		}

		//TODO:if changed to dir name for file name add dir var back form throwaway _
		_, file := path.Split(mirrorLink)

		if file == "" {
			//this means it's a directory
			//TODO: change file name to path.Base(dir) + ".html"
			file = "index.html"
		}

		if file != "" {
			mirrorLink = path.Join(mirrorLink, file)
		}

		if !seenRenderLinks[httpsURLLink] {
			if strings.HasPrefix(httpsURLLink, "#") {
				fmt.Println("BAD!!")
			}
			mirrorJobQue = append(mirrorJobQue, mirrorJob{HttpsLink: httpsURLLink, Filepath: mirrorLink})
			seenRenderLinks[httpsURLLink] = true
		}
		/*
			if !seenFiles[mirrorLink] {
				linksToMirror = append(linksToMirror, httpsURLLink)
				seenFiles[mirrorLink] = true
			}
		*/
		//WHY are we tracking both? because we are sending the https link to be parsed and rendered to a file
		//we need to send the mirrorLink back to the html node to replace for a render
		if hasJumpTag {

			//TODO: check if we need to replace and resend the httpsURLLink
			//httpsURLLink += jumpTag
			mirrorLink += jumpTag
		}
		mirrorLink = fileProtocolPrefix + mirrorLink

		//TODO: what happens if jump tag
		// 1. replacement needs to have jump tag
		// 2. non jump link needs to be added to list if not seen yet
		// add mirrorLink to mapping
		//need to change and add index of other document name if no extension?
		//grrrr:w
		//`
		//if has bookmark then modify

	} else {
		mirrorLink = httpsURLLink
		//needsToBeProcessed = false //?????
	}

	if mirrorLink == "archive/" {
		fmt.Println("uh oh what.")
	}
	//if has some jump link - then only add below if the main url also needs to be added
	//if not seen before then need to add to final map
	//linkToMirrorURLMap[httpsURLLink] = replacementLink
	//linksToMirror = append(linksToMirror, httpsURLLink)
	return
}

func getMirrorLink(link, currentPageURL string) (mirrorLink string, shouldReplaceLink bool) {
	var ok bool
	if mirrorLink, ok = linkToMirrorURLMap[link]; ok {
		shouldReplaceLink = true
		return
	}

	var httpsURLLink string
	if httpsURLLink, shouldReplaceLink = getHTTPSURLLink(link, currentPageURL); !shouldReplaceLink {
		return
	}

	if strings.Contains(httpsURLLink, "#hdr-Download_and_install_packages_and_dependencies") {
		fmt.Println("debug")
	}
	mirrorLink = constructMirrorLink(httpsURLLink) //getReplacementLink(httpsURLLink)

	linkToMirrorURLMap[link] = mirrorLink

	return
}

func createLocalMirror(doc *html.Node, filename string) {
	//fmt.Printf("Creating local file: %s\n", filename)
	f, err := os.Create(filename)
	err = html.Render(f, doc)
	if err != nil {
		log.Fatal(err)
	}
	//n, err = io.Copy(f, resp.Body)
	// Close file, but prefer error from Copy, if any.
	if closeErr := f.Close(); closeErr != nil {
		log.Fatal(closeErr) //err = closeErr
	}
}

func makeMirrorFileDirectories(filename string) {
	err := os.MkdirAll(path.Dir(filename), os.ModePerm.Perm())
	if err != nil {
		log.Fatalf("Something wrong with make mirror file directories: %s", filename)
	}
}
