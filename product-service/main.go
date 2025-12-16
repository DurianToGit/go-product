package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("输出环境")
	fmt.Println(os.Getenv("DB_USER"))
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
