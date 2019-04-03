window.app.views.App = Backbone.View.extend({
	el: '<div class="fill"></div>',
	events: {
		'click .nothing': 'nothing'
	},

	initialize: function() {
		var self = this;

		this.template = window.app.loadTemplate('app', window.app.data);
		this.$el.append(this.template);

		this.resize = _.debounce(function() {
			if (self.pageView) {
				self.pageView.resize();
			}
		}, 50, false);

		$(window).on('resize', function() {
			self.resize();
		});

		$('body').prepend(this.$el);

		this.render();
	},

	route: function(page, path) {
		var self = this;

		if (this.pageView) {
			this.pageView.hide();
		}

		// make sure the content container is empty and the app bar cleared
		$('#content').empty();
		$('#app-bar').empty();

		if (this.pages[page]) {
			this.pageView = this.pages[page];
		} else {
			var view = window.app.getView(page);
			if (view) {
				this.pages[page] = new view({
					$parent: $('#content'),
					path: path
				});
				this.pageView = this.pages[page];
			} else {
				this.pageView = new window.app.views.Error({
					$parent: $('#content'),
					errorMessage: '404: Page not found.'
				});
			}
		}

		this.pageView.show();

		if (path) {
			var matches = path.match(/([^\/]*)\/?(.*)/);
			if (page) {
				this.pageView.route(matches[1], matches[2]);
			}
		}

		// set nav for page
		this.setNavs(page);
	},

	setNavs: function(page) {
		$('#app-navbar .active').removeClass('active');
		var $activeOption = this.$('#app-navbar a').filter(function() {
			return $(this).attr('href').match(new RegExp('#App\\/' + page));
		});

		// check if top level option or drowndown option
		if ($activeOption.closest('ul').hasClass('dropdown-menu')) {
			$activeOption.parent().parents('li').addClass('active');
			$activeOption.addClass('active');
		} else {
			$activeOption.parent().addClass('active');
		}
	},

	render: function() {
		var self = this;

		this.pages = {};
		this.pageView = null;

		$('title').html(window.app.title());

		window.app.localData.collections.nodes = new window.app.collections.Nodes();

		// iterate thru ini load and local collections
		_.each(window.app.localData.collections, function(collection) {
			collection.fetch();
		});
	}
});