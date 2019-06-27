package main

import "fmt"

// Animal 动物类
type Animal interface {
	say()
	Runner
}

// Runner 跑接口
type Runner interface {
	run()
}

// Dog 狗类
type Dog struct {
	Name string
	Cat
}

func (dog Dog) say() {
	fmt.Printf("dog say: my name is %s\n", dog.Name)
}
func (dog Dog) run() {
	fmt.Printf("dog %s is running\n", dog.Name)
}

// Cat 猫类
type Cat struct {
	Name string
	Age  int
}

func (cat Cat) say() {
	fmt.Printf("cat say: my name is %s, age is %d\n", cat.Name, cat.Age)
}
func (cat Cat) run() {
	fmt.Printf("cat %s is running\n", cat.Name)
}
func main() {
	animals := []Animal{
		Dog{"dog1", Cat{"cat2", 10}},
		Cat{"cat1", 10},
	}
	for _, animal := range animals {
		animal.say()
		animal.run()
	}
}
