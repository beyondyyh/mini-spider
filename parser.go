package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

// Convert raw to utf8.
func Convert2UTF8(raw []byte, contentType string) ([]byte, error) {
	reader := bytes.NewReader(raw)
	utf8Reader, err := charset.NewReader(reader, contentType)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(utf8Reader)
}

// Get all sub url of node.
func getUrlList(node *html.Node, refUrl *url.URL) []string {
	var urlList []string
	if node.Type == html.ElementNode && node.Data == "a" {
		for _, a := range node.Attr {
			if a.Key == "href" {
				if a.Val != "javascript:;" && a.Val != "javascript:void(0)" {
					url, err := refUrl.Parse(a.Val)
					if err == nil {
						urlList = append(urlList, url.String())
					}
				}
				break
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		childUrlList := getUrlList(child, refUrl)
		urlList = append(urlList, childUrlList...)
	}

	return urlList
}

// Get sub url list of pre url from data.
func FetchUrlList(data []byte, preUrl string) ([]string, error) {
	// parse html
	node, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("html.Parse(): %s", err.Error())
	}

	// parse url
	refUrl, err := url.ParseRequestURI(preUrl)
	if err != nil {
		return nil, fmt.Errorf("%s: url.ParseRequestURL(): %s", preUrl, err.Error())
	}

	urlList := getUrlList(node, refUrl)

	return urlList, nil
}

// Parse hostname from raw url.
func ParseHostname(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if u.Host == "" {
		return "", fmt.Errorf("empty host")
	}

	// 可能出现如xxx.baidu.com:8080这样带端口号的情况
	hostname := strings.Split(u.Host, ":")
	if len(hostname) == 0 {
		return "", fmt.Errorf("invalid hostname")
	}

	return hostname[0], nil
}
