package common

import (
	"encoding/json"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

var (
	GOOGLE_TRANSLATE_URL = "http://translate.google.com/m?tl=zh-CN&sl=en&q="
	TRANSLATE_PROXY      = "http://proxy01.uniontech.com:3128"
	TRANSFILE            = "locale/zh-ch.json"
	TRANSMAP             = make(map[string]string)
)

// generate new translation file
func Generate(sourceContent string) {
	
	err := os.MkdirAll("locale", 0755)
	if err != nil {
		return
	}
	
	// create zh-cn.json
	file, err := os.OpenFile(TRANSFILE, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	// json file to map
	fileInfo, err := file.Stat()
	if err != nil {
		return
	}
	var mapOrigin = make(map[string]string)
	if fileInfo.Size() > 0 {
		jsondata, err := ioutil.ReadFile(TRANSFILE)
		if err != nil {
			return
		}
		err = json.Unmarshal(jsondata, &mapOrigin)
		if err != nil {
			return
		}
	}
	// translate en to zh-cn
	pattern := "^[\\s\n]+$"
	matched, _ := regexp.MatchString(pattern, sourceContent)
	if sourceContent != "" && !matched && mapOrigin[sourceContent] == "" {
		mapOrigin[sourceContent] = translateDescription(sourceContent)
	} else {
		return
	}
	err = ioutil.WriteFile(TRANSFILE, nil, 0644)
	if err != nil {
		return
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(mapOrigin)
	if err != nil {
		return
	}
}

// translation by google api
func translateDescription(sourceContent string) string {
	// 内容转换为URLCode
	Content := url.QueryEscape(sourceContent)
	requestURL := GOOGLE_TRANSLATE_URL + Content
	// 保证访问间隔0.2s
	time.Sleep(200 * time.Microsecond)
	// 创建一个HTTP客户端
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				proxyURL, err := url.Parse(TRANSLATE_PROXY)
				if err != nil {
					return nil, err
				}
				return proxyURL, nil
			},
		},
	}
	// 创建一个HTTP请求
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "创建请求失败"
	}
	// 发送请求
	response, err := client.Do(req)
	if err != nil {
		return "发送请求失败"
	}
	defer response.Body.Close()
	// 解析Body结构中的数据
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return sourceContent
	}
	// 正则过滤除翻译内容的其他信息
	expr := regexp.MustCompile(`(?s)class="(?:t0|result-container)">(.*?)<`)
	result := expr.FindAllStringSubmatch(string(data), -1)
	if len(result) == 0 {
		return sourceContent
	}
	targetContent := html.UnescapeString(result[0][1])
	// 字符串json样本校验
	exampleJson := `{"example": "` + targetContent + `"}`
	if !json.Valid([]byte(exampleJson)) {
		return sourceContent
	}
	return targetContent
}

// translate en to zh-cn
func TranslateDescription(sourceContent string) string {
	if len(TRANSMAP) == 0 {
		TRANSMAP = Load(TRANSFILE)
	}
	if TRANSMAP == nil {
		return sourceContent
	}
	if TRANSMAP[sourceContent] == "" {
		return sourceContent
	}
	return TRANSMAP[sourceContent]
}

// load json file to map
func Load(filepath string) map[string]string {
	if !Exist(filepath) {
		return nil
	}
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil
	}
	var mapLoad = make(map[string]string)
	err = json.Unmarshal(file, &mapLoad)
	if err != nil {
		return nil
	}
	return mapLoad
}

// deteemine whether the file exists
func Exist(filepath string) bool {
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsExist(err) {
			return true
		} else {
			return false
		}
	}
	return true
}
