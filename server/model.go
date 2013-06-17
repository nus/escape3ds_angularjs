/**
 * データモデルの定義
 * @file
 */
package escape3ds

import (
	"appengine"
	"appengine/datastore"
	"crypto/sha1"
	"strings"
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
 * @property {string} Type ユーザアカウントの種類 "Twitter"/"Facebook"/"normal"
 * @property {string} Name ユーザ名
 * @property {[]byte} Pass ユーザの暗号化済パスワード（user_type == "normal"の場合のみ）
 * @property {string} Mail ユーザのメールアドレス（user_type == "normal"の場合のみ）
 * @property {string} OAuthId OAuthのサービスプロバイダが決めたユーザID
 */
type User struct {
	Type string
	Name string
	Pass []byte
	Mail string
	Salt string
	OAuthId string
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
	user.Type = data["user_type"]
	user.Name = data["user_name"]
	user.Mail = data["user_mail"]
	user.OAuthId = data["user_oauth_id"]
	user.Pass, user.Salt = this.hashPassword(data["user_pass"])
	this.c.Infof("pass: %s", user.Pass)
	return user
}

/**
 * ユーザのパスワードをハッシュ化する
 * @method
 * @memberof Model
 * @param {string} pass 平文パスワード
 * @returns {[]byte} 暗号化されたパスワード
 * @returns {string} 使用したソルト
 */
func (this *Model) hashPassword(pass string) ([]byte, string) {
	salt := ""
	for i := 0; i < 4; i++ {
		salt = strings.Join([]string{salt, getRandomizedString()}, "")
	}
	pass = strings.Join([]string{pass, salt}, "")
	this.c.Infof("salt: %s", salt)
	hash := sha1.New()
	hash.Write([]byte(pass))
	hashedPass := hash.Sum(nil)
	return hashedPass, salt
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
