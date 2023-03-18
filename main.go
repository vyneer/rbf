package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
)

var rumbleBase64 string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
var rumbleBase64Split []string = strings.Split(rumbleBase64, "")
var rumbleBase64SplitLength int = len(rumbleBase64Split)
var rumbleQuality []string = []string{"oaa", "baa", "caa", "gaa", "haa", "oaa.rec", "baa.rec", "caa.rec", "gaa.rec", "haa.rec"}
var rumbleCDNs []string = []string{"sp.rmbl.ws"}

func PowInts(x, n int) int {
	if n == 0 {
		return 1
	}
	if n == 1 {
		return x
	}
	y := PowInts(x, n/2)
	if n%2 == 0 {
		return y * y
	}
	return x * y * y
}

func rumbleEncode(i int) string {
	buf := ""
	count := 0
	for PowInts(rumbleBase64SplitLength, count) < i {
		index := (i / PowInts(rumbleBase64SplitLength, count)) % rumbleBase64SplitLength
		buf += rumbleBase64Split[index]
		count++
	}
	return buf
}

func checkAvailability(vid string, randomBit int, channel chan string) {
	vidSplit := strings.Split(vid, "")
	for _, v := range rumbleCDNs {
		for _, q := range rumbleQuality {
			url := fmt.Sprintf("https://%s/s8/%d/%s/%s/%s/%s/%s.%s.mp4", v, randomBit, vidSplit[0], vidSplit[1], vidSplit[2], vidSplit[3], vid, q)
			resp, err := http.Get(url)
			if err == nil && resp.StatusCode == 200 {
				channel <- url
			}
		}
	}
}

func checkAllLinks(vid string) []string {
	var wg sync.WaitGroup

	buf := []string{}

	results := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
			}
			checkAvailability(vid, i, results)
			cancel()
		}(i)
	}

	go func(chan string, *[]string) {
		for v := range results {
			buf = append(buf, v)
		}
	}(results, &buf)

	wg.Wait()

	return buf
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Please provide only one argument - a rumble.com public URL (not the embed URL)")
		return
	}

	url, err := url.Parse(os.Args[1])
	if err != nil {
		fmt.Println("Wasn't able to parse the provided URL")
	}

	publicId := strings.TrimPrefix(strings.Split(url.Path, "-")[0], "/v")

	parsed, err := strconv.ParseInt(publicId, 36, 64)
	fmt.Println(parsed)
	if err != nil {
		fmt.Println("Wasn't able to convert the provided video ID to base 36")
	}
	vid := rumbleEncode(int(parsed))

	fmt.Printf("Rumble CDN ID for the provided link - %s\n", vid)

	fmt.Println("Checking availability...")
	availableLinks := checkAllLinks(vid)
	for _, link := range availableLinks {
		fmt.Println(link)
	}
}
