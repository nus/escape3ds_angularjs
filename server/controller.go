/**
 + controller.go
 * ユーザからのリクエストに合わせて処理を振り分ける
 * @file
 */
package escape3ds

import (
	"net/http"
	"appengine"
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
	
	oauth := NewOAuth(config)
	oauth.createHeader()
}