window.app.views.About = Backbone.View.extend({
	el: '<div class="fill"></div>',
	events: {},

	initialize: function(options) {
		var self = this;
		this.template = window.app.loadTemplate('user/about');
		this.views = {};

		this.$parent = options.$parent;

		this.$parent.append(this.$el);
		this.render();
	},

	render: function() {
		var self = this;

		this.$el.append(this.template($.extend({}, window.app.data)));
	}
});