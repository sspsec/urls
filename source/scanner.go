package source

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/gookit/color"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"
	"urls/common"
)

const (
	concurrencyLimit = 10
)

func GetUrl(url string) {
	client := &http.Client{}
	var title string
	var reurl string

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("User-Agent", getRandomUserAgent())
	req.Header.Set("Accept", common.Accept)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Connection", "close")

	resp, err := client.Do(req)

	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := GetRespBody(resp)
	if err != nil {
		return
	}

	if !utf8.Valid(body) {
		body, _ = simplifiedchinese.GBK.NewDecoder().Bytes(body)
	}
	title = GetTitle(body)
	length := resp.Header.Get("Content-Length")
	if length == "" {
		length = fmt.Sprintf("%v", len(body))
	}
	redirURL, err1 := resp.Location()
	if err1 == nil {
		reurl = redirURL.String()
	}

	statusCode := resp.StatusCode
	switch statusCode {
	case 200:
		color.Set(color.FgGreen)
	case 403, 301, 302:
		color.Set(color.FgRed)
	default:
		color.Set(color.FgWhite)
	}

	fmt.Printf("%-25v [%d] [%v] [%v]\n", resp.Request.URL, statusCode, length, title)
	if reurl != "" {
		fmt.Printf(" 跳转url: %s\n", reurl)
	}
	if err != nil {
		return
	}
	if resp.StatusCode == 400 && !strings.HasPrefix(url, "https") {
		return
	}
	return
}

func GetRespBody(oResp *http.Response) ([]byte, error) {
	var body []byte
	if oResp.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(oResp.Body)
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		for {
			buf := make([]byte, 1024)
			n, err := gr.Read(buf)
			if err != nil && err != io.EOF {
				return nil, err
			}
			if n == 0 {
				break
			}
			body = append(body, buf...)
		}
	} else {
		raw, err := io.ReadAll(oResp.Body)
		if err != nil {
			return nil, err
		}
		body = raw
	}
	return body, nil
}

func GetTitle(body []byte) (title string) {
	re := regexp.MustCompile("(?ims)<title.*?>(.*?)</title>")
	find := re.FindSubmatch(body)
	if len(find) > 1 {
		title = string(find[1])
		title = strings.TrimSpace(title)
		title = strings.Replace(title, "\n", "", -1)
		title = strings.Replace(title, "\r", "", -1)
		title = strings.Replace(title, "&nbsp;", " ", -1)
		if len(title) > 100 {
			title = title[:100]
		}
		if title == "" {
			title = "\"\"" //空格
		}
	} else {
		title = "None" //没有title
	}
	return
}

func getRandomUserAgent() string {
	randomIndex := rand.Intn(len(common.UserAgents))
	return common.UserAgents[randomIndex]
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
		url = formatURL(url)
		urls = append(urls, url)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

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

func UrlScan() {
	if common.UrlFile != "" {
		urlsFromFile, err := readUrlFromFile(common.UrlFile)
		if err != nil {
			log.Printf("Error reading URLs from file: %v\n", err)
			return
		}

		var wg sync.WaitGroup
		semaphore := make(chan struct{}, concurrencyLimit)

		for _, url := range urlsFromFile {
			wg.Add(1)
			semaphore <- struct{}{} // acquire semaphore
			go func(u string) {
				defer func() {
					<-semaphore // release semaphore
					wg.Done()
				}()
				GetUrl(u)
			}(url)
		}

		wg.Wait()
	}

	if common.Urls != nil {
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, concurrencyLimit)

		for _, url := range common.Urls {

			wg.Add(1)
			semaphore <- struct{}{} // acquire semaphore
			go func(u string) {
				defer func() {
					<-semaphore // release semaphore
					wg.Done()
				}()
				GetUrl(u)
			}(url)
		}

		wg.Wait()
	}
}
