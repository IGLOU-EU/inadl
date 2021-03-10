package ina

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type MRSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		ID          string `xml:"id"`
		Title       string `xml:"title"`
		Description string `xml:"description"`
		Link        string `xml:"link"`
		PubDate     string `xml:"pubDate"`
		Category    string `xml:"category"`
		Item        struct {
			Content struct {
				Hq struct {
					URL string `xml:"url,attr"`
				} `xml:"hq"`
				Mq struct {
					URL string `xml:"url,attr"`
				} `xml:"mq"`
				Bq struct {
					URL string `xml:"url,attr"`
				} `xml:"bq"`
				Thumbnail []struct {
					URL    string `xml:"url,attr"`
					Height string `xml:"height,attr"`
					Width  string `xml:"width,attr"`
				} `xml:"thumbnail"`
			} `xml:"content"`
		} `xml:"item"`
	} `xml:"channel"`
}

type Config struct {
	UrlMrss string
}

var config Config

func init() {
	config.UrlMrss = "https://player.ina.fr/notices/%s.mrss"
}

func MediaNew(u string) (MRSS, error) {
	if u == "" {
		return MRSS{}, fmt.Errorf("URL can't be empty")
	}

	// On check que le lien est bien ina
	// On recupere l'id
	id := urlExtractID(u)
	if id == "" {
		return MRSS{}, fmt.Errorf("URL not recognized `%s`", u)
	}

	// On recupere les info
	gm, err := getMrss(fmt.Sprintf(config.UrlMrss, id))
	if err != nil {
		return MRSS{}, err
	}

	// On check les infos
	mrss := MRSS{}
	if err := xml.Unmarshal(gm, &mrss); err != nil {
		return MRSS{}, err
	}

	// on retourne
	return mrss, nil
}

func urlExtractID(u string) string {
	idPos := 0
	paths := strings.Split(u, "/")

	for i := range paths {
		if paths[i] == "video" {
			idPos = i + 1
			break
		}
	}

	if idPos == 0 || len(paths) < idPos {
		return ""
	}

	return paths[idPos]
}

func getMrss(u string) ([]byte, error) {
	if u == "" {
		return nil, fmt.Errorf("Url can't be an empty string")
	}

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Get `%s` error returned `%s`", u, resp.Status)
	}

	if resp.Header.Get("Content-Type") != "application/xml" {
		return nil, fmt.Errorf("Get `%s` bad content type returned `%s`", u, resp.Header.Get("Content-Type"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
