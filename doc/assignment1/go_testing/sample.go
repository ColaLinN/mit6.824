package main

import (
	"fmt"
	"sync"
	"time"
)

func setStatus(status chan string, lock *sync.Mutex) {
	time.Sleep(10 * time.Second)
	lock.Lock()
	defer lock.Unlock()
	status <- "fail"
}

func main() {
	sharedData := "initial"
	fmt.Println(sharedData)

	lock := sync.Mutex{}
	statusChan := make(chan string)
	go setStatus(statusChan, &lock)

	lock.Lock() // 加锁以防止其他goroutine更改状态
	defer lock.Unlock()
	time.Sleep(11 * time.Second)

	select {
	case newStatus := <-statusChan:
		fmt.Println(newStatus)
	default: // 如果在11s内未收到新状态，则打印旧状态
		fmt.Println(sharedData)
	}
}
