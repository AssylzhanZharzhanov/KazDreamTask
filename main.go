package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func makeRequest(url string, channel chan []string, wg *sync.WaitGroup) {

	start := time.Now()

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		res.Body.Close()
		wg.Done()
	}()

	elapsed := time.Since(start).Seconds()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Print(err)
	}

	length := len(string(body))
	status := res.StatusCode

	if status != 200 {
		fmt.Println("Cannot reach this URL, status: ", status)
	}
	channel <- []string{url, strconv.Itoa(status), strconv.Itoa(length), strconv.FormatFloat(elapsed, 'f', 6, 64)}
}

func saveToCSV(name string, c chan []string) {
	fmt.Println("Saving file ...")
	data := [][]string{}

	for val := range c {
		data = append(data, val)
	}

	f, err := os.Create(name)
	if err != nil {
		log.Fatalf("Cannot open '%s': %s\n", name, err.Error())
	}

	defer func() {
		e := f.Close()
		if e != nil {
			log.Fatalf("Cannot close '%s': %s\n", name, e.Error())
		}
	}()

	w := csv.NewWriter(f)

	err = w.WriteAll(data)
	if err != nil {
		log.Print(err)
	}
}

func main() {
	start := time.Now()
	reader := bufio.NewReader(os.Stdin)
	var wg sync.WaitGroup
	c := make(chan []string)
	msg := make(chan string)
	killSignal := make(chan os.Signal, 1)
	signal.Notify(killSignal, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Enter URL")
	goroutinesCount := 0
	go func() {
		for {
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\r\n", "", -1)
			if text != "" {
				msg <- text
			}
		}
	}()

loop:
	for {
		select {
		case <-killSignal:
			fmt.Println("Exiting the program...")
			go func() {
				defer close(c)
				wg.Wait()
			}()
			saveToCSV("test2.csv", c)
			fmt.Println("Total time: ", time.Since(start))
			fmt.Println("Total requests: ", goroutinesCount)
			break loop
		case s := <-msg:
			wg.Add(1)
			go makeRequest(s, c, &wg)
			goroutinesCount++
		}
	}
}
