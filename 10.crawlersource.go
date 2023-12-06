package main

import (
	"fmt"
	"sync"
	"time"
)

var (
	visitedSet = make(map[string]bool)
	wg         sync.WaitGroup
	signalling = make(chan struct{}, 1)
)

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	defer wg.Done()

	if depth <= 0 {
		return
	}

	signalling <- *new(struct{}) // new

	if visited, _ := visitedSet[url]; !visited {
		visitedSet[url] = true
		<-signalling // new

		body, urls, err := fetcher.Fetch(url)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("found:[depth:%d] %s %q\n", depth, url, body)

		for _, u := range urls {
			wg.Add(1)
			go Crawl(u, depth-1, fetcher)
		}
	} else {
		<-signalling // new
	}
}

func main() {
	wg.Add(1)
	Crawl("https://golang.org/", 4, fetcher)
	wg.Wait()
	fmt.Println("============DONE=============")

	for k := range visitedSet {
		fmt.Println(k)
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

// every url has a body + urls
type fakeResult struct {
	body string
	urls []string
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {

	fmt.Printf("Fetching: %s\n", url)

	time.Sleep(500 * time.Millisecond)

	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

///////////////////////////////////////////////////////////////////////////////////////////////
