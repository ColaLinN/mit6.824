package main

import "fmt"
type Person struct {
    Name string
    Age int
}

// 传递指向结构体的指针
func updatePerson(p Person) {
    p.Age = 50
}

// 传递结构体本身
func printPerson(p Person) {
    fmt.Println(p.Name, p.Age)
}

func main() {
    // 创建一个 Person 结构体实例
    p := Person{"Alice", 25}

    // 传递指向结构体的指针
    updatePerson(p)

    // 传递结构体本身
    printPerson(p)
}