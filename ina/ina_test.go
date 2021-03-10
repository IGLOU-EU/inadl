// Copyright 2021 Iglou.eu
// license that can be found in the LICENSE file

package ina

import (
	"fmt"
	"testing"
)

func tError(r bool, s string, t *testing.T) {
	if r {
		t.Errorf("%s", s)
		t.Fail()
	}
}

func TestUrlExtractID(t *testing.T) {
	r := urlExtractID("")
	tError(r != "", fmt.Sprintf("\nRequeste `%s` give `%s`", "", r), t)

	r = urlExtractID("https://www.ina.fr/TUT TUT GROSSE P*TE/PUB232175070/playstation-lancement-video.html")
	tError(r != "", fmt.Sprintf("\nRequeste `%s` give `%s`", "https://www.ina.fr/TUT TUT GROSSE P*TE/PUB232175070/playstation-lancement-video.html", r), t)

	r = urlExtractID("https://www.ina.fr/video/PUB2393641146")
	tError(r != "PUB2393641146", fmt.Sprintf("\nRequeste `%s` give `%s`", "https://www.ina.fr/video/PUB2393641146", r), t)

	r = urlExtractID("https://www.ina.fr/video/PUB232175070/playstation-lancement-video.html")
	tError(r != "PUB232175070", fmt.Sprintf("\nRequeste `%s` give `%s`", "https://www.ina.fr/video/PUB232175070/playstation-lancement-video.html", r), t)
}

func TestGetMrss(t *testing.T) {
	r, e := getMrss("")
	tError(e == nil, "Empty URL expect to fail", t)

	r, e = getMrss("https://duckduckgo.com/")
	tError(e == nil, "This bad request expect to fail, for bad header type", t)

	r, e = getMrss("https://player.ina.fr/notices/.mrss")
	tError(e == nil, "This bad request expect to fail, for bad header status", t)

	r, e = getMrss("https://player.ina.fr/notices/PUB2393641146.mrss")
	tError(e != nil, fmt.Sprintf("%s", e), t)
	tError(r == nil, "This ressource expect to return an MRSS code", t)
}

func TestMediaNew(t *testing.T) {
	r, e := MediaNew("")
	tError(e == nil, "Empty URL expect to fail", t)

	r, e = MediaNew("https://duckduckgo.com/")
	tError(e == nil, "This media expect to fail, bad url", t)

	r, e = MediaNew("https://player.ina.fr/video/")
	tError(e == nil, "This media expect to fail, request fail ...", t)

	r, e = MediaNew("https://www.ina.fr/video/PUB2393641146")
	tError(e != nil, fmt.Sprintf("%s", e), t)
	tError(r.Channel.ID == "", "This ressource expect to be a none empty MRSS struct", t)
}
