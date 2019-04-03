
window.app = {
	views: {},
	models: {},
	collections: {},
	templates: {},
	router: {},
	dataPorts: {},
	timezones: {},

	// local storage of re-usable models/collections
	localData: {
		models: {},
		collections: {}
	},

	errorView: null,
	displayAppError: function(errorMsg) {
		var self = this;

		if (this.errorView) {
			errorView.addError(errorMsg);
		} else {
			this.errorView = new window.app.views.Modal({
				error: true,
				message: errorMsg,
				onAction: function() {
					self.errorView = null;
				}
			});
		}
	},

	// for blocking usage during critical server sync events
	syncView: {},

	name: 'AMP Console',
	version: '1.0',
	date: '01/12/2017',

	compiled: false,

	rootPath: '',
	relPath: 'static/js',
	tplPath: 'static/tpls',
	localesPath: 'locales',

	dataSource: 'AppData.json',
	timezoneSource: 'Timezones.json',

	// some rest services respond with http error codes (4xx), but not because of a failure
	validHTTPErrorResponses: [404],

	title: function() {
		return this.name;
	},

	initialize: function() {
		var self = this;

		// run ini script
		appIni();

		// load the timezones
		$.getJSON(concatURLs(this.rootPath, this.relPath, this.timezoneSource), function(json) {
			moment.tz.load(json);
		});

		// load the master template file if this build is compiled
		if (self.compiled) {
			$.ajax({
				url: concatURLs(self.rootPath, self.tplPath, self.data.build.tplSource),
				dataType: 'text',
				async: true,
				error: function() {
					console.error('Unable to load compiled template file.');
				},
				success: function(data) {
					$('body').append(data);
					self.initializationComplete();
				}
			});
		} else {
			SourceDependencies.loadSources(self.data.sources, function() {
				self.initializationComplete();
			});
		}

		/* not done yet (the initializationComplete function will be called after the scripts load) */
	},

	initializationComplete: function() {
		var self = this;

		// sync view singleton
		this.syncView = new this.views.Sync();

		// show framework view (template causes bitching about xhr on core thread)
		this.views.App = new this.views.App();

		//define router class
		var AppRouter = Backbone.Router.extend({
			routes: {
				'App/:p': 'appRouter',
				'App/:p/*path': 'appRouter'
			},

			appRouter: function(page, path) {
				self.views.App.route(page, path);
			}
		});

		// start routing requests
		this.router = new AppRouter();
		Backbone.history.start();

		// important global stuff
		this.$domPhantom = $('body').append('<div id="dom-phantom" style="display:none"></div>');

		// let there be light!
		$(this).trigger('initialized');

		console.log('all scripts loaded in ' + ((new Date()).getTime() - window.beginLoadTime.getTime()) + ' ms');

		// load the default page in the absence of one
		if (!location.hash) {
			self.views.App.route('Nodes');
		}
	},

	navTo: function(urlHash) {
		this.router.navigate('#', {
			trigger: false,
			replace: false
		});
		this.router.navigate(urlHash, {
			trigger: true,
			replace: true
		});
	},

	printlatLng: function(latLng) {
		if (!latLng) {
			return '';
		}
		return latLng[0].toFixed(4) + ', ' + latLng[1].toFixed(4);
	},

	printDatetime: function(timestamp) {
		if (!moment.isMoment(timestamp)) {
			timestamp = moment.utc(timestamp);
		}
		return timestamp.tz(this.data.session.locale.timezone).format(window.app.data.session.locale.datetimeFormat);
	},

	printTimeDiff: function(timestamp1, timestamp2) {
		if (!moment.isMoment(timestamp1)) {
			timestamp1 = moment.utc(timestamp1);
		}
		if (!moment.isMoment(timestamp2)) {
			timestamp2 = moment.utc(timestamp2);
		}

		return moment.utc(timestamp1.diff(timestamp2)).format('DD HH:mm');
	},

	momentInLocaleTZ: function(obj) {
		if (!moment.isMoment(obj)) {
			return null;
		}
		return obj.tz(this.data.session.locale.timezone);
	},

	getView: function(name) {
		var view = this.views[name] || null;
		return view;
	},

	// load a html template
	loadTemplate: function(name, data) {

		var retTemplate = null;

		if (!this.templates[name]) {
			var templateString;

			if (this.compiled) {
				var templateName = '';
				if (name.lastIndexOf('/') > 0) {
					templateName += name.substr(name.lastIndexOf('/') + 1);
				} else {
					templateName += name;
				}

				templateString = $('#' + templateName).html();

				if (!templateString) {
					console.error('failed to load template: ' + templateName);
				}
			} else {
				$.ajax({
					url: concatURLs(this.rootPath, this.tplPath, name) + '.html',
					dataType: 'text',
					method: 'GET',
					async: false,
					success: function(data) {
						if (/^\w*<script/i.test(data)) {
							var matches = /^\w*<script[^>]*>([\S\s]*)<\/script>/i.exec(data);
							data = matches[1];
						}
						templateString = data;
					}
				});
			}

			this.templates[name] = _.template(templateString);
		}

		if (data) {
			retTemplate = this.templates[name](data);
		} else {
			retTemplate = this.templates[name];
		}

		return retTemplate;
	}
};


/* common global namespace stuff */


function subVars(template, kvs, classWrapper) {
	var retString = template;

	for (var key in kvs) {
		retString = retString.replace(new RegExp('\{\{\s*' + key + '\s*\}\}', 'gm'), function(match) {
			if (classWrapper) {
				return '<div class="' + classWrapper + '">' + kvs[key] + '</div>';
			}
			return kvs[key];
		});
	}

	return retString;
};

// parse a key val collection and return the map (direction key -> val or val -> key)
function parseMap(map, input, valToKey) {
	if (!valToKey) {
		return map[input];
	} else {
		var key = null;
		for (key in map) {
			if (map[key] == input) {
				return key;
			}
		}
		return key;
	}
}

// search the DOM up from an element and then check all descendents of each parent for the matching selector (go x levels up)
function $searchDOMUp($element, selector, levels) {

	var $matchedElement = null;

	for (var i = 0; i < levels; ++i) {
		$element = $element.parent();
		if (!$element.length) {
			break;
		}

		$matchedElement = $element.find(selector);
		if ($matchedElement.length) {
			break;
		}
	}

	return $matchedElement;
}

function populatePath(obj, path) {

	var pointer = obj;

	var pathNames = path.split('.');

	_.each(pathNames, function(name) {

		if (pointer[name]) {
			return;
		}

		var matches = name.match(/\[[^\]]*\]/g);
		name = name.split('[')[0];

		// if it's an array
		if (matches) {
			for (var i = 0; i < matches.length; ++i) {
				var match = matches[i];
				var index = parseInt(match.substr(1, match.length - 1));

				pointer[name] = [];
				for (var i2 = 0; i2 <= index; ++i2) {
					pointer[name].push({});
				}
				pointer = pointer[name][index];
			}
		} else {
			// if it's an object
			pointer[name] = {};
			pointer = pointer[name];
		}
	});
}

// eval a path on an object (might cause issue with old ref being replaced)
function ref(obj, path) {
	var val = null;

	try {
		if (!path.length) {
			val = obj;
		} else {
			with(obj) {
				val = eval(path);
			}
		}
	} catch (e) { /* ignore */ }

	return val;
}

function genBind(bindName, classes) {
	return '<div data-bind-name="' + bindName + '" class="' + classes.join(' ') + '"></div>';
}

function rgb2hex(r, g, b) {
	return ('0' + parseInt(r, 10).toString(16)).slice(-2) + ('0' + parseInt(g, 10).toString(16)).slice(-2) + ('0' + parseInt(b, 10).toString(16)).slice(-2);
}

function momentToISOString(moment) {
	return moment.utc().format('YYYY-MM-DDTHH:mm:ssZ');
}

function distanceInMeters(lat1, lng1, lat2, lng2) {

	var R = 6378100; // meters

	var dLat = (lat2 - lat1) * Math.PI / 180;
	var dLon = (lng2 - lng1) * Math.PI / 180;
	var lat1 = lat1 * Math.PI / 180;
	var lat2 = lat2 * Math.PI / 180;

	var a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
		Math.sin(dLon / 2) * Math.sin(dLon / 2) * Math.cos(lat1) * Math.cos(lat2);
	var c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
	var d = R * c;

	return d;
}

function keyValueToArray(kv, keyName, valName) {

	var retArray = [];

	if (!keyName) {
		keyName = 'key';
	}
	if (!valName) {
		valName = 'value';
	}

	for (var key in kv) {
		var item = {};
		item[keyName] = key;
		item[valName] = kv[key];
		retArray.push(item);
	}

	return retArray;
}

// real mod function
Number.prototype.mod = function(n) {
	return ((this % n) + n) % n;
}

function getParameterByName(name) {
	name = name.replace(/[\[]/, "\\[").replace(/[\]]/, "\\]");

	var regex = new RegExp("[\\?&]" + name + "=([^&#]*)");
	var results = regex.exec(location.search);

	return results === null ? '' : decodeURIComponent(results[1].replace(/\+/g, ' '));
}

function concatURLs() {
	var retURL = '';

	if (!arguments || arguments.length === 0) {
		return '';
	}

	try {
		retURL += String(arguments[0]);

		for (var i = 1; i < arguments.length; ++i) {

			var lengthBefore = retURL.length;

			var arg = String(arguments[i]);

			if (arg.length == 0) {
				continue;
			}

			if (retURL.charAt(retURL.length - 1) == '/') {
				if (arg.charAt(0) == '/') {
					retURL += arg.substr(1, arg.length - 1);
				} else {
					retURL += arg;
				}
			} else {
				if (arg.charAt(0) == '/' || lengthBefore == 0) {
					retURL += arg;
				} else {
					retURL += '/' + arg;
				}
			}
		}
	} catch (e) {
		return '';
	}

	return retURL;
}

function substituteObject(obj, ref) {
	var retString = '';

	for (var i in obj) {
		if (i == '%v') {
			retString = _.template(obj[i], {
				v: ref
			})
			break;
		} else if (i == ref) {
			retString = obj[i];
			break;
		}
	}

	return retString;
}

// data ports
function DataPort() {
	this.data = [];
	this.callback = null;
}
DataPort.prototype = {
	read: function() {
		if (this.data.length) {
			return this.data.splice(this.data.length - 1, 1)[0];
		}

		return null;
	},

	write: function(data) {
		this.data.push(data);
		if (this.callback) {
			this.callback.call(this);
		}
	},

	listen: function(callback) {
		this.callback = callback;
		if (this.data.length) {
			this.callback.call(this);
		}
	}
};

// loads sources and handles in referenced order
function SourceDependencies() {}
SourceDependencies.loadSources = function(sources, callback) {

	var sourceDependencies = new SourceDependencies();

	sourceDependencies.sources = sources;
	sourceDependencies.callback = callback || new function() {};

	sourceDependencies._orderSources(sources);
	sourceDependencies._load(0);
};
SourceDependencies.prototype = {

	_orderSources: function() {

		this.orderedSources = [];

		for (var i = 0; i < Object.keys(this.sources).length; ++i) {

			var cleanRun = true;
			for (var source in this.sources) {
				var dependsOn = this.sources[source];
				var indexSource = this.orderedSources.indexOf(source);

				if (dependsOn && dependsOn.length) {

					var maxIndexDependency = -1;
					for (var index in dependsOn) {

						var dependencyIndex = this.orderedSources.indexOf(dependsOn[index]);

						if (dependencyIndex > maxIndexDependency) {
							maxIndexDependency = dependencyIndex;
						}
					}

					if (indexSource == -1 && maxIndexDependency == -1) {
						this.orderedSources.push(source);
						cleanRun = false;
					} else {
						if (indexSource == -1) {
							this.orderedSources.splice(maxIndexDependency + 1, 0, source);
							cleanRun = false;
						} else if (indexSource < dependencyIndex) {
							this.orderedSources.splice(indexSource, 1);
							this.orderedSources.splice(maxIndexDependency, 0, source);
							cleanRun = false;
						}
					}
				} else if (indexSource == -1) {

					this.orderedSources.push(source);
					cleanRun = false;
				}
			}

			if (cleanRun) {
				break;
			}

			if (i == this.sources.length - 1 && !cleanRun) {
				throw exception('circular dependencies');
			}
		}
	},

	_load: function(i) {
		var self = this;

		$.ajax({
			crossDomain: true,
			dataType: 'script',
			url: self.orderedSources[i],
			cache: true,
			async: true,
			success: function() {
				if (i == self.orderedSources.length - 1) {
					self.callback();
				} else {
					self._load(i + 1);
				}
			},
			error: function() {
				console.log('failed to load script: ' + this.url);
			}
		});
	}
};

function getFile(source, credentials, callback) {
	var xmlHTTP = new XMLHttpRequest();

	xmlHTTP.onreadystatechange = function() {
		if (xmlHTTP.readyState == 4) {
			// check for zero for localhost loads
			if (xmlHTTP.status == 200 || xmlHTTP.status == 0) {
				callback(xmlHTTP.responseText);
			}
		}
	}

	xmlHTTP.open('get', source, true);
	if (credentials) {
		xmlHTTP.setRequestHeader('Authorization', 'Basic ' + btoa(credentials));
	}
	xmlHTTP.send();
}

/* pre-load thirdparty scripts, then call main() */
(function() {
	var loadRemainingCount = 0;

	// manual flag for data
	getFile(concatURLs(window.app.rootPath, window.app.relPath, window.app.dataSource), null, function(response) {

		window.app.data = JSON.parse(response);

		window.app.rootPath = window.app.data.build.rootPath;
		window.app.compiled = window.app.data.build.compiled;

		if (window.app.compiled) {
			signal();
			return;
		}

		var sources = window.app.data['thirdpartySources'];

		loadRemainingCount = sources.length;
		for (var i = 0; i < sources.length; ++i) {
			var script = document.createElement('script');
			script.src = sources[i];
			script.async = false;

			if (script.readyState) {
				script.onreadystatechange = function() {
					if (script.readyState == 'loaded' || script.readyState == 'complete') {
						script.onreadystatechange = null;
						signal();
					}
				};
			} else {
				script.onload = function() {
					signal();
				};
			}

			document.head.appendChild(script);
		}
	});

	function signal() {
		if (--loadRemainingCount <= 0) {
			window.app.initialize();
		}
	}
})();