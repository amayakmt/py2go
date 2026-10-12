package main

import "github.com/amayakmt/py2go/runtime"

func greet() {
	runtime.Print("hello")
}

func main() {
	defer runtime.Report()
	greet()
}
