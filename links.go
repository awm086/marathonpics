package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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

// Copied from gopl.io/ch5/outline2.
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

func getPage(url string) (int, error) {

	response, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	defer response.Body.Close()
	out, err := os.Create("output.txt")
	if err != nil {
		fmt.Println("Error while creating file", err)
		return 0, err
	}

	defer out.Close()
	bodyRes := response.Body
	// copy to file
	n, err := io.Copy(out, bodyRes)
	if err != nil {
		panic(err)
	}
	// read it and get it's lengths
	fmt.Println(n)
	body, err := ioutil.ReadAll(bodyRes)

	if err != nil {
		return 0, err
	}

	return len(body), nil

}

func worker(urlCh chan string, sizeCh chan string, byteCh chan []string, id int) {
	for {
		url := <-urlCh
		length, err := getPage(url)
		links, _ := extractImageUrls(url)
		if err == nil {
			sizeCh <- fmt.Sprintf("%s has length %d with worker %d \n ", url, length, id)
			byteCh <- links
		} else {
			sizeCh <- fmt.Sprintf("error getting %s: %s", url, err)
		}
	}
}

func Filter(s []string, fn func(string, string) bool, keep string) []string {
	var p []string // == nil
	for _, v := range s {
		if fn(v, keep) {
			p = append(p, v)
		}
	}
	return p
}

func downloadFromUrl(url string) {
	tokens := strings.Split(url, "/")
	fileName :=  tokens[len(tokens)-1]
	fileName += ".jpg"
	fmt.Println("Downloading", url, "to", fileName)

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

	//downloadFromUrl("http://offsiteimages-02.marathonfoto.com/MFT2015-02/84/778984/1001/0005.jpg%3Fpreset=TRUE")
	url := "http://www.marathonfoto.com/Proofs?PIN=G5C410&LastName=ALMUSTAFA"
	
	imgUrls, _ := extractImageUrls(url);
	fmt.Println(len(imgUrls));

	// urls := []string{"http://www.google.com", "http://www.yahoo.com", "http://www.google.com"}
	 keep := "offsiteimages"

	 urls_filterd := Filter(imgUrls, strings.Contains, keep);
	 fmt.Println(urls_filterd);

	// urlCh := make(chan string)
	// sizeCh := make(chan string)
	// byteCh := make(chan []string)
	// // get page

	// // extract urls

	// // dispatch workers to download files

	// for i := 0; i < 10; i++ {
	// 	go worker(urlCh, sizeCh, byteCh, i)

	// }
	 for _, url := range urls_filterd {
	 	go downloadFromUrl(url);
	 }
	// for i := 0; i < len(urls); i++ {
	// 	fmt.Printf("%s\n", <-sizeCh)
	// }

	// for i := 0; i < len(urls); i++ {
	// 	//fmt.Printf("%s\n", <-byteCh)

	// }
}
