package main

import (
	"errors"
	"fmt"
)

// type Test struct {
// 	*Car
// }

func main() {

	text := ""
	var err error
	defer func() {
		if err != nil {
			fmt.Println(err, text)
		}
		fmt.Println(err, text)
	}()

	x := 5
	if x > 3 {
		text = "valid"
		// err = errors.New("")
	} else {
		text = "tidak valid"
		err = errors.New("tidak valid")
	}
}
