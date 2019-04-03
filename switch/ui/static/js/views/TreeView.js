window.app.views.Tree = Backbone.View.extend({
	el: '<div class="fill"></div>',
	events: {
		'click .nothing': 'nothing'
	},

	initialize: function(options) {
		var self = this;

		this.template = window.app.loadTemplate('tree', window.app.data);
		this.$el.append(this.template);

		// execute throttle for collection changes
		this.updateThrottle = _.debounce(function() {

			var plugins = ['types'];
			if (self.edit) {
				plugins = _.union(plugins, ['dnd']);
			}

			var nodes = [];
			self.collection.each(function(model) {
				var data = model.attributes;

				var node = {
					id: ref(data, self.dataNames.id),
					parent: ref(data, self.dataNames.parentID) || '#',
					text: ref(data, self.dataNames.name),
					type: model.attributes.Entity.Kind,
					state: {},
					li_attr: {}
				};

				// auto-open and flag the selected node
				if (node.parent == '#' || self.selected && node.id == self.selected.id) {
					node.state.opened = true;
					node.state.selected = true;

					node.li_attr.loaded = true;
				}

				nodes.push(node);
			});

			self.$tree = self.$('.tree-container').jstree({
				core: {
					animation: 0,
					data: nodes
				},
				plugins: plugins,
				types: {
					master: {
						icon: 'icon-crown'
					},
					slave: {
						icon: 'icon-pawn'
					}
				}
			});

			self.$tree.bind('select_node.jstree', function(e, data) {
				self.trigger('selected', data.node.id);
			});

			//self.refresh();
			self.trigger('selected', nodes[0].id);
		}, 100);

		this.views = {};

		this.$container = options.$container;
		this.collection = options.collection;
		this.dataNames = {
			id: 'Entity.id',
			parentID: 'Parent.id',
			name: 'Entity.id'
		};

		this.selected = null;

		this.collection.on('loading', function() {});
		this.collection.on('sync add remove', function() {
			self.update();
		});

		this.render();
	},

	update: function() {
		this.updateThrottle();
	},

	getSelectedIDs: function() {
		return this.$tree.jstree('get_selected');
	},

	select: function(id) {
		var self = this;

		_.each(this.getSelectedIDs(), function(id) {
			self.$tree.jstree('deselect_node', '#' + id);
		});

		this.$tree.jstree('select_node', '#' + id);
	},

	render: function() {
		var self = this;

		this.$container.append(this.$el);

		if (this.collection.loaded) {
			this.update();
		}
	},

	refresh: function() {
		this.$tree.jstree('refresh');
	}
});