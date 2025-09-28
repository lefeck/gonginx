package main

// 定义一个接口
type Animal interface {
	Speak() string
	Eat() string
}

// 定义实现结构体
type Dog struct{}
type Cat struct{}

// 实现接口方法
func (d Dog) Speak() string {
	return "Woof!"
}

func (d Dog) Eat() string {
	return "Dog is eating chicken."
}

func (c Cat) Speak() string {
	return "Meow!"
}

func (c Cat) Eat() string {
	return "Cat is eating fish."
}

func main() {
	animal := []Animal{Dog{}, Cat{}}
	for _, a := range animal {
		println(a.Speak())
		println(a.Eat())
	}

	dog := Dog{}
	println(dog.Speak())
	println(dog.Eat())

	cat := Cat{}
	println(cat.Speak())
	println(cat.Eat())
}
