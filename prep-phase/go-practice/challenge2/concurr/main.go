package main

import (
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
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

var seenRenderLinks map[string]bool
var baseURL string
var baseFilePath string
var currentLink string
var linksToMirror []string
var linkToMirrorURLMap map[string]string
var httpsLinksToMirror []string
var mirrorJobQue []mirrorJob
var linkURLMutex sync.Mutex
var renderLinkMutex sync.Mutex
var mirrorJobChan chan mirrorJob
var rootWG sync.WaitGroup
var fileTokens chan struct{}
var webTokens chan struct{}
var procLinkWorkers chan struct{}
var mJobQueMutex sync.Mutex
var doneMJobChan chan struct{}

func main() {

	// If the file doesn't exist, create it or append to the file
	file, err := os.OpenFile("concurrent_mirror_logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)

	start := time.Now()

	seenRenderLinks = make(map[string]bool)
	linkToMirrorURLMap = make(map[string]string)
	fileTokens = make(chan struct{}, 20)
	webTokens = make(chan struct{}, 20)
	procLinkWorkers = make(chan struct{}, 10)
	linkURLMutex = sync.Mutex{}
	renderLinkMutex = sync.Mutex{}
	mJobQueMutex = sync.Mutex{}
	baseURL = "https://golang.org"
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Getwd failed\n")
	} else {
		baseFilePath = dir + "/mirror"
		fmt.Printf("pwd: %s\n", dir)
	}
	seenRenderLinks[baseURL] = true

	mirrorJobQue = append(mirrorJobQue, mirrorJob{HttpsLink: baseURL, Filepath: baseFilePath + "/index.html"})
	mirrorJobChan = make(chan mirrorJob)
	doneMJobChan = make(chan struct{}, 1)

	go runMJobWorker()
	go runMJobQueWorker()

	mirrorJobChan <- mirrorJob{HttpsLink: baseURL, Filepath: baseFilePath + "/index.html"}
	time.Sleep(100 * time.Millisecond)
	rootWG.Wait()
	close(mirrorJobChan)
	doneMJobChan <- struct{}{}
	t := time.Now()
	elapsed := t.Sub(start)

	fmt.Printf("Finished!! --  %v \n", elapsed)
}

func runMJobWorker() {
	for {
		//log.Println("SELECT done or Proc Link WOrker")
		select {
		case <-doneMJobChan:
			goto stopRunning
		case procLinkWorkers <- struct{}{}:
			mJobQueMutex.Lock()
			lastIdx := len(mirrorJobQue) - 1
			if lastIdx < 0 {
				//log.Println("No Jobs")
				mJobQueMutex.Unlock()
				<-procLinkWorkers
				time.Sleep(100 * time.Millisecond)
			} else {
				//log.Println("Process MJob")
				mJob := mirrorJobQue[lastIdx]
				mirrorJobQue = mirrorJobQue[:lastIdx]
				mJobQueMutex.Unlock()
				rootWG.Add(1)
				go processMJob(mJob)
			}
		}
		//log.Println("NEXT SELECT done or Proc Link WOrker")
	}
stopRunning:
}

func runMJobQueWorker() {
	for mJob := range mirrorJobChan {
		//log.Println("Queue job")
		mJobQueMutex.Lock()
		mirrorJobQue = append(mirrorJobQue, mJob)
		mJobQueMutex.Unlock()
		//log.Println("FINISH Queue job")
	}
}

func processMJob(mJob mirrorJob) {
	httpsLink, filepath := mJob.HttpsLink, mJob.Filepath

	fmt.Printf("START Processing job https: %s || filename: %s \n", httpsLink, filepath)

	doc := getAndParseHTML(httpsLink)
	updateLinks(doc, httpsLink)

	makeMirrorFileDirectories(filepath)

	createLocalMirror(doc, filepath)
	<-procLinkWorkers
	rootWG.Done()
	fmt.Printf("FINISH Processing job https: %s || filename: %s \n", httpsLink, filepath)

}

func getAndParseHTML(urlLink string) (doc *html.Node) {
	log.Println("Acuire web token")
	webTokens <- struct{}{} //acquire token
	resp, err := http.Get(urlLink)
	<-webTokens //release token
	if err != nil {
		fmt.Fprintf(os.Stderr, "Get Request for %s: %v\n", urlLink, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	doc, err = html.Parse(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parsing html for url: %s : %v\n", urlLink, err)
		os.Exit(1)
	}

	return
}

func updateLinks(n *html.Node, currentPageURL string) { //, jobWG *sync.WaitGroup) {
	//linkWorkerTokens <- struct{}{} //acquire a token
	if n.Type == html.ElementNode && n.Data == "a" {
		for idx, a := range n.Attr {
			if a.Key == "href" {
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

func getHTTPSURLLink(originalLink, currentPageURL string) (httpsURLLink string, shouldReplaceLink, isRelative bool) {
	isRelative = false

	switch {
	case strings.HasPrefix(originalLink, "https"):
		shouldReplaceLink = false
	case strings.HasPrefix(originalLink, "http"):
		httpsURLLink = strings.Replace(originalLink, "http", "https", 1)
		shouldReplaceLink = true
	case strings.HasPrefix(originalLink, "//"):
		httpsURLLink = "https:" + originalLink
		shouldReplaceLink = true
	case strings.HasPrefix(originalLink, "/"):
		httpsURLLink = baseURL + originalLink
		shouldReplaceLink = true
	default:
		shouldReplaceLink = true
		isRelative = true
		httpsURLLink = currentPageURL
		httpsURLLink = strings.TrimRight(currentPageURL, "/") + "/" + strings.TrimLeft(originalLink, "/")
	}
	return
}

func constructMirrorLink(httpsURLLink string) (mirrorLink string) {
	if ext := path.Ext(httpsURLLink); ext != "" && !strings.Contains(ext, htmlExt) {
		return httpsURLLink
	}
	//log.Println("Acuire web token")
	webTokens <- struct{}{} //acquire a token
	resp, err := http.Get(httpsURLLink)
	<-webTokens //release token
	if err != nil {
		log.Println(httpsURLLink)
		log.Printf("http.Get => %v", err.Error())
		mirrorLink = httpsURLLink
		return
	}
	var mirrorFpath string
	finalURL := resp.Request.URL.String()
	log.Printf("FINAL URL: %s \n", finalURL)
	if strings.HasPrefix(strings.ToLower(finalURL), baseURL) {
		if !strings.HasPrefix(finalURL, baseURL) {
			log.Fatalf("Issue replacing url: %s using baseURL: %s || capitalization issue?", finalURL, baseURL)
		}

		mirrorFpath = strings.Replace(finalURL, baseURL, baseFilePath, 1)

		hasJumpTag := false
		var jumpTag string
		if strings.Contains(mirrorFpath, "#") {
			hasJumpTag = true
			if strings.Count(mirrorFpath, "#") > 1 {
				log.Fatalf("Multiple jumptags in finalURL: %s \n", finalURL)
			}
			idx := strings.Index(mirrorFpath, "#")
			jumpTag = mirrorFpath[idx:]
			mirrorFpath = mirrorFpath[:idx]

			idx = strings.Index(httpsURLLink, "#")
			httpsURLLink = httpsURLLink[:idx]
		}

		mirrorFpath = strings.TrimRight(mirrorFpath, "/") + "/index.html"

		httpsURLLink = strings.TrimRight(httpsURLLink, "/")
		if seen, ok := protectedSeenRenderLinksRead(httpsURLLink); !ok || !seen {
			if strings.HasPrefix(httpsURLLink, "#") {
				fmt.Println("BAD!!")
			}
			if success := protectedSeenRenderLinksWrite(httpsURLLink); success {
				mirrorJobChan <- mirrorJob{HttpsLink: httpsURLLink, Filepath: mirrorFpath}
			} else {
				log.Printf("Write didn't succeeed httpsURLLink: %s ;  skip mapping and channel \n", httpsURLLink)
			}
		}
		if hasJumpTag {
			mirrorFpath += jumpTag
		}
		mirrorLink = fileProtocolPrefix + mirrorFpath
		log.Printf("FINAL Mirror Link: %s \n", mirrorLink)
	} else {
		log.Printf("HTTPS URL Link when final URL not base ?: %s \n", httpsURLLink)
		mirrorLink = httpsURLLink
	}

	return
}

func getMirrorLink(link, currentPageURL string) (mirrorLink string, shouldReplaceLink bool) {
	var ok bool
	if mirrorLink, ok = protectedURLMapRead(link); ok {
		return mirrorLink, ok
	}

	var httpsURLLink string
	var isRelative bool
	if httpsURLLink, shouldReplaceLink, isRelative = getHTTPSURLLink(link, currentPageURL); !shouldReplaceLink {
		return
	}

	mirrorLink = constructMirrorLink(httpsURLLink)

	if !isRelative {
		success := protectedURLMapWrite(link, mirrorLink)
		if !success {
			verifyLinkMap(link, mirrorLink)
		}
	}

	return
}

func createLocalMirror(doc *html.Node, filename string) {
	log.Println("Acquire file token")
	fileTokens <- struct{}{} //acquire a token
	f, err := os.Create(filename)

	if err != nil {
		log.Println("On Create")
		log.Fatal(err)
	}
	err = html.Render(f, doc)
	if err != nil {
		log.Println("On Render")
		log.Fatal(err)
	}
	if closeErr := f.Close(); closeErr != nil {
		log.Println("On Close")
		log.Fatal(closeErr) //err = closeErr
	}
	<-fileTokens //release token
}

func makeMirrorFileDirectories(filename string) {
	log.Println("Acquire file token")
	fileTokens <- struct{}{} //acquire a token
	err := os.MkdirAll(path.Dir(filename), os.ModePerm.Perm())
	if err != nil {
		log.Fatalf("Something wrong with make mirror file directories: %s", filename)
	}
	<-fileTokens //release token
}

func protectedURLMapRead(link string) (url string, ok bool) {
	log.Println("lock link URL Map")
	linkURLMutex.Lock()
	url, ok = linkToMirrorURLMap[link]
	linkURLMutex.Unlock()
	return
}

func protectedURLMapWrite(link, url string) (writeSuccess bool) {
	log.Println("lock link URL Map")
	linkURLMutex.Lock()
	if _, ok := linkToMirrorURLMap[link]; ok {
		writeSuccess = false
	} else {
		linkToMirrorURLMap[link] = url
		writeSuccess = true
	}
	linkURLMutex.Unlock()
	return
}

func verifyLinkMap(link, altURL string) bool {
	var mapURL string
	if mURL, ok := protectedURLMapRead(link); ok && mURL == altURL {
		mapURL = mURL
		return true
	}
	if len(mapURL) == 0 {
		fmt.Printf("VERIFY NOT OK!!! for link: %s -- altURL: %s -- mapURL: NONE \n", link, altURL)
	} else {
		fmt.Printf("VERIFY NOT OK!!! for link: %s -- altURL: %s -- mapURL: %s \n", link, altURL, mapURL)
	}
	return false
}

func protectedSeenRenderLinksRead(link string) (seen, ok bool) {
	//log.Println("lock render Link Map")
	renderLinkMutex.Lock()
	seen, ok = seenRenderLinks[link]
	renderLinkMutex.Unlock()
	return
}

func protectedSeenRenderLinksWrite(link string) (writeSuccess bool) {
	//log.Println("lock render Link Map")
	renderLinkMutex.Lock()
	if seen, ok := seenRenderLinks[link]; seen || ok {
		writeSuccess = false
	} else {
		seenRenderLinks[link] = true
		writeSuccess = true
	}
	renderLinkMutex.Unlock()
	return
}
