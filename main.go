package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/api/calendar/v3"
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

func newCalenderService() *calendar.Service {
	ctx := context.Background()
	service, err := calendar.NewService(ctx)
	if err != nil {
		log.Fatalf("Unable to create YouTube service: %v", err)
	}

	return service
}

func printChannelInfo(channelID string) {
	service := newYoutubeService(newClient())
	call := service.Search.List([]string{"snippet", "id"}).
		ChannelId(channelID).
		Order("date").
		MaxResults(1)

	response, err := call.Do()
	if err != nil {
		log.Fatalf("%v", err)
	}

	for i, v := range response.Items {
		fmt.Printf("%d. %+v\n", i, v.Snippet.Title)
		youtube_video_details(v.Id.VideoId)
	}
}

func youtube_video_details(videoId string) {
	service := newYoutubeService(newClient())
	call := service.Videos.List([]string{"snippet", "liveStreamingDetails"}).
		Id(videoId)

	response, err := call.Do()
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Printf("配信予定時刻：%+v\n", response.Items[0].LiveStreamingDetails.ScheduledStartTime)
	createEvent(response.Items[0])
}

func createEvent(liveDetail *youtube.Video) {

	startTime, _ := time.Parse(time.RFC3339Nano, liveDetail.LiveStreamingDetails.ScheduledStartTime)
	endTime := startTime.Add(1 * time.Hour)

	event := &calendar.Event{
		Summary:     liveDetail.Snippet.Title,
		Description: liveDetail.Snippet.Description,
		Start: &calendar.EventDateTime{
			DateTime: liveDetail.LiveStreamingDetails.ScheduledStartTime,
			TimeZone: "Europe/London",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339Nano),
			TimeZone: "Europe/London",
		},
	}

	service := newCalenderService()
	call := service.Events.Insert("m323k3iij27jdlq1m0qvoppldo@group.calendar.google.com", event)

	_, err := call.Do()
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Println("success")
}

func main() {
	printChannelInfo("UCdre9A9clPahkJBdlKMCZpw")
}

func get_developer_key() string {
	developer_key := os.Getenv("YOUTUBE_KEY")
	return developer_key
}
