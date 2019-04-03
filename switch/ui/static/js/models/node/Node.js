window.app.models.Node = Backbone.Model.extend({
	defaults: function() {
		return {
			Entity: {},
			Parent: {},
			Master: {},
		};
	},

	initialize: function(attrs, options) {
		var self = this;
		this.loaded = false;
	},

	parse: function(response, options) {
		var self = this;

		this.set('id', response.Entity.id);

		response.GeoAffinities = [];
		
		_.each(response.Entity.geo_affinities, function(geo) {
			var polygon = [];

			_.each(geo.geo_fence.polygon.points, function(coord) {
				polygon.push([coord.latitude, coord.longitude]);
			});

			var isClockwise = gregzuro.mapping.isPolyClockwise(polygon);
			if (geo.exclude && isClockwise || !geo.exclude && !isClockwise) {
				polygon = polygon.reverse();
			}

			response.GeoAffinities.push(polygon);
		});
		
		return response;
	},

	validate: function(attrs) {
		return false;
	},

	sync: function(method, model, options) {
		var self = this;

		switch (method) {
			case 'create':
				break;
			case 'update':
				break;
			case 'read':
				break;
			case 'destroy':
				break;
		}
	}
});