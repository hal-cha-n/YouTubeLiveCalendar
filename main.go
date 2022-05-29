package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	developer_key := get_developer_key()
	youtube_search_list := search_youtube_list(developer_key)
	fmt.Println(youtube_search_list)
}

func get_developer_key() string {
	developer_key := os.Getenv("YOUTUBE_KEY")
	return developer_key
}

func search_youtube_list(developer_key string) string {
	url := "https://www.googleapis.com/youtube/v3/search"

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	//クエリパラメータ
	params := request.URL.Query()
	params.Add("key", developer_key)
	params.Add("q", "洋楽")
	params.Add("part", "snippet, id")
	params.Add("maxResults", "1")

	request.URL.RawQuery = params.Encode()

	timeout := time.Duration(5 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(string(body))
	return string(body)
}
