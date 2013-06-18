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
	"encoding/json"
)

/**
 * @class
 */
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
		this.top(w, r)
	})
	
	// Twitter ログイン
	http.HandleFunc("/login_twitter", func(w http.ResponseWriter, r *http.Request) {
		this.loginTwitter(w, r)
	})
	
	// Twitter からのコールバック
	http.HandleFunc("/oauth_callback", func(w http.ResponseWriter, r *http.Request) {
		this.oauthCallback(w, r)
	})
	
	// Facebook ログイン
	http.HandleFunc("/login_facebook", func(w http.ResponseWriter, r *http.Request) {
		this.loginFacebook(w, r)
	})
	
	// Facebook からのコールバック
	http.HandleFunc("/callback_facebook", func(w http.ResponseWriter, r *http.Request) {
		if(r.FormValue("access_token") == "") {
			this.requestFacebookToken(w, r)
		} else {
			fmt.Fprintf(w, "ログイン完了")
		}
	})
	
	// ログイン成功したらエディタを表示
	http.HandleFunc("/editor", func(w http.ResponseWriter, r *http.Request) {
		this.editor(w, r)
	})
	
	// 仮登録ページの表示
	http.HandleFunc("/interim_registration", func(w http.ResponseWriter, r *http.Request) {
		this.interimRegistration(w, r)
	})
	
	// DEBUG: デバッグページの表示
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		this.debug(w, r)
	})

	// API: ユーザの追加
	http.HandleFunc("/add_user", func(w http.ResponseWriter, r *http.Request) {
		this.addUser(w, r)
	})
	
	// API: ログイン
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		this.login(w, r);
	})
}

/**
 * ログインページの表示
 * @method
 * @memberof Controller
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func (this *Controller) top(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	view := NewView(c, w)
	view.login()
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
	
	view := NewView(c, w)
	
	if result["oauth_token"] != "" {
		view.editor()
		fmt.Fprintf(w, "あなたのidは %s です<br>あなたのユーザ名は %s です", result["user_id"], result["screen_name"])
	} else {
		view.login()
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
	response := oauth.requestAPI(w, "https://graph.facebook.com/me", token)
	
	type UserInfo struct {
		Id string `json:"id"`
		Name string `json:"name"`
	}
	userInfo := new(UserInfo)
	err := json.Unmarshal(response, userInfo)
	check(c, err)
	
	fmt.Fprintf(w, "info: %#v", userInfo)
}

/**
 * エディタの表示
 * @method
 * @memberof Controller
 * @param {http.ResponseWRiter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func (this *Controller) editor(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	view := NewView(c, w)
	view.editor()
}

/**
 * ユーザの追加
 * @method
 * @memberof Controller
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func (this *Controller) addUser(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	params := make(map[string]string, 5)
	params["user_type"] = r.FormValue("user_type")
	params["user_name"] = r.FormValue("user_name")
	params["user_pass"] = r.FormValue("user_pass")
	params["user_mail"] = r.FormValue("user_mail")
	params["user_oauth_id"] = r.FormValue("user_oauth_id")
	
	model := NewModel(c)
	model.addUser(params)
}

/**
 * デバッグツールの表示
 * @method
 * @memberof Controller
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func (this *Controller) debug(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	view := NewView(c, w)
	view.debug()
}

/**
 * ログイン
 * @method
 * @memberof Controller
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 * @returns {Ajax JSON} result 成功したらtrue
 * @returns {Ajax JSON} to 成功した時のリダイレクト先URL
 * @returns {Ajax JSON} message 失敗した時のエラーメッセージ
 */
func (this *Controller) login(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	mail := r.FormValue("mail")
	pass := r.FormValue("pass")
	
	model := NewModel(c)
	key, _ := model.loginCheck(mail, pass)
	if key != "" {
		fmt.Fprintf(w, `{"result":true, "to":"/editor?key=%s"}`, key)
	} else {
		fmt.Fprintf(w, `{"result":false, "message":"メールアドレスまたはパスワードが間違っています"}`)
	}
}

/**
 * 仮登録
 * @method
 * @memberof Controller
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func (this *Controller) interimRegistration(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	
	name := r.FormValue("name")
	mail := r.FormValue("mail")
	pass := r.FormValue("password")
	
	model := NewModel(c)
	key := model.interimRegistration(name, mail, pass)
	
	sendMail(c, "infomation@escape-3ds.appspotmail.com", mail, "仮登録完了のお知らせ", fmt.Sprintf(config["interimMailBody"], name, key))
	
	view := NewView(c, w)
	view.interimRegistration()
}