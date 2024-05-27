package source

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/gookit/color"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
	"urls/common"
)

const (
	concurrencyLimit = 10
)

func UrlScan() {
	if common.UrlFile != "" {
		urls, err := readUrlsFromFile(common.UrlFile)
		if err != nil {
			log.Printf("Error reading URLs from file: %v\n", err)
			return
		}
		if common.OutputToFile {
			var err error
			common.OutputFile, err = os.Create("result.txt")
			if err != nil {
				log.Fatalf("Failed to create result file: %v", err)
			}
			log.SetOutput(common.OutputFile)
		}
		processUrls(urls)
	} else if common.Urls != nil {
		processUrls(common.Urls)
	}
}

// readUrlsFromFile reads URLs from a given file path.
func readUrlsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var urls []string
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		url = formatURL(url)
		urls = append(urls, url)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}

// processUrls handles the concurrency for URL processing.
func processUrls(urls []string) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrencyLimit)

	for _, url := range urls {
		wg.Add(1)
		semaphore <- struct{}{} // acquire a slot
		go func(u string) {
			defer wg.Done()
			GetUrl(u)
			<-semaphore // release the slot
		}(url)
	}
	wg.Wait()
}

// formatURL ensures the URL has the correct scheme.
func formatURL(url string) string {
	if !strings.Contains(url, "://") {
		if strings.Contains(url, ":443") {
			url = strings.Replace(url, ":443", "", 1)
			url = "https://" + url
		} else {
			url = "http://" + url
		}
	}
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return url
}

// GetUrl makes an HTTP GET request to the specified URL.
func GetUrl(url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &http.Client{}
	req, err := createRequest(ctx, url)
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	processResponse(resp)
}

// createRequest initializes a new HTTP request with common headers.
func createRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", getRandomUserAgent())
	req.Header.Set("Accept", common.Accept)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Connection", "close")
	return req, nil
}

// processResponse handles the HTTP response.
func processResponse(resp *http.Response) {
	body, err := GetRespBody(resp)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return
	}

	title, lengthInKB, reurl := extractDetailsFromResponse(resp, body)
	printDetails(resp.Request.URL.String(), resp.StatusCode, lengthInKB, title, reurl)
}

// extractDetailsFromResponse extracts relevant data from the response.
func extractDetailsFromResponse(resp *http.Response, body []byte) (string, float64, string) {
	title := GetTitle(body)
	lengthInBytes, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	lengthInKB := float64(lengthInBytes) / 1024.0
	reurl := ""
	if redirURL, err := resp.Location(); err == nil {
		reurl = redirURL.String()
	}
	return title, lengthInKB, reurl
}

// printDetails formats and outputs the details to console and optionally to a file.
func printDetails(url string, statusCode int, lengthInKB float64, title, reurl string) {
	color.FgWhite.Printf("%-25v ", url)
	printStatusCode(statusCode)
	color.FgYellow.Printf("[%.2fKB] ", lengthInKB)
	color.FgCyan.Printf("[%v]\n", title)
	if reurl != "" {
		fmt.Printf("Redirect URL: %s\n", reurl)
	}
	color.Reset()
	PrintResult("%-25v [%d] [%.2fKB] [%v]\n", url, statusCode, lengthInKB, title)
	if reurl != "" {
		PrintResult("Redirect URL: %s\n", reurl)
	}
}

// printStatusCode colors the status code based on its value.
func printStatusCode(statusCode int) {
	switch {
	case statusCode == 200:
		color.FgGreen.Printf("[%d] ", statusCode)
	case statusCode >= 400:
		color.FgRed.Printf("[%d] ", statusCode)
	default:
		fmt.Printf("[%d] ", statusCode)
	}
}

// GetRespBody reads and decompresses the HTTP response body if necessary.
func GetRespBody(oResp *http.Response) ([]byte, error) {
	var body []byte
	if oResp.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(oResp.Body)
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(gr)
		if err != nil {
			return nil, err
		}
		body = buf.Bytes()
	} else {
		raw, err := io.ReadAll(oResp.Body)
		if err != nil {
			return nil, err
		}
		body = raw
	}
	return body, nil
}

// GetTitle extracts the title from the HTML body.
func GetTitle(body []byte) string {
	if !utf8.Valid(body) {
		body, _ = simplifiedchinese.GBK.NewDecoder().Bytes(body)
	}
	re := regexp.MustCompile("(?ims)<title.*?>(.*?)</title>")
	find := re.FindSubmatch(body)
	if len(find) > 1 {
		title := strings.TrimSpace(string(find[1]))
		title = strings.ReplaceAll(title, "\n", "")
		title = strings.ReplaceAll(title, "\r", "")
		title = strings.ReplaceAll(title, "&nbsp;", " ")
		if len(title) > 100 {
			title = title[:100]
		}
		return title
	}
	return "None"
}

// getRandomUserAgent selects a random user agent from a predefined list.
func getRandomUserAgent() string {
	randomIndex := rand.Intn(len(common.UserAgents))
	return common.UserAgents[randomIndex]
}

// PrintResult prints or logs the formatted result.
func PrintResult(format string, a ...interface{}) {
	// 如果指定了输出文件，则将结果写入文件
	if common.OutputFile != nil {
		fmt.Fprintf(common.OutputFile, format, a...)
	}
}
