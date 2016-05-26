package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"log"
    "net/url"
)

// Extract makes an HTTP GET request to the specified URL, parses
// the response as HTML, and returns the links in the HTML document.
func extractImageUrls(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("getting %s: %s", url, resp.Status)
	}
	doc, err := html.Parse(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("parsing %s as HTML: %v", url, err)
	}

	var links []string
	visitNode := func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, img := range n.Attr {
				if img.Key != "src" {
					continue
				}

				link, err := resp.Request.URL.Parse(img.Val)

				if err != nil {
					continue // ignore bad URLs
				}
				links = append(links, link.String())
			}
		}
	}
	forEachNode(doc, visitNode, nil)
	return links, nil
}


func forEachNode(n *html.Node, pre, post func(n *html.Node)) {
	if pre != nil {
		pre(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		forEachNode(c, pre, post)
	}
	if post != nil {
		post(n)
	}
}

func Filter(s []string, fn func(string, string) bool, keep string) []string {
	var p []string 
	for _, v := range s {
		if fn(v, keep) {
			p = append(p, v)
		}
	}
	return p
}

func downloadFromUrl(url string, wg *sync.WaitGroup, fileName string) {
	defer wg.Done();
	
	// TODO: check file existence first with io.IsExist
	output, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}

	fmt.Println(n, "bytes downloaded.")
}

func main() {

	//url := "http://www.marathonfoto.com/Proofs?PIN=G5C410&LastName=ALMUSTAFA"
	urlarg := os.Args[1];
	fmt.Println(urlarg);
	// Valid url?
	_, err1 := url.Parse(urlarg)
    if err1 != nil {
    	fmt.Println("Need to provide Valid URL in the argument")
       	log.Fatal(err1)
    }


	dir := "pics/";

	err := os.MkdirAll(dir,0711)

	if err != nil {
		fmt.Println("Error while creating", dir, "-", err)
		os.Exit(1);
	}

	imgUrls, _ := extractImageUrls(urlarg)
	if len(imgUrls) == 0 {
		fmt.Println("No pictures found.")
	}
	keep := "offsiteimages"

	urls_filterd := Filter(imgUrls, strings.Contains, keep)
	
	var wg sync.WaitGroup
	for _, url := range urls_filterd {
		url = strings.Replace(url, "preset=t", "preset=TRUE", -1)
		// Construct filenames. 
		tokens := strings.Split(url, "/")
		fileName := tokens[len(tokens)-2] + tokens[len(tokens)-1];
		fileName =  strings.Replace(fileName, "?preset=TRUE", "", -1);

		// Increment syncgroup
		wg.Add(1)
		log.Println(url);
		// Dispatch routines.
		path := dir + fileName;
		go downloadFromUrl(url, &wg, path);

	}

	wg.Wait()

}
