
(function (exports) {

	var basic_func = function (type, api, data, suCallBack, errCallBack, contentType, async) {
		$.ajax({
			type: type,
			url: api,
			data: data,
			cache: false,
			async: (async === undefined  || async) ? true : false,
			contentType: (!contentType || contentType.length === 0) ? 'application/x-www-form-urlencoded; charset=UTF-8' : contentType,
			xhrFields: {
				withCredentials: true
			},
			success: suCallBack,
			error: errCallBack
		});
	};

	var new_recode = function(data, suCallback, failCallback) {
		basic_func("POST", "/recodes", data, suCallback, failCallback);
	};

	exports.ddns_new_recode = new_recode;

})((typeof (exports) === "object" ? exports : window), jQuery);