window.app.views.Nodes = Backbone.View.extend({
	el: '<div class="fill"></div>',

	initialize: function(options) {
		this.template = window.app.loadTemplate('node/nodes', window.app.data);
		this.$el.append(this.template);

		this.$parent = options.$parent;

		this.views = {};
		this.currentView = {};

		this.nodes = window.app.localData.collections.nodes;

		// base class configuration
		this.viewType = 'Node';
		this.modelType = 'Node';

		this.render();
	},

	render: function() {
		var self = this;

		this.$parent.append(this.$el);

		/* layout */
		this.layout = this.$el.layout($.extend({
			west__onresize_end: function() {},
			center__onresize_end: function() {}
		}, window.app.layoutDefaults.objectView));

		/* list view */
		this.views.list = new window.app.views.NodeList({
			$parent: this.$('.view-pane'),
			collection: this.nodes
		});
		this.views.list.on('selected', function(nodeID) {
			self.views.nodeTree.select(nodeID);
		});

		/* group tree */
		this.views.nodeTree = new window.app.views.Tree({
			$container: this.$('.node-tree-container'),
			collection: this.nodes
		});
		this.views.nodeTree.on('selected', function(nodeID) {
			self.views.list.select(nodeID);
		});

		this.currentView = this.views.list;
	},

	resize: function() {
		this.layout.resizeAll();
	}
});