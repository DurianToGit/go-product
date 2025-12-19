package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {
	/*var i time.Duration
	for i = 0; i <= 3; i++ {
		result, err := doWorkWithTimeout(i*time.Second, 2*time.Second)
		if err != nil {
			fmt.Println("错误:", err)
		} else {
			fmt.Println("结果:", result)
		}
	}*/
	// retryWithTimeout(3, 1*time.Second)
	test1()
}

func test1() {
	// 1. 有缓冲 vs 无缓冲演示
	unbuffered := make(chan int)
	buffered := make(chan int, 2)

	// 2. select 多路复用
	tick := time.Tick(100 * time.Millisecond)
	boom := time.After(500 * time.Millisecond)

	for {
		select {
		case <-tick:
			fmt.Println("滴答")
		case <-boom:
			fmt.Println("boom！")
			return
		default:
			// 3. 非阻塞操作
			select {
			case buffered <- 1:
				fmt.Println("向缓冲通道发送成功")
			default:
				// 缓冲区满时不阻塞
				fmt.Println("缓冲区已满")
			}

			// 检查无缓冲通道（非阻塞）
			select {
			case unbuffered <- 1:
				fmt.Println("向无缓冲通道发送成功")
			default:
				fmt.Println("无缓冲通道暂无接收者")
			}

			time.Sleep(50 * time.Millisecond)
		}
	}
}

func retryWithTimeout(maxRetries int, timeout time.Duration) {
	for i := 0; i < maxRetries; i++ {
		ch := make(chan bool, 1)

		go func(attempt int) {
			time.Sleep(time.Duration(attempt) * time.Second)
			ch <- true
		}(i)

		select {
		case <-ch:
			fmt.Printf("第%d次尝试成功\n", i+1)
			// return
		case <-time.After(timeout):
			fmt.Printf("第%d次尝试超时\n", i+1)
		}
	}
	fmt.Println("所有尝试都失败了")
}

func doWorkWithTimeout(duration, timeout time.Duration) (string, error) {
	ch := make(chan string, 1)

	go func() {
		time.Sleep(duration)
		ch <- "工作完成"
	}()

	select {
	case result := <-ch:
		return result, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("超时: %v", timeout)
	}
}

func WordCount(s string) map[string]int {
	words := strings.Fields(s)
	counts := make(map[string]int)
	for _, word := range words {
		counts[word]++
	}
	return counts
}

func fibonacci() func() int {
	a, b := 0, 1

	// 返回下一个斐波那契数的函数
	next := func() int {
		result := a
		a, b = b, a+b
		return result
	}
	return next
}

type MyReader struct{}

// TODO: 为 MyReader 添加一个 Read([]byte) (int, error) 方法。

func (R MyReader) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = 'A'
	}

	return len(b), nil
}

// List 表示一个可以保存任何类型的值的单链表。
type List[T any] struct {
	next *List[T]
	val  T
}

func newNode[T any](val T) *List[T] {
	return &List[T]{val: val, next: nil}
}

func (head *List[T]) InsertAtHead(val T) *List[T] {
	newNode := newNode(val)
	newNode.next = head
	return newNode
}

func (head *List[T]) InsertAtEnd(val T) *List[T] {
	newNode := newNode(val)
	if head == nil {
		return newNode
	}
	for cur := head; cur != nil; cur = cur.next {
		if cur.next == nil {
			cur.next = newNode
			break
		}
	}
	return head
}

func bufferTest() {
	ch := make(chan int) // 无缓冲通道

	// 发送操作会阻塞，直到有接收方准备好
	go func() {
		fmt.Println("发送前")
		ch <- 42 // 阻塞直到主goroutine接收
		fmt.Println("发送后")
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("接收前")
	val := <-ch // 接收数据
	fmt.Println("接收后，值:", val)
}
