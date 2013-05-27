package escape3ds

import (
	"math/rand"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"strconv"
	"time"
	"fmt"
	"net/url"
	"appengine"
	"net/http"
	"crypto/hmac"
	"crypto/sha1"
	"log"
)

type OAuth struct {
	params map[string]string
	url string
	context appengine.Context
}

func NewOAuth(c appengine.Context) *OAuth {
	params := make(map[string]string, 7)
	params["oauth_callback"] = "http://localhost:8080"
	params["oauth_consumer_key"] = config["consumer_key"]
	params["oauth_signature_method"] = "HMAC-SHA1"
	params["oauth_version"] = "1.0"
	
	oauth := new(OAuth)
	oauth.params = params
	oauth.url = "https://api.twitter.com/oauth/request_token"
	oauth.context = c
	return oauth
}

func (this *OAuth) requestToken() {
	// リクエストごとに異なるパラーメタを作成
	this.params["oauth_nonce"] = this.createNonce()
	this.params["oauth_timestamp"] = strconv.Itoa(int(time.Now().Unix()))
	this.params["oauth_signature"] = this.createSignature()
	
	// ヘッダ作成
	params := make([]string, 0)
	for key, val := range this.params {
		key = url.QueryEscape(key)
		val = url.QueryEscape(val)
		set := fmt.Sprintf(`%s="%s"`, key, val)
		params = append(params, set)
	}
	header := strings.Join(params, ", ")
	header = fmt.Sprintf("OAuth %s", header)
	log.Printf("%s", header)
	
	// リクエスト送信
	request, err := http.NewRequest("POST", this.url, nil)
	check(this.context, err)
	request.Header.Add("Authorization", header)
	
	client := new(http.Client)
	response, err := client.Do(request)
	check(this.context, err)
	
	result := make([]byte, 1024)
	response.Body.Read(result)
	log.Printf("response:%s", result)
}

func (this *OAuth) createNonce() string {
	r := rand.Int63()
	b := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(b, r)
	e := base64.StdEncoding.EncodeToString(b)
	e = strings.Replace(e, "+", "", -1)
	e = strings.Replace(e, "/", "", -1)
	e = strings.Replace(e, "=", "", -1)
	return e
}

func (this *OAuth) createSignature() string {
	sort := []string {
		"oauth_callback",
		"oauth_consumer_key",
		"oauth_nonce",
		"oauth_signature_method",
		"oauth_timestamp",
		"oauth_version",
	}
	params := make([]string, len(sort))
	for i := 0; i < len(sort); i++ {
		key := sort[i]
		val := this.params[key]
		params[i] = fmt.Sprintf("%s=%s", url.QueryEscape(key), url.QueryEscape(val))
	}
	paramString := strings.Join(params, "&")
	baseString := fmt.Sprintf("POST&%s&%s", url.QueryEscape(this.url), url.QueryEscape(paramString))
	signatureKey := fmt.Sprintf("%s&", url.QueryEscape(config["consumer_secret"]))
	
	hash := hmac.New(sha1.New, []byte(signatureKey))
	hash.Write([]byte(baseString))
	signature := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature)
}