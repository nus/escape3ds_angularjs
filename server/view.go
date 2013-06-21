/**
 * HTMLの表示
 * @file
 */
package escape3ds

import(
	"net/http"
	"html/template"
	"appengine"
)

/**
 * 画面表示を行うクラス
 * @class
 * @property {appengine.Context} c コンテキスト
 * @property {http.ResponseWriter} w 応答先
 */
type View struct {
	c appengine.Context
	w http.ResponseWriter
}

/**
 * View の作成
 * @function
 * @param {appengine.Context} c コンテキスト
 * @returns {*View} 作成したView
 */
func NewView(c appengine.Context, w http.ResponseWriter) *View {
	view := new(View)
	view.c = c
	view.w = w
	return view
}

/**
 * ログイン画面を表示する
 * @method
 * @memberof View
 */
func (this *View) login() {
	t, err := template.ParseFiles("server/html/login.html")
	check(this.c, err)
	t.Execute(this.w, nil)
}

/**
 * エディタ画面を表示する
 * @method
 * @memberof View
 * @param {string} key ユーザキー
 */
func (this *View) editor(key string) {
	t, err := template.ParseFiles("server/html/editor.html")
	check(this.c, err)
	t.Execute(this.w, nil)
}

/**
 * デバッグ画面の表示
 * @method
 * @memberof View
 */
func (this *View) debug() {
	t, err := template.ParseFiles("server/html/debug.html")
	check(this.c, err)
	t.Execute(this.w, nil)
}

/**
 * 仮登録ページの表示
 * @method
 * @memberof View
 */
func (this *View) interimRegistration() {
	t, err := template.ParseFiles("server/html/interim_registration.html")
	check(this.c, err)
	t.Execute(this.w, nil)
}

/**
 * 本登録完了ページの表示
 * @method
 * @memberof View
 */
func (this *View) registration() {
	t, err := template.ParseFiles("server/html/registration.html")
	check(this.c, err)
	t.Execute(this.w, nil)
}

/**
 * ゲーム一覧の表示
 * @method
 * @memberof View
 * @param {string} userKey エンコード済みのユーザキー
 */
func (this *View) gamelist(userKey string) {
	t, err := template.ParseFiles("server/html/gamelist.html")
	check(this.c, err)
	m := make(map[string]string, 1)
	m["Key"] = userKey
	t.Execute(this.w, m)
}