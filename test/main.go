package main

import "fmt"

func echo(args ...string) {
	fmt.Println(args)
}

func main() {
	echo()

	echo("11")
	echo("22")
}
