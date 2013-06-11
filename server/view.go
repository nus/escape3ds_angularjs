/**
 * HTMLの表示
 */
package escape3ds

import(
	"net/http"
	"html/template"
	"appengine"
)

type View struct {

}

/**
 * ログイン画面を表示する
 * @method
 * @memberof View
 * @param {appengine.Context} c コンテキスト
 * @param {http.ResponseWriter} w 応答先
 */
func (this *View) login(c appengine.Context, w http.ResponseWriter) {
	t, err := template.ParseFiles("server/html/login.html")
	check(c, err)
	t.Execute(w, c)
}

/**
 * エディタ画面を表示する
 * @method
 * @memberof View
 * @param {appengine.Context} c コンテキスト
 * @param {http.ResponseWriter} w 応答先
 */
func (this *View) editor(c appengine.Context, w http.ResponseWriter) {
	t, err := template.ParseFiles("server/html/editor.html")
	check(c, err)
	t.Execute(w, c)
}