/**
 * ゲーム一覧ページのスクリプト
 * @file
 */
$(function() {
	// ゲームの新規作成
	$('#add_game').click(function() {
		var div = $('#add_game_div');
		var name = div.find('.name').val();
		var description = div.find('.description').val();
		if(name == "") {
			alert('ゲームの名前が入力されていません');
			return false;
		} else if(description == "") {
			alert('ゲームの説明が入力されていません');
			return false;
		}
		$.ajax('/add_game', {
			method: 'POST',
			data: {
				user_key: userKey,
				game_name: name,
				game_description: description
			},
			dataType: 'json',
			success: function(data) {
			},
			error: function(xhr, err) {
				console.log(err);
			}
		});
	});
});