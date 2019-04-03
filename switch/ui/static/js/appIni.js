// script to initialize all thirdparty plugins and perform additional app setup
function appIni() {

	window.app.layoutDefaults = {
		objectView: {
			west__size: 300,
			west__minSize: 260,
			west__maxSize: 600,
			west__togglerContent_open: '&#8249;',
			west__togglerContent_closed: '&#8250;',
			resizerClass: 'search-layout-resizer',
			togglerClass: 'search-layout-toggler'
		}
	};

	// only fade screen once
	$(document).on('show.bs.modal', '.modal', function() {

		var zIndex = 1040 + (10 * $('.modal:visible').length);
		$(this).css('z-index', zIndex);

		setTimeout(function() {
			$('.modal-backdrop').not('.modal-stack').css('z-index', zIndex - 1).addClass('modal-stack');
		}, 0);
	});


	// check for session timeout
	$(document).ajaxError(function(e, xhr, settings, error) {
		if (xhr.status == 401 || xhr.status == 403) {
			window.location.href = window.location.href;
		}
	});

	// master ajax setup ini
	$.ajaxSetup({
		async: true,
		cache: false,
		beforeSend: function(xhr, settings) {
			/* placeholder for preflight capturing */
		}
	});


	google.maps.Polygon.prototype.getBounds = function() {

		var bounds = new google.maps.LatLngBounds();
		var paths = this.getPaths();
		var path;

		for (var i = 0; i < paths.getLength(); ++i) {
			path = paths.getAt(i);
			for (var i2 = 0; i2 < path.getLength(); ++i2) {
				bounds.extend(path.getAt(i2));
			}
		}

		return bounds;
	};

	// template parsing
	_.templateSettings = {
		interpolate: /\{\{(.+?)\}\}/g, // print value: {{ value_name }}
		evaluate: /\{%([\s\S]+?)%\}/g, // excute code: {% code_to_execute %}
		escape: /\{%-([\s\S]+?)%\}/g // excape HTML: {%- <script> %} prints &lt;script&gt
	};


	/*
	 * Backbone overloads
	 */

	Backbone.View.prototype.resize = function() {

		// if it's not being shown, don't resize
		if (!this.$el.parent().length) {
			return;
		}

		if (this.views) {
			_.each(this.views, function(view) {
				if (view.$el.parent().length) {
					if (view.$el.parent().width() > 0 && view.$el.parent().height() > 0) {
						view.resize();
					}
				}
			});
		}
	};

	Backbone.View.prototype.show = function() {

		this.$parent.append(this.$el);
		this.resize();

		if (this.currentView) {
			this.currentView.show();
		}
	};

	Backbone.View.prototype.reload = function() {
		this.hide();

		this.$el.empty();
		this.$el = $(this.el);

		this.show();

		this.render();
		this.delegateEvents();
	};

	Backbone.View.prototype.hide = function() {

		if (this.currentView) {
			this.currentView.hide();
		}

		this.$el.detach();
	};

	Backbone.View.prototype.route = function() {
		this.show();
	};

	Backbone.View.prototype.error = function(e) {
		console.error('Failed to communicate with the server:', e);

		var errorMsg = 'Failed to communicate with the server.';

		if (this.loadingView) {
			this.loadingView.hide();
		}
		if (this.errorView) {
			this.errorView.hide();
		}

		this.errorView = new window.app.views.Error({
			$parent: this.$parent,
			errorMessage: errorMsg
		});
	};

	// perform scan for data-ref attributes and make the appropriate replacements via the passed model
	Backbone.View.prototype.bindDataRefs = function(model, inputToModel) {
		var self = this;

		inputToModel = !!inputToModel;
		obj = model.attributes;

		if (obj) {
			this.$('[data-test]').each(function() {
				/*$input = $(this);
				var refFunc = $input.data('test');
			
				var func = refFunc.substring(0, refFunc.)
			
				if (refFunc) {
					var refOwner = eval(refFunc.substring(0, refFunc.lastIndexOf('.')));
					$input.html(eval(refFunc));
				}*/
			});
			this.$('[data-ref]').each(function() {
				$input = $(this);

				if (!inputToModel) {
					inputToModel = false;
				}

				var refPath = $input.data('ref');
				var refFunc = $input.data('func');

				var objVar = ref(obj, refPath);
				var objVarOwner = ref(obj, refPath.substring(0, refPath.lastIndexOf('.')));
				var objVarName = refPath.split('.').pop();

				if (refFunc) {
					var refOwner = eval(refFunc.substring(0, refFunc.lastIndexOf('.')));
					objVar = eval(refFunc).call(refOwner, objVar);
				}

				if (!objVarOwner) {
					// create the path
					populatePath(obj, refPath.substring(0, refPath.lastIndexOf('.')));
				}

				var $inputActivator = $searchDOMUp($input, 'input.activator', 3);

				var type = $input.prop('tagName').toLowerCase() == 'input' ? $input.attr('type') : $input.prop('tagName');
				type = type.toLowerCase();

				if (!inputToModel) {
					switch (type) {
						case 'text':
						case 'select':
							$input.val(objVar);
							break;
						case 'checkbox':
							$input.prop('checked', parseMap($input.data('ref-map'), objVar));
							break;
						default:
							$input.html(objVar);
					}
				} else {
					switch (type) {
						case 'text':
						case 'select':
							objVarOwner[objVarName] = $input.val();
							break;
						case 'checkbox':
							objVarOwner[objVarName] = parseMap($input.data('ref-map'), $input.prop('checked'), true);
							break;
						default:
							objVarOwner[objVarName] = $input.html();
					}
				}
			});
		}
	};


	// overide backbones sync function (small change to error and complete functions below; rest is stock)
	Backbone.sync = function(method, model, options) {
		var type = methodMap[method];

		// trigger loading
		if (type == 'GET') {
			model.loaded = false;
			model.trigger('loading');
		}

		// Default options, unless specified.
		_.defaults(options || (options = {}), {
			emulateHTTP: Backbone.emulateHTTP,
			emulateJSON: Backbone.emulateJSON
		});

		// Default JSON-request options.
		var params = {
			type: type,
			dataType: 'json'
		};

		if (!options.url) {
			params.url = _.result(model, 'url');
		}

		// Ensure that we have the appropriate request data.
		if (options.data == null && model && (method === 'create' || method === 'update' || method === 'patch')) {
			params.contentType = 'application/json';
			params.data = JSON.stringify(options.attrs || model.toJSON(options));
		}

		// For older servers, emulate JSON by encoding the request into an HTML-form.
		if (options.emulateJSON) {
			params.contentType = 'application/x-www-form-urlencoded';
			params.data = params.data ? {
				model: params.data
			} : {};
		}

		// For older servers, emulate HTTP by mimicking the HTTP method with `_method`
		// And an `X-HTTP-Method-Override` header.
		if (options.emulateHTTP && (type === 'PUT' || type === 'DELETE' || type === 'PATCH')) {
			params.type = 'POST';
			if (options.emulateJSON) params.data._method = type;
			var beforeSend = options.beforeSend;
			options.beforeSend = function(xhr) {
				xhr.setRequestHeader('X-HTTP-Method-Override', type);
				if (beforeSend) return beforeSend.apply(this, arguments);
			};
		}

		// Don't process data on a non-GET request.
		if (params.type !== 'GET' && !options.emulateJSON) {
			params.processData = false;
		}

		if (params.type != 'GET') {
			window.app.syncView.show(model.cid);
		}

		// Pass along `textStatus` and `errorThrown` from jQuery.
		var error = options.error;
		options.error = function(xhr, textStatus, errorThrown) {
			options.textStatus = textStatus;
			options.errorThrown = errorThrown;

			window.app.displayAppError('A server API error has occured: ' + errorThrown);

			if (error) error.call(options.context, xhr, textStatus, errorThrown);
		};

		var complete = options.complete;
		var context = options.context;
		options.complete = function(e, xhr, options) {

			window.app.syncView.hide(model.cid);

			model.loaded = true;
			model.trigger('loaded');

			if (complete) complete.call(context, e, xhr, options);
		};

		// Make the request, allowing the user to override any Ajax options.
		var xhr = options.xhr = Backbone.ajax(_.extend(params, options));

		model.trigger('request', model, xhr, options);

		return xhr;
	};

	// overide methods too
	var methodMap = {
		'create': 'POST',
		'update': 'PUT',
		'patch': 'PATCH',
		'delete': 'DELETE',
		'read': 'GET'
	};


	/************************************************
		TODO
	************************************************/

	//moment.tz.setDefault("America/Los_Angeles);


	// init localization
	/*
	$.i18n.init({
		lng: 'en-US',
		useCookie: false,
		fallbackLng: false,
		load: 'current',
		ns: 'app',
		debug: true,
		getAsync: false,
		postAsync: false,
		//resGetPath: concatURLs(window.app.rootPath, window.app.localesPath, '__lng__/__ns__.json')
	    //interpolationPrefix: '{{',
	    //interpolationSuffix: '}}',
	});
	*/

}