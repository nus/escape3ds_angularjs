/**
 + controller.go
 * ユーザからのリクエストに合わせて処理を振り分ける
 * @file
 */
package escape3ds

import (
	"net/http"
	"net/url"
	"appengine"
	"fmt"
)

type Controller struct {

}

/**
 * URLから処理を振り分ける
 * @method
 * @memberof Controller
 */
func (this *Controller) handle() {
	// トップ
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		this.login(w, r)
	})
	
	// Twitter ログイン
	http.HandleFunc("/login_twitter", func(w http.ResponseWriter, r *http.Request) {
		this.loginTwitter(w, r)
	})
	http.HandleFunc("/oauth_callback", func(w http.ResponseWriter, r *http.Request) {
		this.oauthCallback(w, r)
	})
	
	// Facebook ログイン
	http.HandleFunc("/login_facebook", func(w http.ResponseWriter, r *http.Request) {
		this.loginFacebook(w, r)
	})
	http.HandleFunc("/callback_facebook", func(w http.ResponseWriter, r *http.Request) {
		if(r.FormValue("access_token") == "") {
			this.requestFacebookToken(w, r)
		} else {
			fmt.Fprintf(w, "ログイン完了")
		}
	})
}

/**
 * ログインページの表示
 * @method
 * @memberof Controller
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func (this *Controller) login(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	view := new(View)
	view.login(c, w)
}

/**
 * Twitter でログイン
 * @method
 * @memberof Controller
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func (this *Controller) loginTwitter(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	oauth := NewOAuth1(c)
	result := oauth.requestToken("https://api.twitter.com/oauth/request_token")
	oauth.authenticate(w, r, "https://api.twitter.com/oauth/authenticate", result["oauth_token"])
}

/**
 * OAuth で他のサイトでログインしてから戻ってきた時
 * @method
 * @memberof Controller
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func (this *Controller) oauthCallback(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("oauth_token")
	verifier := r.FormValue("oauth_verifier")
	
	c := appengine.NewContext(r)
	oauth := NewOAuth1(c)
	result := oauth.exchangeToken(token, verifier, "https://api.twitter.com/oauth/access_token")
	
	view := new(View)
	
	if result["oauth_token"] != "" {
		view.editor(c, w)
		fmt.Fprintf(w, "あなたのidは %s です<br>あなたのユーザ名は %s です", result["user_id"], result["screen_name"])
	} else {
		view.login(c, w)
		fmt.Fprintf(w, "ログインに失敗しました")
	}
}

/**
 * Facebook でログイン
 * @method
 * @memberof Controller
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func (this *Controller) loginFacebook(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	oauth := NewOAuth2(c, config["facebook_client_id"], config["facebook_client_secret"])
	oauth.requestAuthorizationCode(w, r, "https://www.facebook.com/dialog/oauth", url.QueryEscape("http://escape-3ds.appspot.com/callback_facebook"))
}

/**
 * Facebook へアクセストークンを要求する
 * この関数は Facebook から認証コードをリダイレクトで渡された時に呼ばれる
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func (this *Controller) requestFacebookToken(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	code := r.FormValue("code")
	oauth := NewOAuth2(c, config["facebook_client_id"], config["facebook_client_secret"])
	token := oauth.requestAccessToken(w, r, "https://graph.facebook.com/oauth/access_token", url.QueryEscape("http://escape-3ds.appspot.com/callback_facebook"), code)
	oauth.requestAPI(w, "https://graph.facebook.com/me", token)
}