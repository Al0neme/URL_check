package main

import (
	"bufio"
	"fmt"
	"log"

	// "net"
	// "net/http"
	"os"
	"strconv"
	"strings"

	// "time"

	"github.com/gocolly/colly/v2"
)

func getUrls(urlChan chan string) {
	//读取文件
	file, err := os.OpenFile("targets.txt", os.O_RDONLY, 0)
	if err != nil {
		fmt.Println("open file failed: ", err)
	}
	defer file.Close()
	//按行读取
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		url := fileScanner.Text()
		//判断是否是http形式的url
		if strings.Contains(url, "http") {
			urlChan <- url
		} else {
			urlChan <- "http://" + url
			urlChan <- "https://" + url
		}
	}
	//文件错误处理
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("read error: %s", err)
	}
	close(urlChan)
}

// 定义一个存活结果结构体
type LiveResults struct {
	url       string
	statecode int
	title     string
	err       string
}

func checkLive(url string) string {
	liveresults := &LiveResults{}
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		//设置ua头
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36")
		fmt.Println("Check url: ", r.URL)
		liveresults.url = r.URL.String()

	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Visit %s failed\n", r.Request.URL)
		fmt.Printf("Error: %v\n", err)
		liveresults.err = err.Error()

	})

	//超时设置，才开始学，不懂这一块，所以注释了这一块和相关包，按需修改吧
	// c.WithTransport(&http.Transport{
	// 	DialContext: (&net.Dialer{
	// 		Timeout:   20 * time.Second,
	// 		KeepAlive: 20 * time.Second,
	// 	}).DialContext,
	// 	MaxIdleConns:          100,
	// 	IdleConnTimeout:       90 * time.Second,
	// 	TLSHandshakeTimeout:   10 * time.Second,
	// 	ExpectContinueTimeout: 1 * time.Second,
	// })

	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("ResponseCode %d\n", r.StatusCode)
		liveresults.statecode = r.StatusCode

	})

	c.OnHTML("title", func(h *colly.HTMLElement) {
		fmt.Printf("Site Title: %s\n", h.Text)
		liveresults.title = h.Text

	})

	c.Visit(url)
	if liveresults.err != "" {
		return ""
	}
	return liveresults.url + " || " + strconv.Itoa(liveresults.statecode) + " || " + liveresults.title

}

// 获取验活结果
func getResult(urlChan chan string) {
	for {
		url, ok := <-urlChan
		if !ok {
			break
		}
		res := checkLive(url)
		fmt.Println(res)
		if res != "" {
			saveResult(res)
		}
	}
}

// 保存结果
func saveResult(res string) {
	filePath := "results.txt"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("open file failed: ", err)
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	write.WriteString(res + "\n")
	write.Flush()
}
func main() {
	urlChan := make(chan string)
	go getUrls(urlChan)
	getResult(urlChan)
}
