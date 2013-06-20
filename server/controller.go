/**
 + controller.go
 * http.HandleFunc() を書きやすくするためクラス化はしない
 * URL パターンに該当する処理を書く
 * クラス化されたModelとViewを使って処理を進める
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
 * ログインページの表示
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func top(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	view := NewView(c, w)
	view.login()
}

/**
 * Twitter でログイン
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func loginTwitter(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	oauth := NewOAuth1(c, "http://escape-3ds.appspot.com/callback_twitter")
	result := oauth.requestToken("https://api.twitter.com/oauth/request_token")
	oauth.authenticate(w, r, "https://api.twitter.com/oauth/authenticate", result["oauth_token"])
}

/**
 * Twitter からのコールバック * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func callbackTwitter(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	token := r.FormValue("oauth_token")
	verifier := r.FormValue("oauth_verifier")
	
	oauth := NewOAuth1(c, "http://escape-3ds.appspotcom/callback_twitter")
	result := oauth.exchangeToken(token, verifier, "https://api.twitter.com/oauth/access_token")
	
	view := NewView(c, w)
	model := NewModel(c)
	
	if result["oauth_token"] != "" {
		// ログイン成功
		if model.existOAuthUser("Twitter", result["user_id"]) {
			// 既存ユーザ
			params := make(map[string]string, 2)
			params["Type"] = "Twitter"
			params["OAuthId"] = result["user_id"]
			key := model.getUserKey(params)
			if key != "" {
				view.editor(key)
			} else {
				c.Errorf("既存のTwitterのアカウントを検索出来ませんでした")
			}
		} else {
			// 新規ユーザ
			params := make(map[string]string, 4)
			params["user_type"] = "Twitter"
			params["user_name"] = result["screen_name"]
			params["user_oauth_id"] = result["user_id"]
			params["user_pass"] = ""
			user := model.NewUser(params)
			key := model.addUser(user)
			view.editor(key)
		}
	} else {
		// ログイン失敗
		view.login()
		fmt.Fprintf(w, "ログインに失敗しました")
	}
}

/**
 * Facebook でログイン
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func loginFacebook(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	oauth := NewOAuth2(c, config["facebook_client_id"], config["facebook_client_secret"])
	oauth.requestAuthorizationCode(w, r, "https://www.facebook.com/dialog/oauth", url.QueryEscape("http://escape-3ds.appspot.com/callback_facebook"))
}

/**
 * Facebook へアクセストークンを要求する
 * この関数は Facebook から認証コードをリダイレクトで渡された時に呼ばれる
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 * @returns {map[string]string} ユーザ情報
 */
func requestFacebookToken(w http.ResponseWriter, r *http.Request) map[string]string {
	c := appengine.NewContext(r)
	code := r.FormValue("code")
	oauth := NewOAuth2(c, config["facebook_client_id"], config["facebook_client_secret"])
	token := oauth.requestAccessToken(w, r, "https://graph.facebook.com/oauth/access_token", url.QueryEscape("http://escape-3ds.appspot.com/callback_facebook"), code)
	response := oauth.requestAPI(w, "https://graph.facebook.com/me", token)
	
	// JSON を解析
	type UserInfo struct {
		Id string `json:"id"`
		Name string `json:"name"`
	}
	userInfo := new(UserInfo)
	err := json.Unmarshal(response, userInfo)
	check(c, err)
	
	result := make(map[string]string, 2)
	result["oauth_id"] = userInfo.Id
	result["name"] = userInfo.Name
	return result
}

/**
 * エディタの表示
 * @param {http.ResponseWRiter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func editor(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	key := r.FormValue("key")
	view := NewView(c, w)
	view.editor(key)
}

/**
 * ユーザの追加
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func addUser(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	params := make(map[string]string, 5)
	params["user_type"] = r.FormValue("user_type")
	params["user_name"] = r.FormValue("user_name")
	params["user_pass"] = r.FormValue("user_pass")
	params["user_mail"] = r.FormValue("user_mail")
	params["user_oauth_id"] = r.FormValue("user_oauth_id")
	
	model := NewModel(c)
	user := model.NewUser(params)
	model.addUser(user)
}

/**
 * デバッグツールの表示
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func debug(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	view := NewView(c, w)
	view.debug()
}

/**
 * ログイン * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 * @returns {Ajax JSON} result 成功したらtrue
 * @returns {Ajax JSON} to 成功した時のリダイレクト先URL
 * @returns {Ajax JSON} message 失敗した時のエラーメッセージ
 */
func login(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	mail := r.FormValue("mail")
	pass := r.FormValue("pass")
	
	model := NewModel(c)
	key, _ := model.loginCheck(mail, pass)
	if key != "" {
		fmt.Fprintf(w, `{"result":true, "to":"/gamelist?key=%s"}`, key)
	} else {
		fmt.Fprintf(w, `{"result":false, "message":"メールアドレスまたはパスワードが間違っています"}`)
	}
}

/**
 * 仮登録
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func interimRegistration(w http.ResponseWriter, r *http.Request) {
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

/**
 * 本登録する
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func registration(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	key := r.FormValue("key")
	
	model := NewModel(c)
	model.registration(key)
	
	view := NewView(c, w)
	view.registration()
}

/**
 * Facebookからのコールバック
 */
func callbackFacebook(w http.ResponseWriter, r*http.Request) {
	c := appengine.NewContext(r)
	userInfo := requestFacebookToken(w, r)
	
	model := NewModel(c)
	view := NewView(c, w)
	
	if model.existOAuthUser("Facebook", userInfo["oauth_id"]) {
		// 既存のユーザ
		params := make(map[string]string, 2)
		params["OAuthId"] = userInfo["oauth_id"]
		params["Type"] = "Facebook"
		key := model.getUserKey(params)
		if key != "" {
			view.editor(key)
		}
	} else {
		// 新規ユーザ
		params := make(map[string]string, 4)
		params["user_type"] = "Facebook"
		params["user_name"] = userInfo["name"]
		params["user_oauth_id"] = userInfo["oauth_id"]
		params["user_pass"] = ""
		user := model.NewUser(params)
		key := model.addUser(user)
		view.editor(key)
	}
}

/**
 * ゲーム一覧の表示
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func gamelist(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
//	key := r.FormValue("key")
	
//	model := NewModel(c)
//	games := model.getGameList(key)

	view := NewView(c, w)
//	view.gamelist(games)
	view.gamelist()
}

/**
 * ゲームの新規追加
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func addGame(w http.ResponseWriter, r *http.Request) {
	
}

/**
 * 仮登録ユーザ一覧の取得
 * Ajax で呼び出す
 * @function
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func getInterimUsers(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	model := NewModel(c)
	interimUsers := model.getInterimUsers()
	
	// キーと名前だけを返す
	result := make(map[string]string, len(interimUsers))
	for key, val := range interimUsers {
		result[key] = val.Name
	}
	
	bytes, err := json.Marshal(result)
	check(c, err)
	fmt.Fprintf(w, "%s", bytes)
}

/**
 * ユーザ一覧の取得
 * Ajax で呼び出す
 * @function
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func getUsers(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	model := NewModel(c)
	users := model.getAllUser()
	
	result := make(map[string]string, len(users))
	for key, val := range users {
		result[key] = val.Name
	}
	
	bytes, err := json.Marshal(result)
	check(c, err)
	fmt.Fprintf(w, "%s", bytes)
}