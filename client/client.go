package client

import (
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)


type Rss struct {
    Channel struct {
        Title string `xml:"title"`
        Items []Item `xml:"item"`
    } `xml:"channel"`
}

type Item struct {
    PodcastTitle string
    Title string `xml:"title"`
    Link string `xml:"link"`
    Date string `xml:"pubDate"`
    Id string
    LastElement bool
    Image struct {
        Url string `xml:"href,attr"`
    } `xml:"image"`
    Enclosure struct {
        Url string `xml:"url,attr"`
    } `xml:"enclosure"`
}

type Response struct {
    Podcasts []struct {
        Title string `json:"title"`
        Url string `json:"url"`
        Image string `json:"image"`
    } `json:"feeds"`
}

func SearchPodcast(title string) Response{
    client := &http.Client{}
    apiKey := ""
    apiSecret := ""
    authDate := time.Now().Unix()
    authorization := fmt.Sprintf("%s%s%d", apiKey, apiSecret, authDate)
    fmt.Println("auth date", fmt.Sprintf("%d", authDate))

    h := sha1.New()
    h.Write([]byte(authorization))
    authorizationEncrypted := fmt.Sprintf("%x", h.Sum(nil))
    fmt.Println("auth", authorizationEncrypted)

    req, _ := http.NewRequest("GET", "https://api.podcastindex.org/api/1.0/search/bytitle", nil)
    req.Header.Add("X-Auth-Date", fmt.Sprintf("%d", authDate))
    req.Header.Add("X-Auth-Key", apiKey)
    req.Header.Add("Authorization", authorizationEncrypted)

    q := req.URL.Query()
    q.Add("q", title)
    req.URL.RawQuery = q.Encode()

    res, err := client.Do(req)
    if err != nil {
        fmt.Println("Error on podcastindex API")
        fmt.Println(err)
    }
    fmt.Println(res.StatusCode)
    defer res.Body.Close()

    var response Response
    body, err := io.ReadAll(res.Body)
    json.Unmarshal(body, &response)
    
    return response
}

func ParseRssFeed(feedUrl string) *Rss {
    resp, err := http.Get(feedUrl)
    if err != nil {
        fmt.Println("Error retrieving feed")
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error parsing body")
    }

    var parsedXml Rss
    if err = xml.Unmarshal(body, &parsedXml); err != nil {
        fmt.Println("Error parsing xml")
    }

    return &parsedXml
}
