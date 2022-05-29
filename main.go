package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

func newClient() *http.Client {
	client := &http.Client{
		Transport: &transport.APIKey{Key: os.Getenv("YOUTUBE_KEY")},
	}
	return client
}

func newYoutubeService(client *http.Client) *youtube.Service {
	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Unable to create YouTube service: %v", err)
	}

	return service
}

func printChannelInfo(channelID string) {
	service := newYoutubeService(newClient())
	call := service.Search.List([]string{"snippet", "id"}).
		ChannelId(channelID).
		Order("date")

	response, err := call.Do()
	if err != nil {
		log.Fatalf("%v", err)
	}

	for i, v := range response.Items {
		fmt.Printf("%d. %+v: %+v\n", i, v.Snippet.Title, v.Snippet.PublishedAt)
	}
}

func main() {
	printChannelInfo("UCdre9A9clPahkJBdlKMCZpw")
}

func get_developer_key() string {
	developer_key := os.Getenv("YOUTUBE_KEY")
	return developer_key
}
