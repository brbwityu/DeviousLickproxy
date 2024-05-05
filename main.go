package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var cookie = ""

var snipehook string = ""
var discordid string = ""

var workercount int = 0

var profitmargin float64 = 0

var ids []int64

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

var rap = make(map[int64]interface{})

func getCSRF() string {
	// Create an HTTP client
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	csrfURL := "https://auth.roblox.com/v2/login"

	// Create a new POST request to the login endpoint
	req, err := http.NewRequest("POST", csrfURL, nil)
	if err != nil {
		fmt.Println(err)
	}

	cookie := &http.Cookie{Name: ".ROBLOSECURITY", Value: cookie}

	// Add the cookie to the request headers
	req.AddCookie(cookie)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	// Find the X-CSRF-TOKEN in the response headers
	csrfToken := resp.Header.Get("X-CSRF-TOKEN")
	if csrfToken == "" {
		fmt.Println(err)
	}

	return csrfToken
}

func loadConfig() {
	file, err := os.Open("config.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)

		// Check if parts has at least 2 elements
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "cookie":
			cookie = value
		case "snipehook":
			snipehook = value
		case "discordid":
			discordid = value
		case "workercount":
			id, _ := strconv.ParseInt(value, 10, 64)
			workercount = int(id)
		case "profitmargin":
			id, _ := strconv.ParseFloat(value, 64)
			profitmargin = id
		case "ids":
			idStrings := strings.Split(value, ",")
			for _, idString := range idStrings {
				id, err := strconv.ParseInt(strings.TrimSpace(idString), 10, 64)
				if err != nil {
					panic(err)
				}
				ids = append(ids, id)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func getRandomCookie() string {
	file, err := os.Open("falsecookies.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	randomLine := lines[r.Intn(len(lines))]

	return randomLine
}

func getRandomProxy() string {
	file, err := os.Open("proxies.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	randomLine := lines[r.Intn(len(lines))]

	return randomLine
}

func sendWebhook(itemid string, price string, statuscode int, reason string) {
	webhookURL := snipehook
	content := `{
		"content": "<@` + discordid + `> Successfully sniped item https://www.roblox.com/catalog/` + itemid + ` at the price of: ` + price + ` robux!"
	}`

	if statuscode != 200 {
		content = `{
		"content": "Failed to snipe item https://www.roblox.com/catalog/` + itemid + ` at the price of: ` + price + ` robux! Reason: ` + reason + `"
	}`
	}

	req, _ := http.NewRequest("POST", webhookURL, strings.NewReader(content))
	req.Header.Set("Content-Type", "application/json")

	http.DefaultClient.Do(req)
}

func createTransportWithProxy() (*http.Transport, error) {

	// Get a random proxy
	proxy := getRandomProxy()

	// Split the proxy details
	proxyDetails := strings.Split(proxy, ":")
	if len(proxyDetails) < 4 {
		return nil, fmt.Errorf("invalid proxy format")
	}

	// Define the proxy url
	proxyURL, err := url.Parse(fmt.Sprintf("http://%s:%s", proxyDetails[0], proxyDetails[1]))
	if err != nil {
		return nil, err
	}

	// Set up the proxy credentials
	proxyURL.User = url.UserPassword(proxyDetails[2], proxyDetails[3])

	// Setup HTTP Transport with the proxy
	return &http.Transport{Proxy: http.ProxyURL(proxyURL)}, nil
}

func snipeItem(jsonPayload []byte, productId int64, idstring string, priceAttr string) {
	targetURL, _ := url.Parse("https://economy.roblox.com/v1/purchases/products/" + strconv.FormatInt(productId, 10))

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	cookies := []*http.Cookie{
		{Name: ".ROBLOSECURITY", Value: cookie},
		{Name: "x-csrf-token", Value: getCSRF()},
		{Name: "_gcl_au", Value: "1.1.435265357.1710645925"},
		{Name: "GuestData", Value: "UserID=-999927455"},
		{Name: "authority", Value: "auth.roblox.com"},
		{Name: "content-type", Value: "application/json"},
		{Name: "accept", Value: "application/json, text/plain, */*"},
	}
	jar.SetCookies(targetURL, cookies)

	// Initialize HTTP Transport with proxy
	tr, err := createTransportWithProxy()
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Transport: tr, // Attaching transport with proxy to client
		Jar:       jar,
	}

	secondRequest, err := http.NewRequest("POST", targetURL.String(), bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Fatal(err)
	}

	secondRequest.Header.Set("authority", "auth.roblox.com")
	secondRequest.Header.Set("accept", "application/json, text/plain, */*")
	secondRequest.Header.Set("content-type", "application/json")
	secondRequest.Header.Set("_gcl_au", "1.1.435265357.1710645925")
	secondRequest.Header.Set("GuestData", "UserID=-999927455")
	secondRequest.Header.Set("x-csrf-token", getCSRF())

	req, err := client.Do(secondRequest)
	if err != nil {
		log.Fatal(err)
	}

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Always close the response body
	defer req.Body.Close()

	bodyString := string(bodyBytes)
	fmt.Println(bodyString)

	var dat map[string]interface{}

	// Decoding/Unmarshalling the JSON string
	if err := json.Unmarshal([]byte(bodyString), &dat); err != nil {
		log.Fatal(err)
	}

	// Getting the value of "purchased" from the map
	purchased, ok := dat["purchased"]
	if !ok {
		log.Fatal("Key 'purchased' not found in JSON")
	}

	fmt.Println(purchased)
	var statuscode int = 400

	if purchased.(bool) == true {
		statuscode = 200
	}

	reason, ok := dat["reason"].(string)
	if !ok {
		log.Fatal("reason attribute is not a string!")
	}

	sendWebhook(idstring, priceAttr, statuscode, reason)
}

func getRecentAveragePrices() {
	for _, id := range ids {
		var newcookie string = ""
		jar, err := cookiejar.New(nil)
		if err != nil {
			log.Fatal(err)
		}

		// Initialize HTTP Transport with proxy
		tr, err := createTransportWithProxy()
		if err != nil {
			log.Fatal(err)
		}

		client := &http.Client{
			Jar:       jar,
			Transport: tr, // Attaching transport with proxy to client
		}

		client.Jar.SetCookies(&url.URL{
			Scheme: "https",
			Host:   "www.roblox.com",
		}, []*http.Cookie{
			{
				Name:  ".ROBLOSECURITY",
				Value: newcookie,
			},
		})

		req, err := http.NewRequest("GET", "https://economy.roblox.com/v1/assets/"+strconv.FormatInt(id, 10)+"/resale-data", nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Cookie", ".ROBLOSECURITY="+newcookie)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			log.Fatal(err)
		}

		averagePrice := data["recentAveragePrice"].(float64)
		rap[id] = int64(averagePrice)
		strNum := strconv.Itoa(int(id))
		stringNum := strconv.Itoa(int(averagePrice))
		fmt.Println("ID: " + string(strNum) + ", RAP: " + stringNum)
	}
}

func checkId(id int64) (string, error) {
	var rv string = ""
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", err
	}

	var idstring = strconv.FormatInt(id, 10)

	url, _ := url.Parse("https://www.roblox.com/catalog/" + idstring)

	cookies := []*http.Cookie{
		&http.Cookie{Name: ".ROBLOSECURITY", Value: getRandomCookie()},
	}
	jar.SetCookies(url, cookies)

	// Initialize HTTP Transport with proxy
	tr, err := createTransportWithProxy()
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Jar:       jar,
		Transport: tr, // Attaching transport with proxy to client
	}

	firstRequest, err := http.NewRequest("GET", url.String(), strings.NewReader("1"))

	if err != nil {
		return "", err
	}

	response, err := client.Do(firstRequest)

	// ... rest of the function ...

	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", err
	}

	attributes := make(map[string]interface{})

	var productId int64 = 0

	doc.Find("div.content>div.page-content").Each(func(i int, s *goquery.Selection) {
		price, exists := s.Attr("data-expected-price")

		if exists {
			if reflect.TypeOf(price).Kind() == reflect.String {
				attributes["expectedPrice"] = price
				attributes["expectedCurrency"] = "1"
			}
		}

		seller, exists := s.Attr("data-expected-seller-id")

		if exists {
			if reflect.TypeOf(seller).Kind() == reflect.String {
				attributes["expectedSellerId"] = seller
			}
		}

		productid, exists := s.Attr("data-product-id")

		if exists {
			newid, err := strconv.ParseInt(productid, 10, 64)
			checkErr(err)

			productId = newid
		}

		lowestassetid, exists := s.Attr("data-lowest-private-sale-userasset-id")

		if exists {
			if reflect.TypeOf(lowestassetid).Kind() == reflect.String {
				attributes["userAssetId"] = lowestassetid
			}
		}

		// add more attributes as needed
	})

	//convert the map to a JSON
	jsonPayload, err := json.Marshal(attributes)
	if err != nil {
		log.Fatal(err)
	}

	//Test for snipe level
	if attributes["expectedPrice"] != nil {
		priceAttr, ok := attributes["expectedPrice"].(string)

		if !ok {
			log.Fatal("expectedPrice attribute is not a string!")
		}

		newval, _ := strconv.ParseInt(priceAttr, 10, 64)

		valueInterface, exist := rap[id]
		value, ok := valueInterface.(int64)
		value = value

		if !exist {
			fmt.Println("ID does not exist in map!")
		}

		fmt.Println(newval)

		if float64(newval) != float64(0) {
			if float64(newval) <= (profitmargin * float64(rap[id].(int64))) {
				go snipeItem(jsonPayload, productId, idstring, priceAttr)
				go snipeItem(jsonPayload, productId, idstring, priceAttr)
				go snipeItem(jsonPayload, productId, idstring, priceAttr)
			}
		}
	}
	return rv, nil
	// now jsonPayload contains your desired JSON payload
}

func main() {
	fmt.Println("Getting config and initializing Devious Lick")
	time.Sleep(time.Second * 1)

	loadConfig()

	fmt.Println("Getting proxies...")
	time.Sleep(time.Second * 1)
	fmt.Println("Proxies retrieved.")
	time.Sleep(time.Second * 2)
	fmt.Println("Retrieving RAP Info.")
	time.Sleep(time.Second * 1)

	getRecentAveragePrices()

	time.Sleep(time.Second * 1)

	fmt.Println("RAP Info retrieved. Devious Lick initialized. Starting Devious Lick.")
	time.Sleep(time.Second * 3)

	for true {
		var wg sync.WaitGroup
		workerCount := workercount
		jobs := make(chan int64, len(ids)) // changed here

		// start worker goroutines
		for i := 1; i <= workerCount; i++ {
			go func(i int) {
				for id := range jobs {
					checkId(id)
					wg.Done() // signal job completion
				}
			}(i)
		}

		// distribute jobs
		for _, id := range ids {
			jobs <- id
			wg.Add(1) // add a job to the wait group
		}

		close(jobs) // signal no more jobs
		wg.Wait()   // wait for all jobs to finish
		fmt.Println("Loop Finished!")
	}
}
