package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	urlFile string
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:54.0) Gecko/20100101 Firefox/54.0",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36,Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/27.0.1453.93 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36,Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.17 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36,Mozilla/5.0 (X11; NetBSD) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/27.0.1453.116 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML like Gecko) Chrome/44.0.2403.155 Safari/537.36",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US) AppleWebKit/533.20.25 (KHTML, like Gecko) Version/5.0.4 Safari/533.20.27",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:23.0) Gecko/20130406 Firefox/23.0",
	"Opera/9.80 (Windows NT 5.1; U; zh-sg) Presto/2.9.181 Version/12.00",
}

func init() {
	flag.StringVar(&urlFile, "f", "url.txt", "urls file default url.txt")
}

func getStatusCode(url string) (int, error) {
	transport := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
	}
	http.DefaultTransport.(*http.Transport).DisableKeepAlives = true
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	req.Close = true
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return 0, err
	}

	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func checkURL(url string) {

	statusCode, err := getStatusCode(url)
	if err != nil {
		//fmt.Printf("%s [EOF]\n", url)
		return
	}

	title, err := getWebsiteTitle(url)
	if err != nil {
		//fmt.Printf("Error fetching title for URL %s: %v\n", url, err)
		return
	}

	fmt.Printf("%s [%d] [%s]\n", url, statusCode, title)

}

func getWebsiteTitle(url string) (string, error) {
	transport := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
	}
	http.DefaultTransport.(*http.Transport).DisableKeepAlives = true
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	req.Close = true
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return "", err
	}

	req.Header.Set("User-Agent", getRandomUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		//fmt.Printf("Error sending request: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if !utf8.Valid(body) {
		body, _ = simplifiedchinese.GBK.NewDecoder().Bytes(body)
	}
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return "", err
	}

	//fmt.Printf(string(body))

	title := getTitleFromHTML(body)
	return title, nil
}
func getRandomUserAgent() string {
	randomIndex := rand.Intn(len(userAgents))
	return userAgents[randomIndex]
}

func getTitleFromHTML(html []byte) string {

	titleStart := "<title>"
	titleEnd := "</title>"

	startIndex := strings.Index(string(html), titleStart)
	endIndex := strings.Index(string(html), titleEnd)

	if startIndex == -1 || endIndex == -1 || endIndex <= startIndex+len(titleStart) {
		return "Title Not Found"
	}

	return strings.TrimSpace(string(html[startIndex+len(titleStart) : endIndex]))
}

func readUrlFromFile(filePath string) ([]string, error) {
	var urls []string

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if strings.Contains(url, ":443") && !(strings.Contains(url, "://")) {
			url = strings.Replace(url, ":443", "", 1)
			url = "https://" + url
		} else if !(strings.Contains(url, "://")) {
			url = "http://" + url
		}
		if !strings.Contains(url[len(url)-1:], "/") {
			url = url + "/"
		}
		urls = append(urls, url)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, err
}

func main() {
	flag.Parse()
	urls, err := readUrlFromFile(urlFile)

	if err != nil {
		fmt.Printf("Error reading URLs from file: %v\n", err)
	}
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			checkURL(u)
		}(url)
	}

	wg.Wait()
}
