package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.64 Safari/537.36"
)

// Crawl url within timeout.
// A successful call returns data, content type of url and err == nil.
func Crawl(url string, timeout int) ([]byte, string, error) {
	var body []byte
	var contentType string
	var err error

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("%s: http.NewRequest(): %s", url, err.Error())
	}
	req.Header.Add("User-Agent", UserAgent)

	// errChan用于接收下面这个go routine中可能返回的错误
	// 无论是否执行出错 go routine执行结束时一定会往errChan中发送信号
	// 因此其同时也标志着go routine是否执行完毕
	// 使用有缓冲区的channel并设置缓冲区大小为1是为了出现超时状况时向errChan中发送信号不会阻塞 防止go routine泄漏
	errChan := make(chan error, 1)
	timer := time.NewTimer(time.Duration(timeout) * time.Second)

	go func() {
		var resp *http.Response
		resp, err = client.Do(req)
		if err != nil {
			err = fmt.Errorf("%s: client.Do(): %s", url, err.Error())
			errChan <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("%s: status code[%d] not 200", url, resp.StatusCode)
			errChan <- err
			return
		}

		contentType = resp.Header.Get("Content-Type")

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("ioutil.ReadAll(): %s", err.Error())
			errChan <- err
			return
		}

		errChan <- nil
	}()

	// wait until crawl done or timeout
	select {
	case err = <-errChan:
		if err != nil {
			return nil, "", err
		}
	case <-timer.C:
		return nil, "", fmt.Errorf("crawl timeout")
	}

	return body, contentType, err
}
