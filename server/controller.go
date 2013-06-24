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
	sessionId := getSession(c, r)

	if sessionId != "" {
		http.Redirect(w, r, "/gamelist", 302)
	} else {
		view.login()
	}
}

/**
 * Twitter でログイン
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func loginTwitter(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	oauth := NewOAuth1(c, fmt.Sprintf("http://escape-3ds.appspot.com/callback_twitter"))
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
		key := ""
		if model.existOAuthUser("Twitter", result["user_id"]) {
			// 既存ユーザ
			params := make(map[string]string, 2)
			params["Type"] = "Twitter"
			params["OAuthId"] = result["user_id"]
			key = model.getUserKey(params)
		} else {
			// 新規ユーザ
			params := make(map[string]string, 4)
			params["user_type"] = "Twitter"
			params["user_name"] = result["screen_name"]
			params["user_oauth_id"] = result["user_id"]
			params["user_pass"] = ""
			user := model.NewUser(params)
			key = model.addUser(user)
		}
		if getSession(c, r) == "" {
			startSession(w, r, key)
		}
		view.editor(key)
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
 * Facebookからのコールバック
 * @function
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func callbackFacebook(w http.ResponseWriter, r*http.Request) {
	c := appengine.NewContext(r)
	userInfo := requestFacebookToken(w, r)
	
	model := NewModel(c)
	view := NewView(c, w)
	
	key := ""
	if model.existOAuthUser("Facebook", userInfo["oauth_id"]) {
		// 既存ユーザ
		params := make(map[string]string, 2)
		params["OAuthId"] = userInfo["oauth_id"]
		params["Type"] = "Facebook"
		key = model.getUserKey(params)
	} else {
		// 新規ユーザ
		params := make(map[string]string, 4)
		params["user_type"] = "Facebook"
		params["user_name"] = userInfo["name"]
		params["user_oauth_id"] = userInfo["oauth_id"]
		params["user_pass"] = ""
		user := model.NewUser(params)
		key = model.addUser(user)
	}
	
	if getSession(c, r) == "" {
		startSession(w, r, key)
	}
	view.editor(key)
}

/**
 * エディタの表示
 * @param {http.ResponseWRiter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func editor(w http.ResponseWriter, r *http.Request) {
	userKey := session(w, r)
	c := appengine.NewContext(r)
	model := NewModel(c)

	gameKey := r.FormValue("game_key")
	if gameKey == "" {
		c.Warningf("ゲームキー無しでゲームを編集しようとしました")
		http.Redirect(w, r, "/", 302)
	}
	
	game := model.getGame(gameKey)
	if game.UserKey != userKey {
		c.Warningf("ユーザキー: %s が他人のゲーム: %s を編集しようとしました", userKey, gameKey)
		http.Redirect(w, r, "/gamelist", 302)
	}
	
	view := NewView(c, w)
	view.editor(gameKey)
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
 * ログイン
 * @param {http.ResponseWriter} w 応答先
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
		// ログイン成功
		sessionId := getSession(c, r)
		if sessionId == "" {
			startSession(w, r, key)
		}
		fmt.Fprintf(w, `{"result":true, "to":"/gamelist"}`)
	} else {
		// ログイン失敗
		fmt.Fprintf(w, `{"result":false, "message":"メールアドレスまたはパスワードが間違っています"}`)
	}
}

/**
 * ログアウト
 * クッキーとmemcacheに保存されたセッション情報を削除する
 * @function
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func logout(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	cookie, err := r.Cookie("escape3ds")
	check(c, err)
	sessionId := cookie.Value
	
	model := NewModel(c)
	model.removeSession(sessionId)
	deleteCookie(w)
	
	http.Redirect(w, r, "/", 302)
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
 * @function
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
 * ゲーム一覧の表示
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func gamelist(w http.ResponseWriter, r *http.Request) {
	userKey := session(w, r)
	c := appengine.NewContext(r)
	view := NewView(c, w)
	view.gamelist(userKey)
}

/**
 * ゲームの新規追加
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func addGame(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	gameName := r.FormValue("game_name")
	gameDescription := r.FormValue("game_description")
	
	if gameName == "" {
		fmt.Fprintf(w, `{"result": false}`)
		c.Warningf("空のゲーム名でゲームを作成しようとしました")
	} else if gameDescription == "" {
		fmt.Fprintf(w, `{"result": false}`)
		c.Warningf("ゲーム説明文が空のゲームを作成しようとしました")
	}

	sessionId := getSession(c, r)
	if sessionId == "" {
		fmt.Fprintf(w, `{"result": false}`)
		c.Warningf("セッションIDなしでゲームを作成しようとしました")
	}
	
	model := NewModel(c)
	userKey := model.getUserKeyFromSession(sessionId)
	params := make(map[string]string, 4)
	params["name"] = gameName
	params["description"] = gameDescription
	params["thumbnail"] = ""
	params["user_key"] = userKey
	game := model.NewGame(params)
	model.addGame(game)
	
	fmt.Fprintf(w, `{"result":true, "name":"%s", "description":"%s"}`, gameName, gameDescription)
}

/**
 * ゲームの削除
 * ゲームの所有者しか削除することはできない
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 */
func deleteGame(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	sessionId := getSession(c, r)
	gameKey := r.FormValue("game_key")
	
	if sessionId == "" {
		fmt.Fprintf(w, `{"result":false}`)
		c.Warningf("セッションIDなしで deleteGame() が呼び出されました")
		return
	} else if gameKey == "" {
		fmt.Fprintf(w, `{"result":false}`)
		c.Warningf("ゲームキー無しで deleteGame() が呼び出されました")
		return
	}
	
	model := NewModel(c)
	userKey := model.getUserKeyFromSession(sessionId)
	game := model.getGame(gameKey)
	if game.UserKey != userKey {
		fmt.Fprintf(w, `{"result":false}`)
		c.Warningf("ユーザキー: %s が他のユーザのゲームを削除しようとしました", userKey)
		return
	}
	
	model.deleteGame(gameKey)
	fmt.Fprintf(w, `{"result":true}`)
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

/**
 * セッションを開始する
 * ユーザーキーに関連付いたセッションIDを生成して memcache, cookie に保存する。
 * @function
 * @param w {http.ResponseWriter} w 応答先
 * @param r {*http.Request} r リクエスト
 * @param {string} key ユーザのキー
 * @returns {string} セッションID
 */
func startSession(w http.ResponseWriter, r *http.Request, key string) {
	c := appengine.NewContext(r)
	model := NewModel(c)
	sessionId := model.startSession(key)
	cookie := NewCookie("escape3ds", sessionId, "localhost", "/", 24)
	http.SetCookie(w, cookie)
}

/**
 * Cookieに保存されているセッションIDを取得する
 * セッションが存在しない場合は空文字を返す
 * @function
 * @param {appengine.Context} c コンテキスト
 * @param {*http.Request} r リクエスト
 * @returns {string} セッションIDまたは空文字
 */
func getSession(c appengine.Context, r *http.Request) string {
	var result string
	cookie, err := r.Cookie("escape3ds")
	if err == http.ErrNoCookie {
		result = ""
	} else if err != nil {
		result = ""
		c.Errorf(err.Error())
	} else {
		result = cookie.Value
	}
	return result
}

/**
 * セッションを終了する
 * @function
 * @param
 * @param {string} sessionId
 */
func closeSession(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	sessionId := getSession(c, r)
		
	model := NewModel(c)
	model.removeSession(sessionId)
	deleteCookie(w)
}

/**
 * セッションIDが保存されたクッキーを削除する
 * @function
 * @param {http.ResponseWriter} w 応答先
 */
func deleteCookie(w http.ResponseWriter) {
	cookie := NewCookie("escape3ds", "", "localhost", "/", -1)
	http.SetCookie(w, cookie)
}

/**
 * セッションチェック
 * ページ表示時に必ず実行する
 * クライアントがセッションIDを持っているかどうか調べて
 * 持っていなければトップページへ飛ばして空文字を返す
 * 持っていたら対応するユーザIDを返す
 * @function
 * @param {http.ResponseWriter} w 応答先
 * @param {*http.Request} r リクエスト
 * @returns {string} ユーザキー
 */
func session(w http.ResponseWriter, r *http.Request) string {
	c := appengine.NewContext(r)
	sessionId := getSession(c, r)
	if sessionId == "" {
		c.Warningf("セッションIDなしで内部へ入ろうとしました")
		http.Redirect(w, r, "/", 302)
		return ""
	}
	
	model := NewModel(c)
	userKey := model.getUserKeyFromSession(sessionId)
	if userKey == "" {
		c.Warningf("セッションID: %s に該当するユーザーキが存在しません")
		http.Redirect(w, r, "/", 302)
		return ""
	}
	
	return userKey
}


