package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

func GetCount(myurl string) (count int, err error) {
	if _, Err := url.ParseRequestURI(myurl); err != nil {
		return 0, Err
	}

	resp, err := http.Get(myurl)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, errors.New("wrong http server status code")
	}

	scanner := bufio.NewScanner(resp.Body)
	var cnt int
	for scanner.Scan() {
		cnt += strings.Count(scanner.Text(), "Go")
	}

	if Err := scanner.Err(); Err != nil {
		return 0, Err
	}

	return cnt, nil
}

func worker(c chan string, wg *sync.WaitGroup, total *uint64) {
	defer wg.Done()
	for {
		myurl, more := <-c
		if more {
			count, err := GetCount(myurl)
			if err == nil {
				atomic.AddUint64(total, uint64(count))
				fmt.Printf("Count for %s: %d\n", myurl, count)
			} else {
				fmt.Println("Count for", myurl, ": 0, error", err)
			}
		} else {
			return
		}
	}
}

func main() {
	const maxGoroutines = 5
	var wg sync.WaitGroup
	var total uint64
	urls := make(chan string)
	var urlsCount int

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if urlsCount < maxGoroutines {
			wg.Add(1)
			go worker(urls, &wg, &total)
		}
		urls <- scanner.Text()
		urlsCount++
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("reading standard input:", err)
	}
	close(urls)
	wg.Wait()
	fmt.Println("Total:", total)
}
