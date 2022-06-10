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
	"os"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/youtube/v3"
)

func liveEndTime(liveStartTime string) string {
	parsedLiveStartTime, _ := time.Parse(time.RFC3339, liveStartTime)
	return parsedLiveStartTime.Add(1 * time.Hour).Format(time.RFC3339) // 配信時間はYouTubeから取得できないため、1時間とする。
}

func newYoutubeService() *youtube.Service {
	ctx := context.Background()
	service, err := youtube.NewService(ctx)
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

func getChannelInfo(channelID string) []*youtube.SearchResult {
	service := newYoutubeService()
	call := service.Search.List([]string{"snippet", "id"}).
		Type("video").
		EventType("upcoming").
		ChannelId(channelID).
		Order("date")

	response, err := call.Do()
	if err != nil {
		log.Fatalf("%v", err)
	}

	return response.Items
}

func getVideoDetail(videoId string) *youtube.Video {
	service := newYoutubeService()
	call := service.Videos.List([]string{"snippet", "liveStreamingDetails"}).
		Id(videoId)

	response, err := call.Do()
	if err != nil {
		log.Fatalf("%v", err)
	}

	return response.Items[0]
}

func getEvents(liveDetail *youtube.Video) *calendar.Events {
	startTime := liveDetail.LiveStreamingDetails.ScheduledStartTime
	service := newCalenderService()
	call := service.Events.List(os.Getenv("YLC_CALENDAR_ID")).
		TimeMin(startTime).
		TimeMax(liveEndTime(startTime))

	response, err := call.Do()
	if err != nil {
		log.Fatalf("%v", err)
	}
	return response
}

func createEventOrUpdate(liveDetail *youtube.Video) *calendar.Event {
	events := getEvents(liveDetail)
	startTime := liveDetail.LiveStreamingDetails.ScheduledStartTime
	description := "試聴はこちらから: https://www.youtube.com/watch?v=" + liveDetail.Id + "\n\n" + liveDetail.Snippet.Description
	service := newCalenderService()

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

	for i, v := range events.Items {
		fmt.Printf("  同時刻の配信予定[%d]： %+v\n", i, v.Summary)
	}

	// 特定のIDを含む場合
	for i, v := range events.Items {
		if strings.Contains(v.Description, liveDetail.Id) {
			call := service.Events.Update(os.Getenv("YLC_CALENDAR_ID"), v.Id, event)
			response, err := call.Do()
			if err != nil {
				log.Fatalf("%v", err)
			}
			fmt.Printf("  [%d]の予定と一致したため予定を上書きします。\n", i)
			return response
		}
	}

	// 新しい予定として作成する。
	call := service.Events.Insert(os.Getenv("YLC_CALENDAR_ID"), event)
	response, err := call.Do()
	if err != nil {
		log.Fatalf("%v", err)
	}
	return response
}

func main() {
	channelInfos := getChannelInfo(os.Getenv("YLC_CHANNEL_ID"))
	fmt.Printf("%d件の配信予定\n", len(channelInfos))

	for _, channelInfo := range channelInfos {
		fmt.Printf("取得配信: %+v\n", channelInfo.Snippet.Title)

		videoDetail := getVideoDetail(channelInfo.Id.VideoId)
		fmt.Printf("  配信予定時刻: %+v\n", videoDetail.LiveStreamingDetails.ScheduledStartTime)

		event := createEventOrUpdate(videoDetail)
		fmt.Printf("  カレンダー登録完了: %s\n", event.HtmlLink)
	}
}
