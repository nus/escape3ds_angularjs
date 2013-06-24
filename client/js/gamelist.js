/**
 * ゲーム一覧ページのスクリプト
 * @file
 */
$(function() {

	// ゲーム新規作成ボタン
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
	
	// ゲーム削除ボタン
	$('.game .delete').click(function() {
		if(!window.confirm('ゲームを削除しますか？')) {
			return false;
		}
		var key = $(this).parent('.game').attr('key');
		$.ajax('/delete_game', {
			method: 'POST',
			dataType: 'json',
			data: {
				game_key: key
			},
			error: function() {
				console.log('delete game error');
			},
			success: function() {
				console.log('success');
			}
		});
	});
});