package main

import (
	"fmt"

	"github.com/gonzalo-bulnes/net/http"
)

func main() {
	resp, err := http.Get("http://magpie.surge.sh/index.html")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Response: %+v\n", resp)
}
