/**
 * データモデルの定義
 * @file
 */
package escape3ds

import (
	"appengine"
	"appengine/datastore"
	"strings"
	"bytes"
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
	user.Pass, user.Salt = this.hashPassword(data["user_pass"], "")
	return user
}

/**
 * ユーザのパスワードをハッシュ化する
 * @method
 * @memberof Model
 * @param {string} pass 平文パスワード
 * @param {string} salt ソルト。空文字が渡された場合は自動で作成する。
 * @returns {[]byte} 暗号化されたパスワード
 * @returns {string} 使用したソルト
 */
func (this *Model) hashPassword(pass string, salt string) ([]byte, string) {
	if salt == "" {
		for i := 0; i < 4; i++ {
			salt = strings.Join([]string{salt, getRandomizedString()}, "")
		}
	}
	pass = strings.Join([]string{pass, salt}, "")
	hashedPass := SHA1(pass)
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

/**
 * 指定されたメールアドレスとパスワードのユーザがいるか調べる
 * 存在しない場合は戻り値がすべて空文字になる
 * @method
 * @memberof Member
 * @returns {string} エンコードされたキー
 * @returns {string} ユーザ名
 */
func (this *Model) loginCheck(mail string, pass string) (string, string) {
	query := datastore.NewQuery("User").Filter("Mail =", mail)
	iterator := query.Run(this.c)
	key, err := iterator.Next(nil)
	
	if err != nil {
		this.c.Warningf("存在しないメールアドレスによるログインが試されました。アドレス：%s", mail)
		return "", ""
	}
	
	encodedKey := key.Encode()
	user := this.getUser(encodedKey)
	
	hashedPass, _ := this.hashPassword(pass, user.Salt)
	if bytes.Compare(user.Pass, hashedPass) != 0 {
		this.c.Warningf("間違ったパスワードが試されました。アドレス：%s", mail)
		return "", ""
	}
	
	return encodedKey, user.Name
}

/**
 * ユーザの取得
 * @method
 * @memberof Model
 * @param {string} encodedKey エンコードされたキー
 */
func (this *Model) getUser(encodedKey string) *User {
	key, err := datastore.DecodeKey(encodedKey)
	check(this.c, err)
	
	user := new(User)
	err = datastore.Get(this.c, key, user)
	check(this.c, err)
	
	return user
}