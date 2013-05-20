// ルーティング
angular.module('escape3ds', []).config(['$routeProvider', function($routeProvider) {
	$routeProvider.when('/edit', {
		templateUrl: '/client/editor.html',
		controller: EditPageController
	});
	$routeProvider.otherwise({redirectTo: '/edit'});
}]);

/**
 * 編集ページの処理
 * @class
 * @param {object} $scope ゲームを表すデータ
 */
var EditPageController = function($scope) {
	$scope.game = new Game({
		id: '0001',
		name: 'サンプルゲーム'
	});
	var scene = new Scene({
		id: '---1',
		name: '最初の部屋'
	});
	$scope.game.addScene(scene);
};
