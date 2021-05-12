package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	resp, err := http.Get("http://example.com/")
	if err != nil {
		fmt.Println("=== Network is not available upon start: FAILED === ")
		panic(err)
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("=== Network is available upon start: SUCCEEDED === ")
	for {
		time.Sleep(1 * time.Hour)
	}
}
