/**
 * OAuth 2.0 による通信
 * authorization code 方式のみ
 * @file
 */
package escape3ds

import (
	"appengine"
	"net/http"
	"strings"
	"fmt"
)

/**
 * OAuth 2.0
 * @class
 * @property {appengine.Context} context コンテキスト
 * @property {string} clientId クライアントID
 * @property {string} clientSecret クライアントパスワード
 */
type OAuth2 struct {
	context appengine.Context
	clientId string
	clientSecret string
}

/**
 * OAuth2.0 インスタンスを返す
 * @function
 * @param {appengine.Context} c コンテキスト
 * @param {string} clientId OAuthクライアントID
 * @param {string} clientSecret OAuthクライアントパスワード
 * @returns {*OAuth2} インスタンス
 */
func NewOAuth2(c appengine.Context, clientId string, clientSecret string) *OAuth2 {
	oauth := new(OAuth2)
	oauth.context = c
	oauth.clientId = clientId
	
	oauth.clientSecret = clientSecret
	return oauth
}

/**
 * 認証コードをリクエストする
 * 認証ページヘのリダイレクトを行いユーザに認証してもらう
 * 認証が完了したらリダイレクトURIへ認証コードが返ってくる
 * @method
 * @memberof OAuth2
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 * @param {string} targetUri リクエスト先URI
 * @param {string} redirectUri リダイレクトURI
 */
func (this *OAuth2) requestAuthorizationCode(w http.ResponseWriter, r *http.Request, targetUri string, redirectUri string) {
	targetUri = fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code", targetUri, this.clientId, redirectUri)
	http.Redirect(w, r, targetUri, 302)
}

/**
 * アクセストークンをリクエストする
 * 引き換えとして認証コードを渡すこと
 * @method
 * @memberof OAuth2
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 * @param {string} targetUri
 * @param {string} redirectUri
 * @param {string} clientSecret
 * @param {string} code
 * @returns {string} アクセストークン
 */
func (this *OAuth2) requestAccessToken(w http.ResponseWriter, r *http.Request, targetUri string, redirectUri string, code string) string {
	targetUri = fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&client_secret=%s&code=%s", targetUri, this.clientId, redirectUri, this.clientSecret, code)
	
	params := make(map[string]string, 4)
	params["client_id"] = this.clientId
	params["redirect_uri"] = redirectUri
	params["client_secret"] = this.clientSecret
	params["code"] = code
	response := request(this.context, "GET", targetUri, params, "")
	
	body := make([]byte, 1024)
	_, err := response.Body.Read(body)
	check(this.context, err)
	
	// response: oauth_token=******&expires=******
	responseParams := strings.Split(string(body), "&")
	tokenParam := strings.Split(responseParams[0], "=")
	return tokenParam[1]
}

/**
 * アクセストークンを使ってAPIを呼び出す
 * @method
 */
func (this *OAuth2) requestAPI(w http.ResponseWriter, targetUri string, accessToken string) {
	params := make(map[string]string, 1)
	params["access_token"] = accessToken
	response := request(this.context, "GET", targetUri, params, "")
	fmt.Fprintf(w, "response: %v<br><br>", response)
	
	result := make([]byte, 1024)
	response.Body.Read(result)
	fmt.Fprintf(w, "BODY: %s<br><br>", result)
}