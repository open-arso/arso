package main

import (
	"fmt"
)

func main() {
	r := setupRouter()

	if err := r.Run(); err != nil {
	    fmt.Printf("Error running API: %v\n",err.Error())
	}
}
