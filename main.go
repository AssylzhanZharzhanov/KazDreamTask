package main

import (
	"encoding/csv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func makeRequest(url string, channel chan []string, wg *sync.WaitGroup) {
	log.Print(url)
	start := time.Now()

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	elapsed := time.Since(start).Seconds()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
	}

	length := len(string(body))
	status := res.StatusCode
	channel <- []string{url, strconv.Itoa(status), strconv.Itoa(length), strconv.FormatFloat(elapsed, 'f', 6, 64)}
	wg.Done()
}

func saveToCSV(name string, data [][]string) {

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

	urls := []string{"https://habr.com/ru/post/215117/", "https://ru.wikipedia.org/wiki/HTTP", "https://developer.mozilla.org/ru/docs/Web/HTTP",
		"https://ru.bmstu.wiki/HTTP_(Hypertext_Transfer_Protocol)", "https://proselyte.net/tutorials/http-tutorial/introduction/", "https://www.speedcheck.org/ru/wiki/http/",
		"http://pki.gov.kz/index.php/ru/ncalayer", "http://www.edu.gov.kz/", "http://adilet.zan.kz/rus", "https://wiki.merionet.ru/servernye-resheniya/3/protocol-http/", "https://www.opennet.ru/docs/RUS/http/",
		"https://www.w3.org/Protocols/HTTP/1.1/rfc2616bis/draft-lafon-rfc2616bis-03.html", "https://flaviocopes.com/http/", "https://www.extrahop.com/resources/protocols/http/",
	}

	data := [][]string{{"url", "status", "size", "time"}}
	c := make(chan []string)
	goroutines := 0

	var wg sync.WaitGroup

	for _, i := range urls {
		wg.Add(1)
		goroutines = goroutines + 1
		go makeRequest(i, c, &wg)
	}

	for i := 0; i < len(urls); i++ {
		data = append(data, <-c)
	}

	saveToCSV("test.csv", data)

	elapsed := time.Since(start).Seconds()

	log.Print("Goroutines: ", goroutines)
	log.Print("Time elapsed: ", elapsed)

	// wg.Wait()
}
