package user

import (
	"fmt"
	"time"
)

func Hello() {

	for {
		fmt.Println("hello")

		time.Sleep(time.Minute)
	}
}
