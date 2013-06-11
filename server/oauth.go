/**
 * Twitterとの通信
 * OAuth 1.0 Revision A を使う
 * @file
 */
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
	"appengine/urlfetch"
	"net/http"
	"crypto/hmac"
	"crypto/sha1"
	"sort"
	"io"
	"log"
)

/**
 * OAuthの通信を行うクラス
 * @class
 * @param {map[string]string} params oauthパラメータの配列
 * @param {appengine.Context} context コンテキスト
 */
type OAuth struct {
	params map[string]string
	context appengine.Context
}

/**
 * OAuthクラスのインスタンス化
 * @function
 * @params {appengine.Context} c コンテキスト
 * @returns {*OAuth} OAuthインスタンス
 */
func NewOAuth(c appengine.Context) *OAuth {
	params := make(map[string]string, 7)
	params["oauth_callback"] = "http://localhost:8080/oauth_callback"
	params["oauth_consumer_key"] = config["consumer_key"]
	params["oauth_signature_method"] = "HMAC-SHA1"
	params["oauth_version"] = "1.0"
	
	oauth := new(OAuth)
	oauth.params = params
	oauth.context = c
	return oauth
}

/**
 * リクエストボディ用のリーダー
 * @class
 * @member {[]byte} body 本文
 * @member {int} pointer 何バイト目まで読み込んだか表すポインタ
 */
type Reader struct {
	io.Reader
	body []byte
	pointer int
}

/**
 * リーダーの作成
 * @param {string} body 本文
 * @returns {*Reader} 作成したインスタンス
 */
func NewReader(body string) *Reader {
	reader := new(Reader)
	reader.body = []byte(body)
	reader.pointer = 0
	return reader
}

/**
 * 本文を読み出す
 * ２回目以降は前回の続きから読み出せる
 * @method
 * @memberof *Reader
 * @param {[]byte} p 読みだしたデータの保存先
 * @returns {int} 読みだしたバイト数
 * @returns {error} エラー
 */
func (this *Reader) Read(p []byte) (int, error) {
	var l int
	var err error
	if this.pointer + len(p) < len(this.body) {
		l = len(p)
		err = nil
	} else {
		l = len(this.body) - this.pointer
		err = io.EOF
	}
	
	for i := 0; i < l; i++ {
		p[i] = this.body[i + this.pointer]
	}
	
	this.pointer = l + this.pointer
	
	return l, err
}

/**
 * Twitter へリクエストトークンを要求する
 * @method
 * @memberof OAuth
 * @param {string} targetUrl リクエスト要求先のURL
 * @returns {map[string]string} リクエスト結果
 */
func (this *OAuth) requestToken(targetUrl string) map[string]string {
	response := this.request(targetUrl, "")
	datas := strings.Split(response, "&")
	result := make(map[string]string, len(datas))
	for i := 0; i < len(datas); i++ {
		data := strings.Split(datas[i], "=")
		result[data[0]] = data[1]
	}
	
	return result
}

/**
 * リクエストを送信してレスポンスを受信する
 * メソッドは POST 固定
 * @method
 * @memberof OAuth
 * @param {string} targetUrl 送信先
 * @param {string} body リクエストボディ
 * @returns {string} レスポンス
 */
func (this *OAuth) request(targetUrl string, body string) string {

	// リクエストごとに変わるパラメータを設定
	this.params["oauth_nonce"] = this.createNonce()
	this.params["oauth_timestamp"] = strconv.Itoa(int(time.Now().Unix()))
	this.params["oauth_signature"] = this.createSignature(targetUrl)
	
	// Authorization Header を作成
	header := this.createHeader()
	
	// リクエストの作成
	var request *http.Request
	var err error
	if body == "" {
		request, err = http.NewRequest("POST", targetUrl, nil)
	} else {
		request, err = http.NewRequest("POST", targetUrl, NewReader(body))
	}
	check(this.context, err)
	request.Header.Add("Authorization", header)
	
	// リクエストの送信とレスポンスの受信
	client := urlfetch.Client(this.context)
	response, err := client.Do(request)
	check(this.context, err)
	
	// レスポンスボディの読み取り
	result := make([]byte, 256)
	response.Body.Read(result)
	
	return string(result)
}

/**
 * Aouthorization ヘッダを作成する
 * @method
 * @memberof OAuth
 * @returns {string} ヘッダ
 */
func (this *OAuth) createHeader() string {
	params := make([]string, 0)
	for key, val := range this.params {
		key = url.QueryEscape(key)
		val = url.QueryEscape(val)
		set := fmt.Sprintf(`%s="%s"`, key, val)
		params = append(params, set)
	}
	header := strings.Join(params, ", ")
	header = fmt.Sprintf("OAuth %s", header)
	return header
}

/**
 * oauth_nonce を作成する
 * @method
 * @memberof OAuth
 * @returns {string} oauth_nonce
 */
func (this *OAuth) createNonce() string {
	r := rand.Int31()
	b := make([]byte, binary.MaxVarintLen32)
	binary.PutVarint(b, int64(r))
	e := base64.StdEncoding.EncodeToString(b)
	e = strings.Replace(e, "+", "", -1)
	e = strings.Replace(e, "/", "", -1)
	e = strings.Replace(e, "=", "", -1)
	return e
}

/**
 * oauth_signature を作成する
 * @method
 * @memberof OAuth
 * @param {string} targetUrl リクエスト送信先のURL
 * @returns {string} oauth_signature
 */
func (this *OAuth) createSignature(targetUrl string) string {
	keys := make([]string, 0)
	for key := range this.params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	
	params := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		val := this.params[key]
		params[i] = fmt.Sprintf("%s=%s", url.QueryEscape(key), url.QueryEscape(val))
	}
	paramString := strings.Join(params, "&")
	baseString := fmt.Sprintf("POST&%s&%s", url.QueryEscape(targetUrl), url.QueryEscape(paramString))
	
	signatureKey := fmt.Sprintf("%s&", url.QueryEscape(config["consumer_secret"]))
	hash := hmac.New(sha1.New, []byte(signatureKey))
	hash.Write([]byte(baseString))
	signature := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature)
}

/**
 * 認証ページヘリダイレクトする
 * @memberof OAuth
 * @method
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 * @param {string} targetUrl リダイレクト先
 * @param {string} token 未認証リクエストトークン
 */
func (this *OAuth) authenticate(w http.ResponseWriter, r *http.Request, targetUrl string, token string) {
	to := fmt.Sprintf("?oauth_token=%s", token)
	to = strings.Join([]string{targetUrl, to}, "")
	http.Redirect(w, r, to, 302)
}

/**
 * リクエストトークンをアクセストークンに変換する
 * @memberof OAuth
 * @method
 * @param {string} token リクエストトークン
 * @param {string} verifier 認証データ
 * @returns {map[string]string}
 */
func (this *OAuth) exchangeToken(token string, verifier string) map[string]string {
	this.params["oauth_token"] = token
	body := fmt.Sprintf("oauth_verifier=%s", verifier)
	response := this.request("https://api.twitter.com/oauth/access_token", body)
	log.Printf("response:%s", response)
	
	datas := strings.Split(response, "&")
	result := make(map[string]string, len(datas))
	for i := 0; i < len(datas); i++ {
		data := strings.Split(datas[i], "=")
		result[data[0]] = data[1]
	}
	return result
}