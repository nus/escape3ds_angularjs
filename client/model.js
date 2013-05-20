/**
 * ゲームクラス
 * @class
 * @param {Object} config 初期設定オブジェクト
 * config = {
 *     id: ゲームid
 *     name: ゲーム名
 * }
 * @property {String} id ゲームのID
 * @property {String} name ゲーム名
 * @property {Array(Scene)} scenes シーンリスト
 * @property {Number} firstScene ゲーム開始時のシーン番号
 * @property {Array(Item)} items アイテムリスト
 */
var Game = function(config) {
	this.id = config.id;
	this.name = config.name;
	this.scenes = [];
	this.firstScene = null;
	this.items = [];
};

/**
 * ゲームにシーンを追加する
 * @method
 * @param {Scene} scene 追加するシーン
 */
Game.prototype.addScene = function(scene) {
	if(!scene instanceof Scene) {
		console.log('不正なシーンを追加しようとしました');
		return false;
	}
	this.scenes.push(scene);
};

/**
 * シーンクラス
 * @class
 * @param {Object} config 初期設定オブジェクト
 * config = {
 *     id: シーンID
 *     name: シーン名
 * }
 * @property {String} id シーンID
 * @property {String} name シーン名
 * @property {String} background 背景画像のURL
 * @property {Event} シーン開始時に実行するイベント
 * @property {Event} シーン終了時に実行するイベント
 */
var Scene = function(config) {
	this.id = config.id;
	this.name = config.name;
	this.background = '';
	this.enter = null;
	this.leave = null;
};

/**
 * イベントクラス
 * タッチされたときに処理を実行する画像または範囲
 * @class
 * @param {Object} config 初期設定オブジェクト
 * config = {
 *     id: イベントID
 *     name: イベント名
 *     x: シーンの左上の基準としたX座標
 *     y: シーンの左上の基準としたY座標
 *     width: イベントの幅
 *     height: イベントの高さ
 * }
 * @property {String} id イベントID
 * @property {String} name イベント名
 * @property {Number} x X座標
 * @property {Number} y Y座標
 * @property {Number} width 幅
 * @property {Number} height 高さ
 * @property {String} code イベントの処理内容 専用のスクリプト
 * @property {String} image 背景画像のURL 設定されていない場合は透明な範囲を表す
 */
var Event = function(config) {
	this.id = config.id;
	this.name = config.name;
	this.x = config.x;
	this.y = config.y;
	this.width = config.width;
	this.height = config.height;
	this.code = '';
	this.image = '';
};

/**
 * アイテムクラス
 * 十字キーの左右で選択できる
 * 選択した状態でイベントをタッチすると特別な処理を実行できる
 * @class
 * @param {Object} config 初期設定オブジェクト
 * config = {
 *     id: アイテムID
 *     name: アイテム名
 *     image: 画像のURL
 * }
 * @property {String} id アイテムID
 * @property {String} name アイテム名
 * @property {String} image 画像のURL
 */
var Item = function(config) {
	this.id = config.id;
	this.name = config.name;
	this.image = config.image;
};