package main

// 必須環境変数
// YLC_API_KEY: YouTube v3 API へのアクセスを許可したGCPのAPIキー
// YLC_CHANNEL_ID: 配信の予定を取得するYouTubeチャンネルのID
// YLC_CALENDAR_ID: 配信の予定を追加するGoogleカレンダーのID
// GOOGLE_APPLICATION_CREDENTIALS: Googleカレンダーへのアクセスを許可したサービスアカウントの認証情報

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

func liveEndTime(liveStartTime string) string {
	parsedLiveStartTime, _ := time.Parse(time.RFC3339, liveStartTime)
	return parsedLiveStartTime.Add(1 * time.Hour).Format(time.RFC3339) // 配信時間はYouTubeから取得できないため、1時間とする。
}

func newClient() *http.Client {
	client := &http.Client{
		Transport: &transport.APIKey{Key: os.Getenv("YLC_API_KEY")},
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
		log.Fatalf("Unable to create Calender service: %v", err)
	}

	return service
}

func getChannelInfo(channelID string) *youtube.SearchResult {
	service := newYoutubeService(newClient())
	call := service.Search.List([]string{"snippet", "id"}).
		ChannelId(channelID).
		Order("date").
		MaxResults(1)

	response, err := call.Do()
	if err != nil {
		log.Fatalf("%v", err)
	}

	return response.Items[0]
}

func getVideoDetail(videoId string) *youtube.Video {
	service := newYoutubeService(newClient())
	call := service.Videos.List([]string{"snippet", "liveStreamingDetails"}).
		Id(videoId)

	response, err := call.Do()
	if err != nil {
		log.Fatalf("%v", err)
	}

	return response.Items[0]
}

func createEvent(liveDetail *youtube.Video) *calendar.Event {

	startTime := liveDetail.LiveStreamingDetails.ScheduledStartTime
	description := "試聴はこちらから: https://www.youtube.com/watch?v=" + liveDetail.Id + "\n\n" + liveDetail.Snippet.Description

	event := &calendar.Event{
		Summary:     liveDetail.Snippet.Title,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: startTime,
			TimeZone: "Europe/London",
		},
		End: &calendar.EventDateTime{
			DateTime: liveEndTime(startTime),
			TimeZone: "Europe/London",
		},
	}

	service := newCalenderService()
	call := service.Events.Insert(os.Getenv("YLC_CALENDAR_ID"), event)

	response, err := call.Do()
	if err != nil {
		log.Fatalf("%v", err)
	}
	return response
}

func main() {
	channelInfo := getChannelInfo(os.Getenv("YLC_CHANNEL_ID"))
	fmt.Printf("取得配信: %+v\n", channelInfo.Snippet.Title)

	videoDetail := getVideoDetail(channelInfo.Id.VideoId)
	fmt.Printf("配信予定時刻：%+v\n", videoDetail.LiveStreamingDetails.ScheduledStartTime)

	event := createEvent(videoDetail)
	fmt.Printf("カレンダー登録完了: %s\n", event.HtmlLink)
}
