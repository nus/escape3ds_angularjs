/**
 * ログインページのスクリプト
 * @file
 */
$(function() {

	// ログイン処理
	var loginForm = $('.normal_login');
	loginForm.find('.submit').click(function() {
		var mail = loginForm.find('.mail').val();
		var pass = loginForm.find('.pass').val();
		$.ajax('/login', {
			method: 'POST',
			data: {
				mail: mail,
				pass: pass
			},
			dataType: 'json',
			success: function(data) {
				if(data.result == false) {
					alert(data.message);
				} else {
					location.href = data.to;
				}
			},
			error: function() {
				console.log('error');
			}
		});
	});
});