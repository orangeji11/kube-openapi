package common

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type DictRequest struct {
	TransType string `json:"trans_type"`
	Source    string `json:"source"`
	UserID    string `json:"user_id"`
}
type DictResponse struct {
	TranslateResult [][]struct {
		Tgt string `json:"tgt"`
		Src string `json:"src"`
	} `json:"translateResult"`
	ErrorCode   int    `json:"errorCode"`
	Type        string `json:"type"`
	SmartResult struct {
		Entries []string `json:"entries"`
		Type    int      `json:"type"`
	} `json:"smartResult"`
}

func encrypt(str string) string { //md5 加密函数， 传入字符串，返回加密后的字符串
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func TranslateDescription(word string) string {
	t := time.Now().UnixMilli()     //获取时间戳
	lts := strconv.FormatInt(t, 10) //时间戳就是lts
	rand.Seed(time.Now().UnixNano())
	salt := lts + strconv.Itoa(rand.Intn(9))                                //lts + 随机数就是salt
	sign := encrypt("fanyideskweb" + word + salt + "Ygy_4c=r#e#4EX^NUGUc5") //对这些字符串 进行md5加密，返回就是sign
	client := &http.Client{}                                                //创建了一个http client，可以携带很多参数

	//var data = bytes.NewReader(buf)
	var data = strings.NewReader("i=" + word + "&from=AUTO&to=AUTO&smartresult=dict&client=fanyideskweb&salt=" + salt + "&sign=" + sign + "&lts=" + lts + "&bv=d60b9bede0ddd264422f25a5e061c49a&doctype=json&version=2.1&keyfrom=fanyi.web&action=FY_BY_REALTlME")
	req, err := http.NewRequest("POST", "https://fanyi.youdao.com/translate_o?smartresult=dict&smartresult=rule", data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="97", "Chromium";v="97"`)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("Origin", "https://fanyi.youdao.com")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", "https://fanyi.youdao.com/")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cookie", "OUTFOX_SEARCH_USER_ID_NCOO=571199853.2191676; _ntes_nnid=345417059c531595fb4fe238fef920d8,1629898865976; OUTFOX_SEARCH_USER_ID=1226273813@1.85.38.28; JSESSIONID=aaaAUKGsqk_QBgByiQJcy; fanyi-ad-id=305838; fanyi-ad-closed=1; ___rl__test__cookies=1652014252491")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("bad StatusCode:", resp.StatusCode, "body", string(bodyText))
	}

	var dictResponse DictResponse
	err = json.Unmarshal(bodyText, &dictResponse)
	if err != nil {
		log.Printf("error decoding sakura response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Printf("sakura response: %q", bodyText)
		log.Printf("returning origin word")
		return word
	}
	//fmt.Printf("%#v\n", dictResponse)
	//fmt.Println(word, "UK:", dictResponse.SmartResult.Entries, "US:", dictResponse.SmartResult.Type)
	res := dictResponse.TranslateResult
	// 排除长度为0
	if len(res) == 0 {
		return word
	}

	item := res[0][0].Tgt

	return item
}
