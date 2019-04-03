window.app.views.Settings = Backbone.View.extend({
	el: '<div class="fill"></div>',
	events: {
		'click #reset': 'reset'
	},

	initialize: function(options) {
		var self = this;

		this.template = window.app.loadTemplate('user/settings');

		this.views = {};

		this.$parent = options.$parent;

		this.$parent.append(this.$el);
		this.render();
	},

	reset: function() {

		var session = window.app.data.session;

		this.$timezoneSelector.val(session.locale.timezone).trigger('change');
		this.$datetimeSelector.val(session.locale.datetimeFormat).trigger('change');
	},

	render: function() {
		var self = this;

		var datetimeFormats = [
			'MM/DD/YY hh:mm:ss A',
			'DD/MM/YY HH:mm:ss',
			'YYYY-MM-DD hh:mm:ss A',
			'YYYY-MM-DD HH:mm:ss'
		];

		this.$el.append(this.template($.extend({}, window.app.data)));


		this.$timezoneSelector = this.$('#timezone-dropdown').select2({
			data: keyValueToArray(window.app.timezones, 'id', 'text'),
			minimumResultsForSearch: -1,
			allowClear: false
		});

		this.$datetimeSelector = this.$('#datetime-format-dropdown').select2({
			data: _.map(datetimeFormats, function(val) {
				return {
					id: val,
					text: val
				};
			}),
			minimumResultsForSearch: -1,
			allowClear: false
		});
		this.$datetimeSelector.on('change', function() {
			self.$('#datetime-example').html(window.app.momentInLocaleTZ(moment()).format($(this).val()));
		});

		this.reset();
	}
});