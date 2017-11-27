package main

import (
	"fmt"
	"sync"
)

type Visited struct {
	visited map[string]bool
	mux sync.Mutex
    tovisit chan string
    finished chan bool
}

func (v *Visited) Visit(url string) {
	v.mux.Lock()
	v.visited[url] = true
	v.mux.Unlock()
}

func (v *Visited) Finish() {
    v.finished <- true
}

func (v *Visited) HasBeen(url string) bool {
	v.mux.Lock()
	defer v.mux.Unlock()
	return v.visited[url]
}

type Fetcher interface {
	Fetch(url string) (body string, urls []string, err error)
}

func Crawl(url string, visited Visited, fetcher Fetcher) {
	visited.Visit(url)
	defer visited.Finish()
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
    fmt.Printf("found: %s %q\n", url, body)
	for _, u := range urls {
		if (!visited.HasBeen(u)) {
			visited.tovisit <- u
		}
	}
	return
}

func main() {
	visited := Visited{visited: make(map[string]bool), tovisit: make (chan string, 100), finished: make (chan bool, 10)}
    visited.tovisit <- "http://golang.org/"
    threads := 0
    for {
        select {
        case tovisit := <- visited.tovisit:
            threads++
   	        go Crawl(tovisit, visited, fetcher)
        case <- visited.finished:
            threads --
            if threads <=0 {
                return
            }
        }
    }
}

type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
