package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
)

/*
1. count number of words
2. count number of unique words
3. count frequency of each words
*/

const url = "https://www.gutenberg.org/files/2701/2701-0.txt"

func main() {
	// file := os.Args[1]

	// textBytes, err := os.ReadFile(file)
	// if err != nil {
	// 	slog.Error("failed to read file", slog.String("file", file))
	// }
	// slog.Info("read text file", slog.String("file", file))

	// text := string(textBytes)

	text, err := getText(url)
	if err != nil {
		slog.Error("failed to get text", slog.Any("error", err))
		os.Exit(1)
	}

	m := make(map[string]int)
	words := strings.SplitSeq(text, " ")
	numWorkers := 5
	q := make(chan string)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for w := range numWorkers {
		wg.Add(1)
		slog.Info("starting worker", slog.Int("number", w))
		go worker(&wg, q, &m, &mu)
	}

	for i := range words {
		q <- i
	}
	close(q)
	slog.Info("words feeded completely")

	wg.Wait()
	slog.Info("workers done")

	fmt.Printf("unique words: %d\n", len(m))
}

func worker(wg *sync.WaitGroup, q <-chan string, m *map[string]int, mu *sync.Mutex) {
	defer wg.Done()

	for i := range q {
		mu.Lock()
		if _, ok := (*m)[i]; !ok {
			(*m)[i] = 0
		}
		(*m)[i]++
		mu.Unlock()
	}
}

func getText(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get from url: %w", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful (status code 200 OK)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status code: %d %s", resp.StatusCode, resp.Status)
	}

	// Read the entire response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	newBody := make([]byte, 0)
	for i := 0; i < 50; i++ {
		newBody = append(newBody, body...)
	}

	return string(newBody), nil
}

/*
cpu's only job is to compute: add, subtract, ...
accessing map is not cpu's job
map is in RAM
cpu <--> RAM travel is the bottleneck
if we don't use concurrency
there can be so much cpu idle time, waiting for the RAM data
*/

// concurrent data access
/*
first, request access
second, only one request is permitted at a point of time
third, after permitted, read or write to it
fourth, after using it, abandon access, so other people's request can be permitted

this pattern is called mutual exclusion, in short, mutex
*/

// worker pool pattern
/*
q = [word1, word2, word3, ..., word1000]

10 workers all listening to the queue
each worker can be idle or busy
idle workers grab new word from queue


channel
queue designed for concurrent access

*/

// const numWorkers = 5
// const numWorks = 20

// func main() {
// 	q := make(chan int, numWorks)

// 	var wg sync.WaitGroup
// 	for i := 0; i < numWorkers; i++ {
// 		wg.Add(1)
// 		go worker(&wg, q)
// 	}

// 	for i := range numWorks {
// 		q <- i
// 	}

// 	close(q)
// 	wg.Wait()
// }

// // worker has to listen to q if idle, and don't listen to q if busy
// // after done with one word, it should turn idle
// // therefore we need infinite loops
// // but then, how do we break out of it?
// func worker(wg *sync.WaitGroup, q <-chan int) {
// 	for i := range q {
// 		time.Sleep(2 * time.Second)
// 		fmt.Println(i, "finished")
// 	}
// 	wg.Done()
// }
