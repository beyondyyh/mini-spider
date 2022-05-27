package main

import (
	"fmt"
	"regexp"
	"strings"
)

type TaskConfig struct {
	CrawlTimeout    int            // 爬取超时，单位：秒
	OutputDirectory string         // 网页下载目录
	TargetUrlRegex  *regexp.Regexp // 需要存储的目标网页正则表达式
}

type Task struct {
	Url    string      // 爬取url
	Depth  int         // 爬取深度
	Config *TaskConfig // task配置
}

// Create a new endless task.
func NewEndlessTask(maxDepth int) *Task {
	return &Task{
		Url:   "",
		Depth: maxDepth,
	}
}

// Run single task.
// A successful call returns sub url list and err == nil.
func (task *Task) Run() ([]string, error) {
	data, contentType, err := Crawl(task.Url, task.Config.CrawlTimeout)
	if err != nil {
		return nil, fmt.Errorf("%s: Crawl() err: %s", task.Url, err.Error())
	}
	if !strings.Contains(contentType, "text") {
		return nil, fmt.Errorf("%s: Content-Type: %s", task.Url, contentType)
	}
	data, err = Convert2UTF8(data, contentType)
	if err != nil {
		return nil, fmt.Errorf("%s: Convert2UTF8() err: %s", task.Url, err.Error())
	}

	if task.Config.TargetUrlRegex.MatchString(task.Url) {
		if err = task.SaveData(data); err != nil {
			return nil, fmt.Errorf("%s: task.SaveData() err: %s", task.Url, err.Error())
		}
	}

	urlList, err := FetchUrlList(data, task.Url)
	if err != nil {
		return nil, fmt.Errorf("%s: FetchUrlList() err: %s", task.Url, err.Error())
	}

	return urlList, nil
}

// Save data to output directory.
func (task *Task) SaveData(data []byte) error {
	err := SaveData(data, task.Url, task.Config.OutputDirectory)
	if err != nil {
		return fmt.Errorf("SaveData(): %s", err.Error())
	}
	return nil
}
