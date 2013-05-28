/**
 + controller.go
 * ユーザからのリクエストに合わせて処理を振り分ける
 * @file
 */
package escape3ds

import (
	"net/http"
	"appengine"
	"log"
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		this.login(w, r)
	})
	http.HandleFunc("/oauth_callback", func(w http.ResponseWriter, r *http.Request) {
		this.oauthCallback(w, r)
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
	
	oauth := NewOAuth(c)
	tokens := oauth.requestToken()
	oauth.redirect(tokens["oauth_token"], w, r)
}

/**
 * OAuthで他のサイトでログインしてから戻ってきた時
 * @method
 * @memberof Controller
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func (this *Controller) oauthCallback(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("oauth_token")
	verifier := r.FormValue("oauth_verifier")
	
	c := appengine.NewContext(r)
	oauth := NewOAuth(c)
	result := oauth.convertToken(token, verifier)
	log.Printf("RESULT: %#v", result)
	
	view := new(View)
	view.login(c, w)
	fmt.Fprintf(w, "あなたのidは %s です<br>あなたのユーザ名は %s です", result["user_id"], result["screen_name"])
}