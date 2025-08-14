package olivetv

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/iawia002/lux/request"

	"github.com/go-olive/olive/foundation/olivetv/model"
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
	return this.set(tv)
}

func (this *douyin) set(tv *TV) error {
	ttwid, err := this.ttwid()
	if err != nil {
		return err
	}
	cookie := "ttwid=" + ttwid

	api := `https://live.douyin.com/webcast/room/web/enter/`
	resp, err := req.R().
		SetHeaders(map[string]string{
			HeaderUserAgent:   CHROME,
			"referer":         "https://live.douyin.com/",
			"cookie":          cookie,
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
			"a_bogus":          "m70RgttJEd8fCdFGmOnpCWAlE",
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
	tv.streamerName = gjson.Get(text, "owner.nickname").String()

	return nil
}

func (this *douyin) ttwid() (string, error) {
	body := map[string]interface{}{
		"aid":           1768,
		"union":         true,
		"needFid":       false,
		"region":        "cn",
		"cbUrlProtocol": "https",
		"service":       "www.ixigua.com",
		"migrate_info":  map[string]string{"ticket": "", "source": "node"},
	}
	bytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	payload := strings.NewReader(string(bytes))
	resp, err := request.Request(http.MethodPost, "https://ttwid.bytedance.com/ttwid/union/register/", payload, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() // nolint
	cookie := resp.Header.Get("Set-Cookie")
	re := regexp.MustCompile(`ttwid=([^;]+)`)
	if match := re.FindStringSubmatch(cookie); match != nil {
		return match[1], nil
	}
	return "", errors.New("douyin ttwid request failed")
}
