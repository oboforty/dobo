package main

import "regexp"

func main() {
	r, _ := regexp.Compile("^(.*)-([0-9]+).db$")

	match := r.FindStringSubmatch("fostaliga-123-2_24-13.db")

	println(match[2])
}
