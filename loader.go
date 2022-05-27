package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"

	"gopkg.in/gcfg.v1"
)

type Spider struct {
	UrlListFile      string // 种子文件路径
	OutputDirectory  string // 抓取结果存储目录
	MaxDepth         int    // 最大抓取深度
	CrawlInterval    int    // 抓取间隔，单位：秒
	CrawlTimeout     int    // 抓取超时, 单位: 秒
	TargetUrlPattern string // 需要存储的目标网页URL Pattern
	WorkerCount      int    // 抓取routine数
}

type Config struct {
	Spider
}

// Load config from filename
func ConfigLoad(filename string) (Config, error) {
	var cfg Config
	if err := gcfg.ReadFileInto(&cfg, filename); err != nil {
		return cfg, err
	}
	if err := cfg.Check(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Check config
func (c *Config) Check() error {
	if c.UrlListFile == "" {
		return fmt.Errorf("UrlListFile is nil")
	}

	if c.OutputDirectory == "" {
		return fmt.Errorf("OutputDirectory is nil")
	}

	if c.MaxDepth < 1 {
		return fmt.Errorf("MaxDepth is less than 1")
	}

	if c.CrawlInterval < 0 {
		return fmt.Errorf("CrawlInterval is less than 0")
	}

	if c.CrawlTimeout < 1 {
		return fmt.Errorf("CrawlTimeout is less than 1")
	}

	_, err := regexp.Compile(c.TargetUrlPattern)
	if err != nil {
		return fmt.Errorf("%s: regexp.Compile(): %s", c.TargetUrlPattern, err.Error())
	}

	if c.WorkerCount < 1 {
		return fmt.Errorf("WorkerCount is less than 1")
	}

	return nil
}

// Load seeds from filename
func SeedsLoad(filename string) ([]string, error) {
	var seeds []string
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &seeds); err != nil {
		return nil, fmt.Errorf("json.Unmarshal(): %s", err.Error())
	}
	if len(seeds) == 0 {
		return nil, fmt.Errorf("no seeds in %s", filename)
	}
	return seeds, nil
}
