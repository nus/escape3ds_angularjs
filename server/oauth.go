/**
 * OAuthで認証する
 * @file
 */
package escape3ds

import (
	"time"
	"math/rand"
	"encoding/base64"
	"encoding/binary"
	"strconv"
	"strings"
	"net/url"
	"log"
)

/**
 * OAuth 1.0 で認証するためのクラス
 * @class
 * @member {map[string]string} params ヘッダに含める7つのパラメータ
 *	- oauth_consumer_key
 *	- oauth_nonce
 *	- oauth_signature
 *	- oauth_signature_method
 *	- oauth_timestamp
 *	- oauth_token
 *	- oauth_version
 * @member {string} url リクエスト送信先
 */
type OAuth struct {
	params map[string]string
	url string
}

/**
 * OAuth のインスタンス化
 * @function
 * @param {map[string]string} 設定オブジェクト
 * - consumer_key
 * - token
 * @returns {*OAuth} OAuth インスタンス
 */
func NewOAuth(config map[string]string) *OAuth {
	oauth := new(OAuth)
	oauth.params = make(map[string]string, 7)
	oauth.params["oauth_consumer_key"] = config["consumer_key"]
	oauth.params["oauth_nonce"] = ""
	oauth.params["oauth_signature"] = ""
	oauth.params["oauth_signature_method"] = "HMAC-SHA1"
	oauth.params["oauth_timestamp"] = ""
	oauth.params["oauth_token"] = config["token"]
	oauth.params["oauth_version"] = "1.0"
	oauth.url = config["url"]
	return oauth
}

/**
 * oauth_nonce を作成する
 * 重複しないランダムな文字列
 * @method
 * @memberof OAuth
 * @returns {string} ランダムな文字列
 */
func (this *OAuth) createNonce() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, (int64)(r.Int63()))
	return base64.URLEncoding.EncodeToString(buf)
}

/**
 * 認証用のリクエストヘッダを作成する
 * @method
 * @memberof OAuth
 * @returns {string} ヘッダ
 */
func (this *OAuth) createHeader() {
	this.params["oauth_nonce"] = this.createNonce()
	this.params["oauth_timestamp"] = strconv.Itoa((int)(time.Now().Unix()))
	this.params["oauth_signature"] = this.createSignature()
	
	header := "OAuth "
	for key, data := range this.params {
		log.Printf("params[%s] = %s", key, data)
	}
	log.Printf("Header: %s", header)
}

/**
 * シグネチャの作成
 * @method
 * @memberof OAuth
 * @returns {string} シグネチャ
 */
func (this *OAuth) createSignature() string {
	// アルファベット順にパラメータを並べる必要がある
	keySort := []string{
		"oauth_consumer_key",
		"oauth_nonce",
		"oauth_signature_method",
		"oauth_timestamp",
		"oauth_token",
		"oauth_version",
	}
	
	// パラメータをまとめる
	pairs := make([]string, len(keySort))
	params := ""
	for i := 0; i < len(keySort); i++ {
		key := keySort[i]
		value := this.params[key]
		key = url.QueryEscape(key)
		value = url.QueryEscape(value)
		pairs[i] = strings.Join([]string{key, value}, "=")
	}
	params = strings.Join(pairs, "&")
	
	// シグネチャを作成
	signature := ""
	method := "POST"
	targetUrl := url.QueryEscape(this.url)
	params = url.QueryEscape(params)
	signature = strings.Join([]string{method, targetUrl, params}, "&")
	
	return signature
}

/**
 * リクエストを送信する
 * @method
 * @memberof OAuth
 */
func (this *OAuth) request() {
	
}