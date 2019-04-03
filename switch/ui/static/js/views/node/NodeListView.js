window.app.views.NodeList = Backbone.View.extend({
	el: '<div class="fill"></div>',
	events: {
		'click .send-message': 'sendMessage',
		'click .force-location-update': 'forceLocationUpdate'
	},

	initialize: function(options) {
		var self = this;

		this.template = window.app.loadTemplate('node/node-list', window.app.data);
		this.$el.append(this.template);

		this.views = {};

		this.$parent = options.$parent;
		this.nodes = options.collection;

		this.render();
	},

	loadMap: function() {
		var self = this;
		var map = self.views.map;

		map.clear();

		_.each(self.nodes.models, function(node) {

			var obj = node;

			if (!map.has(obj.id)) {

				map.setMapObject(map.createMapObject({
					id: obj.id,
					callback: function() {
						self.trigger('selected', obj.id);
						self.views.map.highlite(obj.id);
						self.views.nodeSummary = new window.app.views.NodeSummary({
							$parent: self.$('.node-summary-container').empty(),
							model: obj
						});
					}
				}));

				_.each(obj.get('GeoAffinities'), function(polygon) {
					map.addToObj(obj.id, map.createPolygon(_.uniqueId(), polygon));
				});
			}
		});
	},

	select: function(nodeID) {

		this.views.map.highliteAndCenter(nodeID);

		this.views.nodeSummary = new window.app.views.NodeSummary({
			$parent: this.$('.node-summary-container').empty(),
			model: this.nodes.get(nodeID)
		});
	},

	render: function() {
		var self = this;

		this.$parent.append(this.$el);

		/* layout */
		this.layout = this.$el.layout({
			north__size: .60,
			north__onresize_end: function() {
				self.views.map.resize();
			},
			center__onresize_end: function() {}
		});

		/* map view */
		this.views.map = new window.app.views.Map({
			$container: this.$('.map-container'),
			disableCompactor: true
		});

		/* node summary view */
		this.views.nodeSummary = new window.app.views.NodeSummary({
			$parent: this.$('.node-summary-container'),
			model: this.nodes.get(0)
		});

		this.nodes.on('loaded', function() {
			self.loadMap();
		});
		if (this.nodes.loaded) {
			self.loadMap();
		}

		this.resize();
	},

	resize: function() {
		this.layout.resizeAll();
	}
});