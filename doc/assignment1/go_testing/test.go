package main

import (
	"fmt"
	"time"
)

func writeAfterReadCannels(taskStatus chan int) {
	// var x int
	fmt.Println("task is started")
	select {
	case x := <-taskStatus:
		fmt.Println("task is completed", x)
		time.Sleep(time.Second * 2)
		taskStatus <- x
		return
	case <-time.After(time.Second * 10):
		fmt.Println("task haven't completed")
		return
	}
}

// func quitLockAfterAwhile(mutex sync.Mutex) {
// 	// mutex := &sync.Mutex{}

// 	// 在等待超时前，尝试获取锁
// 	select {
// 	case mutex.Lock():
// 		fmt.Println("task is completed", x)
// 		time.Sleep(time.Second * 2)
// 		mutex.Unlock()
// 		return
// 		// 成功获取锁后的代码逻辑
// 	case <-time.After(time.Duration(5) * time.Second):
// 		fmt.Println("task haven't completed")
// 		// 等待5秒后仍未获取到锁，执行超时后的代码逻辑
// 		return
// 	}
// }

func main() {
	// taskStatusChan := make(chan int)
	// var mu sync.Mutex
	// go quitLockAfterAwhile(mu)
	// go quitLockAfterAwhile(mu)
	// taskStatusChan <- 6666
	// for {
	// 	time.Sleep(time.Second)
	// 	select {
	// 	case x := <-taskStatusChan:
	// 		fmt.Println("main process receive x", x)
	// 		taskStatusChan <- x
	// 	case <-time.After(time.Millisecond):
	// 		fmt.Println("main process give up waiting for that chan")
	// 	}
	// }

	// ch := make(chan string, 1)

	// go func ()  {
	// 	ch <- "world"
	// 	fmt.Println("hello")
	// 	ch <- "world2"
	// 	fmt.Println("hello2")
	// }()
	// time.Sleep(time.Second * 3)
	// fmt.Println(<-ch)
	// time.Sleep(time.Second * 3)
	// fmt.Println(<-ch)

	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)
	ch1 <- "hello"

	select {
	case a := <-ch1:
		fmt.Println(a)
		ch2 <- "world"
	case b := <-ch2:
		fmt.Println(b)
	}
	b := <-ch2
	fmt.Println(b)
}
