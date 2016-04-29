package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kennygrant/sanitize"
)

var errBanned = errors.New("Banned")

func main() {
	var pos, prox int
	flag.IntVar(&pos, "pos", 0, "Posicao inicial do nome a buscar")
	flag.IntVar(&prox, "prox", 0, "Posicao inicial do proxy")
	flag.Parse()

	fmt.Println("Starting at pos:", pos)

	client, cookies, err := getContext(prox)
	if err != nil {
		log.Fatalf("Error getting cookies1: %s\n", err.Error())
	}

	log.Println("Cookies", cookies)

	for i := pos; i < len(names); i++ {
		domain := sanitize.Path(names[i])
		if len(domain) > 8 {
			continue
		}
		var ok bool
		var err error
		ok, err = check(domain, client, cookies)
		if err != nil {
			wait(30)
			i = i - 1
			var x int
			fmt.Println("Check for proxy ", proxies[x])
			for client, cookies, err = getContext(prox); err != nil; x++ {
				prox = prox + 1
				if x > 10 {
					log.Panicf("Too many banned requests, Ending on proxy %d pos %d", prox, i)
				}
			}

		}
		if ok {
			fmt.Printf("[%d] %s: %v\n", i, domain, ok)
		} else {
			fmt.Printf(".")
		}
		//wait(1)
	}
}

func setClientProxy(i int) (*http.Client, error) {

	if i > len(proxies) {
		return nil, fmt.Errorf("Ops, position %d is more than length of proxies %d", i, len(proxies))
	}

	proxyUrl, err := url.Parse(protocols[i] + "://" + proxies[i] + ":" + ports[i])
	if err != nil {
		return nil, err
	}

	fmt.Printf("Connecting proxy %s \n", proxyUrl.String())

	transport := &http.Transport{}
	transport.Proxy = http.ProxyURL(proxyUrl)                         // set proxy
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //set ssl
	client := &http.Client{Transport: transport}

	return client, nil
}

func getContext(prox int) ([]*http.Cookie, *http.Client, error) {
	/*
		// Not working, maybe because of current proxy
		client, err := setClientProxy(prox)
		if err != nil {
			log.Panicf("Error getting proxy number %d\n", prox)
		}
	*/

	req, err := http.NewRequest("GET", "https://www.register.ca/", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Error in NewRequest: %s", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection timed out") {
			return getContext(prox + 1)
		}
		return nil, nil, fmt.Errorf("Error in Do: %s", err)
	}

	return resp.Cookies(), client, nil
}

func check(domain string, cookies []*http.Cookie, client *http.Client) (bool, error) {

	domain = fmt.Sprintf("%s.ca", domain[0:len(domain)-2])

	//fmt.Printf("%s: ", domain)

	host := "https://www.register.ca/en/register.htm"

	form := url.Values{}

	form.Add("domain", domain)
	form.Add("Submit", "+SEARCH+NOW+")
	form.Add("check", "check")

	req, err := http.NewRequest("POST", host, strings.NewReader(form.Encode()))

	req.Header.Add("Host", "www.register.ca")
	req.Header.Add("Refer", "https://www.register.ca/")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("HERE COOKIES ERRROR: %v\nError: %v", cookies, err)
	}

	// use utfBody using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//html, _ := doc.Html()
	//fmt.Printf("RESP: %s\n", html)

	ok := false
	err = nil
	doc.Find(".greenPriceColor").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if text == "available" {
			ok = true
			return
		}

		// Never occours
		err = errBanned
		fmt.Printf("!!! ERROR BANNED: %s !!!", text)

	})

	//fmt.Printf("%v\n", ok)

	return ok, err
}

func wait(s int) {
	//fmt.Printf("Waiting %ds", s)
	s3 := time.Duration(s / 3 * 1000)

	time.Sleep(s3 * time.Millisecond)
	//fmt.Printf(".")
	time.Sleep(s3 * time.Millisecond)
	//fmt.Printf(".")
	time.Sleep(s3 * time.Millisecond)
	//fmt.Printf(".\n")
}
