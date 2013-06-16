/**
 * データモデルの定義
 * @file
 */
package escape3ds

import (
	"appengine"
	"appengine/datastore"
)

/**
 * モデル
 * @class
 * @property {appengine.Context} c コンテキスト
 */
type Model struct {
	c appengine.Context
}

/**
 * モデルの作成
 * @function
 * @param {appengine.Context} c コンテキスト
 * @returns {*Model} モデル
 */
func NewModel(c appengine.Context) *Model {
	model := new(Model)
	model.c = c
	return model
}

/**
 * ユーザデータ
 * @struct
 * @property {string} user_type ユーザアカウントの種類 "Twitter"/"Facebook"/"normal"
 * @property {string} user_name ユーザ名
 * @property {string} user_pass ユーザのパスワード（user_type == "normal"の場合のみ）
 * @property {string} user_mail ユーザのメールアドレス（user_type == "normal"の場合のみ）
 * @property {string} user_oauth_id OAuthのサービスプロバイダが決めたユーザID
 */
type User struct {
	user_id string
	user_type string
	user_name string
	user_pass string
	user_mail string
	user_salt string
	user_oauth_id string
}

/**
 * ユーザを作成する
 * @method
 * @memberof Model
 * @param {map[string]string} ユーザの設定項目を含んだマップ
 * @returns {*User} ユーザ、失敗したらnil
 */
func (this *Model) NewUser(data map[string]string) *User {
	// ユーザタイプチェック
	if !exist([]string {"Twitter", "Facebook", "normal"}, data["user_type"]) {
		this.c.Errorf("不正なユーザタイプが入力されました")
		return nil
	}
	
	// OAuthアカウントチェック
	if data["user_type"] == "Twitter" || data["user_type"] == "Facebook" {
		if data["user_oauth_id"] == "" {
			this.c.Errorf("OAuthアカウントのidが設定されていません")
			return nil
		}
	}
	
	// 通常アカウントチェック
	if data["user_type"] == "normal"{
		if data["user_mail"] == "" {
			this.c.Errorf("メールアドレスが入力されていません")
			return nil
		}
		if data["user_pass"] == "" {
			this.c.Errorf("パスワードが入力されていません")
			return nil
		}
	}
	
	user := new(User)
	user.user_type = data["user_type"]
	user.user_name = data["user_name"]
	user.user_pass = data["user_pass"]
	user.user_mail = data["user_mail"]
	user.user_oauth_id = data["user_oauth_id"]
	return user
}

/**
 * ユーザの追加
 * @method
 * @memberof Model
 */
func (this *Model) addUser(data map[string]string) {
	user := this.NewUser(data)
	if user == nil {
		this.c.Errorf("ユーザの追加を中止しました")
		return
	}
	key := datastore.NewIncompleteKey(this.c, "User", nil)
	_, err := datastore.Put(this.c, key, user)
	check(this.c, err)
}
