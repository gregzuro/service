window.app.views.NodeSummary = Backbone.View.extend({
	el: '<div class="fill"></div>',
	events: {
		'click .nothing': 'nothing',
	},

	initialize: function(options) {
		var self = this;

		this.template = window.app.loadTemplate('node/node-summary');

		this.views = {};

		this.$parent = options.$parent;
		this.model = options.model || new window.app.models.Node();

		this.$parent.append(this.$el);
		this.render();
	},

	render: function() {
		var self = this;

		if (this.model) {
			this.$el.append(this.template($.extend({}, window.app.data, this.model.toJSON())));
		} else {
			this.$el.append(this.template($.extend({}, window.app.data)));
		}

		this.bindDataRefs(this.model);
	}
});