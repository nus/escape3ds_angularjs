/**
 * デバッグページのスクリプト
 * @file
 */
$(function() {
	// ユーザの追加
	$('#add_user .submit').click(function() {
		var div = $('#add_user')
		var data = {};
		data.user_type = div.find('.user_type option:selected').val();
		data.user_name = div.find('.user_name').val();
		data.user_pass = div.find('.user_pass').val();
		data.user_mail = div.find('.user_mail').val();
		data.user_oauth_id = div.find('.user_oauth_id').val();
		
		$.ajax('/add_user', {
			method: 'POST',
			data: data,
			error: function() {
				console.log('error');
			},
			success: function() {
				console.log('success');
			}
		});
	});
	
	// ログイン
	$('#login .submit').click(function() {
		var div = $('#login');
		var data = {};
		data.mail = div.find('.mail').val();
		data.pass = div.find('.password').val();
		
		$.ajax('/login', {
			method: 'POST',
			data: data,
			dataType: 'json',
			error: function(xhr, status) {
				console.log(status);
			},
			success: function(data) {
				if(data.result == false) {
					alert(data.message);
				} else {
					location.href = data.to;
				}
			}
		});
	});
});