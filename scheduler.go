package main

import (
	"regexp"
	"sync"
	"time"

	"github.com/baidu/go-lib/log"
	"github.com/beyondyyh/libs/queue"
)

type Scheduler struct {
	TaskQueue     queue.Queue   // 任务队列
	TaskCfg       *TaskConfig   // task配置
	TaskChan      chan struct{} // 任务channel
	MaxDepth      int           // 最大爬取深度
	CrawlInterval int           // 爬取间隔，单位：秒
	WorkerCount   int           // 并发数
	UrlTable      sync.Map      // url去重表
	TimerTable    sync.Map      // 站点爬取间隔timer表
}

func NewScheduler(config Config, seeds []string) *Scheduler {
	// we have checked TargetUrl in config's check,
	// so we do not need to check again,
	// just ignore the possible error
	targetUrlRegex, _ := regexp.Compile(config.TargetUrlPattern)
	taskCfg := &TaskConfig{
		CrawlTimeout:    config.CrawlTimeout,
		OutputDirectory: config.OutputDirectory,
		TargetUrlRegex:  targetUrlRegex,
	}

	// Init crawler
	s := &Scheduler{
		TaskCfg:       taskCfg,
		TaskChan:      make(chan struct{}, config.WorkerCount),
		MaxDepth:      config.MaxDepth,
		CrawlInterval: config.CrawlInterval,
		WorkerCount:   config.WorkerCount,
	}

	// initialize task queue
	s.TaskQueue.Init()
	for _, seed := range seeds {
		task := &Task{
			Url:    seed,
			Depth:  0,
			Config: taskCfg,
		}
		s.TaskQueue.Push(task)
	}

	return s
}

// Start to run tasks
func (s *Scheduler) Start() {
	log.Logger.Info("Start to run tasks")
	for {
		if s.TaskQueue.Len() == 0 && len(s.TaskChan) == 0 {
			// 将新任务加入任务队列这一操作包含在任务中
			// len(s.TaskChan) == 0 说明后续一定没有新任务被加入到s.TaskQueue中
			log.Logger.Info("TaskQueue is empty")
			break
		}

		// s.TaskQueue.Len() == 0 && len(s.TaskChan) != 0 说明当前还有任务在运行
		// 只是还没有生成新的任务加入到任务队列 有了EndlessTask的存在 后续一定会有任务被加入到任务队列
		// s.TaskQueue.Remove()可能会等待一段时间 但不会阻塞

		// s.TaskQueue.Len() != 0 && len(s.TaskChan) == 0说明当前没有任务在运行
		// s.TaskQueue不为空 直接从s.TaskQueue里取任务然后运行即可

		// s.TaskQueue.Len() != 0 && len(s.TaskChan) != 0说明当前还有任务在运行
		// s.TaskQueue不为空 直接从s.TaskQueue里取任务然后运行即可

		task := s.TaskQueue.Pop()
		s.RunTask(task.(*Task))
	}

	close(s.TaskChan)
	log.Logger.Info("All tasks has been done")
}

func (s *Scheduler) RunTask(task *Task) {
	if task.Depth >= s.MaxDepth {
		return
	}

	// 避免重复抓取
	// 如果task.Url已经存在于urlTable中了则返回的loaded的值为true，否则loaded的值为false并将task.Url加入到urlTable中
	if _, loaded := s.UrlTable.LoadOrStore(task.Url, true); loaded {
		// 该url的内容正在抓取或者已经抓取过了 直接返回
		return
	}

	s.TaskChan <- struct{}{}
	go func() {
		// endlessTask是为了在任务队列变为空之前排空TaskChan从而优雅退出 endlessTask一进入RunTask方法就会返回不会向TaskChan添加元素
		// 有的任务爬虫任务可能不会取到符合条件的子url（可能某个url下没有子url 也可能有子url但子url不能匹配正则表达式）
		// 不管有没有符合条件的子url都往任务队列里加一个endlessTask可以保证在Start方法的for循环里遇到
		// s.TaskQueue.Len() == 0 && len(s.TaskChan) != 0的情况下不会阻塞
		endlessTask := NewEndlessTask(s.MaxDepth)

		defer func() {
			log.Logger.Info("Task %s done", task.Url)
			// append useless task
			s.TaskQueue.Push(endlessTask)
			<-s.TaskChan
		}()

		// 控制抓取间隔 防止被封禁
		hostname, err := ParseHostname(task.Url)
		if err != nil {
			log.Logger.Error("%s: ParseHostname(): %s", task.Url, err.Error())
			return
		}
		timer, ok := s.TimerTable.LoadOrStore(hostname, time.NewTimer(time.Duration(s.CrawlInterval)*time.Second))
		if ok {
			select {
			case <-timer.(*time.Timer).C:
			}
			timer.(*time.Timer).Reset(time.Duration(s.CrawlInterval) * time.Second)
		}

		log.Logger.Info("Start to crawl %s", task.Url)
		urlList, err := task.Run()
		if err != nil {
			log.Logger.Error("%s", err.Error())
			return
		}

		// generate next new task
		for _, url := range urlList {
			nextTask := &Task{
				Url:    url,
				Depth:  task.Depth + 1,
				Config: s.TaskCfg,
			}
			s.TaskQueue.Push(nextTask)
		}
	}()
}
