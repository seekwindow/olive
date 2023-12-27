package olivetv

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/go-olive/olive/foundation/olivetv/model"
	"github.com/go-olive/olive/foundation/olivetv/util"
	"github.com/imroc/req/v3"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

var (
	ErrCookieNotSet = errors.New("cookie not configured")
)

func init() {
	registerSite("douyin", &douyin{})
}

type douyin struct {
	base
}

func (this *douyin) Name() string {
	return "抖音"
}

func (this *douyin) Snap(tv *TV) error {
	tv.Info = &Info{
		Timestamp: time.Now().Unix(),
	}
	return this.setURL2(tv)
}

func (this *douyin) setURL2(tv *TV) error {
	tv.cookie = this.getCookie(tv)

	api := `https://live.douyin.com/webcast/room/web/enter/`
	resp, err := req.R().
		SetHeaders(map[string]string{
			HeaderUserAgent:   CHROME,
			"referer":         "https://live.douyin.com/",
			"cookie":          tv.cookie,
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"Accept-Language": "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
			"Cache-Control":   "no-cache",
		}).
		SetQueryParams(map[string]string{
			"aid":              "6383",
			"device_platform":  "web",
			"browser_language": "zh-CN",
			"browser_platform": "Win32",
			"browser_name":     "Chrome",
			"browser_version":  "92.0.4515.159",
			"web_rid":          tv.RoomID,
		}).
		Get(api)
	if err != nil {
		return err
	}
	// log.Println(api)
	text := resp.String()

	if !strings.Contains(text, "data") {
		return errors.New("empty text = " + text)
	}

	text = gjson.Get(text, "data.data.0").String()
	// 抖音 status == 2 代表是开播的状态
	if gjson.Get(text, "status").String() != "2" {
		return nil
	}

	streamDataStr := gjson.Get(text, "stream_url.live_core_sdk_data.pull_data.stream_data").String()
	var streamData model.DouyinStreamData
	err = jsoniter.UnmarshalFromString(streamDataStr, &streamData)
	if err != nil {
		return err
	}
	flv := streamData.Data.Origin.Main.Flv
	hls := streamData.Data.Origin.Main.Hls
	_ = hls
	tv.streamURL = flv
	tv.roomOn = true

	tv.roomName = gjson.Get(text, "title").String()

	return nil
}

func (this *douyin) getCookie(tv *TV) string {
	return this.generateCookie(tv)
}

func (this *douyin) generateCookie(tv *TV) string {
	url := "https://live.douyin.com/'740849246012"
	cookie := "__ac_nonce=" + this.AcNonce()
	resp, err := req.C().R().SetHeaders(
		map[string]string{
			"accept":        "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
			HeaderUserAgent: CHROME,
			HeaderCookie:    cookie,
		}).
		Get(url)
	if err != nil {
		return ""
	}
	tv.streamerName, _ = util.Match(`live-room-nickname">([^<]+)<`, resp.String())
	var ttwid string
	for _, c := range resp.Cookies() {
		// log.Println(c.Name, c.Value)
		if c.Name == "ttwid" {
			ttwid = c.Value
		}
	}
	if ttwid != "" {
		cookie += "; ttwid=" + ttwid
	}

	return cookie
}

func (this *douyin) AcNonce() string {
	arr := make([]string, 21)
	cands := strings.Split("1234567890abcdef", "")
	for i := range arr {
		arr[i] = cands[rand.Intn(len(cands))]
	}
	return strings.Join(arr, "")
}
