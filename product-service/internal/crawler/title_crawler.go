package crawler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type TitleCrawler struct {
	workerNum int
	timeout   time.Duration
}

func NewTitleCrawler(workerNum int, timeout time.Duration) *TitleCrawler {
	return &TitleCrawler{
		workerNum: workerNum,
		timeout:   timeout,
	}
}

func (c *TitleCrawler) Fetch(urls []string) map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	taskChan := make(chan string)
	result := make(map[string]string)
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for i := 0; i <= c.workerNum; i++ {
		wg.Add(1)
		go c.Worker(ctx, &wg, taskChan, result, &mu)
	}

	go func() {
		defer close(taskChan)
		for _, url := range urls {
			select {
			case taskChan <- url:
			case <-ctx.Done():
				fmt.Println("投递任务超时")
				return
			}
		}
	}()
	wg.Wait()
	return result
}

func (c *TitleCrawler) Worker(ctx context.Context, wg *sync.WaitGroup, taskChan <-chan string, result map[string]string, mu *sync.Mutex) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case url, ok := <-taskChan:
			if !ok {
				return
			}
			title := fetchTitle(url)
			mu.Lock()
			result[url] = title
			mu.Unlock()
		}
	}
}

// 模拟从url中获取标题
func fetchTitle(url string) string {
	time.Sleep(300 * time.Millisecond)
	return "Title of " + url
}
