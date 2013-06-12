/**
 * OAuth 2.0 による通信
 * authorization code 方式のみ
 * @file
 */
package escape3ds

import (
	"appengine"
)

/**
 * OAuth 2.0
 * @class
 * @property {appengine.Context} context コンテキスト
 */
type OAuth2 struct {
	context appengine.Context
	clientId string
}

/**
 * OAuth2.0 インスタンスを返す
 * @function
 * @param {appengine.Context} c コンテキスト
 * @param {string} clientId OAuthクライアントID
 * @returns {*OAuth2} インスタンス
 */
func NewOAuth2(c appengine.Context, clientId string) *OAuth2 {
	oauth := new(OAuth2)
	oauth.context = c
	oauth.clientId = clientId
	return oauth
}

/**
 * 認証コードをリクエストする
 * @method
 * @memberof OAuth2
 * @param {string} targetUri リクエスト先URI
 * @param {string} redirectUri リダイレクトURI
 */
func (this *OAuth2) requestAuthorizationCode(targetUri string, redirectUri string) {
	
}