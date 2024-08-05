package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
)

var (
	clientId     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	token        *oauth2.Token
)

var oauth2config *clientcredentials.Config

func main() {
	oauth2config = &clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     twitch.Endpoint.TokenURL,
	}
	tken, err := oauth2config.Token(context.Background())
	token = tken
	if err != nil {
		log.Fatal(err)
	}
	category_id := get_id()
	fmt.Println(category_id)
	streams := get_streams(category_id)
	numberOfStreams := len(streams)
	totalViewers := 0
	viewerCountList := make([]int, numberOfStreams)
	for i, stream := range streams {
		totalViewers += stream.Viewer_count
		viewerCountList[numberOfStreams-i-1] = stream.Viewer_count // assumnes we ge decreasing but want increasing
	}
	giniResult := gini(viewerCountList)
	fmt.Println(totalViewers)
	fmt.Println(giniResult)
	append_result(totalViewers, giniResult, len(viewerCountList))
}

func append_result(totalViewers int, giniCoeff float32, numberOfStreams int) {
	now := time.Now()
	f, _ := os.OpenFile(os.Args[1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
	defer f.Close()
	result := fmt.Sprintf("%d,%d,%d,%d,%d,%.3f\n", now.Weekday(), now.Hour(), now.Minute(), numberOfStreams, totalViewers, giniCoeff)
	f.WriteString(result)
}

// https://en.wikipedia.org/wiki/Gini_coefficient#Alternative_expressions
func gini(viewerCountList []int) float32 {
	previous := -1
	iyi := 0
	yi := 0
	n := len(viewerCountList)
	for i, v := range viewerCountList {
		if v < previous {
			log.Fatal("Got non increasing list to compute Gini")
		}
		previous = v
		iyi += (i + 1) * v
		yi += v
	}
	gini := float32(2.0*iyi)/float32(n*yi) - float32(n+1)/float32(n)
	return gini
}

type Category struct {
	BoxArtUrl string `json:"box_art_url"`
	Name      string `json:"name"`
	Id        string `json:"id"`
}
type CategoryResult struct {
	Data []Category `json:"data"`
}

const CategotyName = "Software and Game Development"

func authenticated_get(url string) *http.Response {
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	req.Header.Add("Client-Id", clientId)
	resp, _ := client.Do(req)
	return resp
}

func get_id() string {
	resp := authenticated_get("https://api.twitch.tv/helix/search/categories?query=%22software%20and%20Game&20Development%22")
	body, _ := io.ReadAll(resp.Body)
	var result CategoryResult
	json.Unmarshal([]byte(body), &result)
	for _, category := range result.Data {
		if category.Name == CategotyName {
			return category.Id
		}
	}
	log.Fatal("Could not find category id")
	return ""
}

type Stream struct {
	Id           string   `json:"id"`
	UserId       string   `json:"user_id"`
	UserLogin    string   `json:"user_login"`
	UserName     string   `json:"user_name"`
	GameId       string   `json:"game_id"`
	GameName     string   `json:"game_name"`
	Tags         []string `json:"tags"`
	Viewer_count int      `json:"viewer_count"`
	StartedAt    string   `json:"started_at"`
	Language     string   `json:"language"`
	ThumbnailUrl string   `json:"thumbnail_url"`
	TagIds       []string `json:"tag_ids"`
	IsMature     bool
}

type StreamResult struct {
	Data []Stream `json:"data"`
}

func get_streams(id string) []Stream {
	url := "https://api.twitch.tv/helix/streams?game_id=" + id + "&type=live"
	fmt.Println(url)
	resp := authenticated_get(url)
	body, _ := io.ReadAll(resp.Body)
	var result StreamResult
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		log.Fatal(err)
	}
	return result.Data
}
