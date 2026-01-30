package utils

import (
	"fmt"
	"net/http"
	"time"
)

func HttpWithRetry(f func(string) (*http.Response, error), url string) (*http.Response, error) {
	count := 10
	var response *http.Response
	var err error
	for i := 0; i < count; i++ {
		response, err = f(url)
		if err == nil {
			break
		}

		fmt.Printf("Error calling %v\n", url)
		time.Sleep(5 * time.Second)
	}

	return response, err
}
