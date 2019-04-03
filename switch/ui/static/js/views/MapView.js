window.app.views.Map = Backbone.View.extend({
	el: '<div class="fill"></div>',
	events: {
		'click .zoom-to': 'zoomTo',
		'click .fit-all': 'fitAll'
	},

	initialize: function(options) {
		var self = this;

		this.template = window.app.loadTemplate('map', window.app.data);
		this.$el.append(this.template);

		this.postProcessThrottle = _.debounce(function(callback) {

			// if there's no map bounds, don't do anything
			if (!this.map.getBounds()) {
				return;
			}

			self.bounds = new google.maps.LatLngBounds();
			self.iterateObjects(function(obj) {
				self.addToBounds(obj.mappable);
			});

			if (callback) {
				callback();
			}

			if (!self.disableCompactor) {
				self.compactObjectChildren();
			}
			if (!self.disableClusterer) {
				self.clusterManager.render(self.objects);
			}

			if (self.selectedObject) {
				self.highlite(self.selectedObject);
			}

			this.trigger('refresh');

		}, 50);

		this.fitBoundsThrottle = _.debounce(function(bounds) {
			if (!bounds) {
				bounds = this.bounds;
			}

			if (!bounds.isEmpty()) {
				var NE = bounds.getNorthEast();
				var SW = bounds.getSouthWest();

				// don't zoom in too far on only one marker
				if (NE.equals(SW)) {
					bounds.extend(new google.maps.LatLng(NE.lat() + 0.001, NE.lng() + 0.001));
					bounds.extend(new google.maps.LatLng(NE.lat() - 0.001, NE.lng() - 0.001));
				}

				this.map.fitBounds(bounds);
			}
		}, 50, true);

		this.drawingManager = null;

		this.$container = options.$container;
		this.mapParams = $.extend({
			center: {
				lat: 30,
				lng: 0
			},
			zoom: 2,
			minZoom: 2,
			zoomControl: true,
			zoomControlOptions: {
				position: google.maps.ControlPosition.LEFT_BOTTOM
			},
			mapTypeControl: true,
			mapTypeControlOptions: {
				position: google.maps.ControlPosition.LEFT_TOP
			},
			scaleControl: true,
			streetViewControl: true,
			streetViewControlOptions: {
				position: google.maps.ControlPosition.LEFT_BOTTOM
			},
			rotateControl: true,
			rotateControlOptions: {
				position: google.maps.ControlPosition.LEFT_BOTTOM
			}
		}, window.app.data.session.map, (options.mapSettings || {}));

		this.disableCompactor = !!options.disableCompactor;

		this.objects = {};
		this.tmpObject = null;
		this.selectedObject = null;

		this.clusterKing = new gregzuro.mapping.MapObject({
			map: this.map
		});
		this.setMapObject(this.clusterKing);

		this.bounds = new google.maps.LatLngBounds();

		this.nextZ = 333;

		this.clear();

		this.render();
	},

	postProcess: function(callback) {
		this.postProcessThrottle();

		if (callback) {
			this.once('refresh', function() {
				callback();
			});
		}
	},

	render: function() {
		var self = this;

		this.$container.append(this.$el);

		this.map = new google.maps.Map(this.$('.map')[0], this.mapParams);

		// for clustering multiple objects in same spot
		this.clusterManager = new gregzuro.mapping.ClusterManager(this.map, this.clusterKing);

		google.maps.event.addDomListener(this.map, 'zoom_changed', function() {
			google.maps.event.addListenerOnce(self.map, 'idle', function() {
				self.postProcess();
			});
		});
	},

	// todo bug on zoom to (bounds on object broken too)
	zoomTo: function() {
		if (this.selectedObject) {
			this.map.setCenter(this.selectedObject.getBounds().getCenter());

			// bounds function is not a super close zoom, so don't zoom out
			if (this.map.getZoom() < 15) {
				this.fitBounds(this.selectedObject.getBounds());
			}
		}
	},

	fitAll: function() {
		this.fitBounds();
	},

	compactObjectChildren: function() {
		var self = this;

		var mapBounds = this.map.getBounds();
		var mapSpan = mapBounds.getNorthEast().lat() - mapBounds.getSouthWest().lat();

		_.each(this.objects, function(topLevelObj) {

			// only do this if the top level object has more than 1 child, or if only 1, just use the defined zoom
			//if (_.size(topLevelObj.children) < 2 && this.map.getZoom() > 10) { return; }

			if (topLevelObj.isMarkerOnly()) {
				return;
			}

			var bounds = topLevelObj.getBounds();
			var span = bounds.getNorthEast().lat() - bounds.getSouthWest().lat();

			// hide that children and add a placeholder (if not already collapsed)
			if (span / mapSpan < .05) {
				if (!topLevelObj.collapsed) {
					topLevelObj.collapse();
				}
			} else if (topLevelObj.collapsed) {
				topLevelObj.expand();
			}
		});
	},

	centerMap: function(obj) {

		var bounds = obj.getBounds();

		if (obj.hasChildren) {
			if (!this.map.getBounds().contains(bounds.getCenter())) {

				this.map.setCenter(bounds.getCenter());

				if (!this.map.getBounds().contains(bounds.getNorthEast()) &&
					!this.map.getBounds().contains(bounds.getSouthWest())) {

					this.fitBounds(bounds);
				}
			}
		} else {
			if (!this.map.getBounds().contains(bounds.getCenter())) {
				this.map.setCenter(bounds.getCenter());
			}
		}
	},

	has: function(id) {
		return _.has(this.objects, id);
	},

	setTmpObject: function(idOrMapObj) {

		var mapObj = this.translateInput(idOrMapObj);

		if (this.tmpObject) {
			// check if it's a top level object
			if (this.has(this.tmpObject.id)) {
				this.removeTopLevelObject(this.tmpObject.id)
			} else {
				this.tmpObject.destroy();
			}
		}

		this.tmpObject = mapObj;
		this.tmpObject.tmp = true;
	},

	unsetTmpObject: function(idOrMapObj) {

		var mapObj = this.translateInput(idOrMapObj);

		if (this.tmpObject && this.tmpObject.id == mapObj.id) {
			mapObj.tmp = false;
			this.tmpObject = null;
		}
	},

	setPointer: function(obj) {

		var refMappable = obj.mappable;
		var position = refMappable.getPosition();

		var markerObj = this.createMarker(null, [position.lat(), position.lng()], {
			path: google.maps.SymbolPath.BACKWARD_CLOSED_ARROW,
			strokeWeight: 2,
			callback: function() {
				if (obj.mappable.callback) {
					obj.mappable.callback(obj.id);
				}
			},
			scale: 6,
			zIndex: 9999 // not working (todo)
		});
		markerObj.selectedIconOptions = {
			fillColor: refMappable.getIcon().fillColor
		};

		this.setTmpObject(markerObj);
		this.highliteAndCenter(markerObj);
	},

	highlite: function(idOrMapObj) {
		var self = this;
		var mapObj = this.translateInput(idOrMapObj);

		// kill any tmp objects if they aren't the one selected
		if (this.tmpObject && mapObj.id != this.tmpObject.id) {
			this.removeTopLevelObject(this.tmpObject.id);
		}

		// revert currently selected
		if (this.selectedObject) {

			// check if the object is in a cluster
			if (this.selectedObject.cluster) {
				this.selectedObject.cluster.restoreMappableState();
			}

			this.selectedObject.iterateDown(function(obj) {
				obj.restoreMappableState();
			});


			if (this.selectedObject.collapsedMarker) {
				this.selectedObject.collapsedMarker.setIcon(_.extend(this.selectedObject.collapsedMarker.getIcon(), {
					fillColor: '#00C1FF'
				}));
			}
		}

		// check if the object is in a cluster / manage cluster hilighting
		this.clusterManager.lowlight();
		if (mapObj.cluster) {
			this.clusterManager.hilight(mapObj.cluster);
		}

		// highlite the newly selected object
		mapObj.iterateDown(function(obj) {
			if (obj.collapsedMarker) {
				obj.collapsedMarker.setIcon(_.extend(obj.collapsedMarker.getIcon(), {
					fillColor: '#00FF00'
				}));
				obj.collapsedMarker.setZIndex(self.nextZ++);
			}

			if (obj.mappable) {
				obj.saveMappableState();

				if (obj.mappable instanceof google.maps.Marker) {
					if (obj.selectedIconOptions) {
						obj.mappable.setIcon(_.extend(obj.mappable.getIcon(), obj.selectedIconOptions));
					} else {
						obj.mappable.setIcon(_.extend(obj.mappable.getIcon(), {
							fillColor: '#00FF00'
						}));
					}
					obj.mappable.setZIndex(self.nextZ++);
				} else if (
					obj.mappable instanceof google.maps.Circle ||
					obj.mappable instanceof google.maps.Polygon) {

					obj.mappable.setOptions({
						fillOpacity: .5,
						strokeColor: '#0000FF',
						strokeOpacity: 1,
						strokeWeight: 2,
						zIndex: self.nextZ++
					});
				}
			}
		});

		this.selectedObject = mapObj;
	},

	highliteAndCenter: function(idOrMapObj) {
		this.highlite(idOrMapObj);
		this.centerMap(this.selectedObject);
	},

	/*
	 *  object rendering functions
	 */

	createMarker: function(id, latLng, options) {

		var defaults = {
			path: fontawesome.markers.MAP_MARKER,
			fillColor: '#00C1FF',
			fillOpacity: 1,
			strokeColor: '#000000',
			strokeWeight: 1,
			scale: .5,
		};

		if (!options) {
			options = {};
		}

		var mapObj = new gregzuro.mapping.MapObject({
			id: id,
			map: this.map
		});
		if (options.callback) {
			mapObj.callback = function() {
				options.callback();
			}
		}

		var marker = new google.maps.Marker({
			position: {
				lat: latLng[0],
				lng: latLng[1]
			},
			map: this.map,
			title: options.title
		});
		marker.setIcon(_.extend(defaults, options));
		marker.addListener('click', function() {
			if (mapObj.callback) {
				mapObj.callback(id);
			}
		});

		mapObj.mappable = marker;

		return mapObj;
	},

	createCircle: function(id, latLng, radius, options) {

		var defaults = {
			strokeWeight: 2,
			strokeColor: '#000000',
			fillOpacity: .01
		};

		if (!options) {
			options = {};
		}

		var mapObj = new gregzuro.mapping.MapObject({
			id: id,
			map: this.map
		});
		if (options.callback) {
			mapObj.callback = function() {
				options.callback();
			}
		}

		var fillColor;
		var strokeOpacity;
		if (options.anti) {
			fillColor = '444444';
			strokeOpacity = 0;
		} else {
			fillColor = '00C1FF';
			strokeOpacity = 1;
		}

		var circle = new google.maps.Circle(_.extend({
			map: this.map,
			center: {
				lat: latLng[0],
				lng: latLng[1]
			},
			radius: radius,
			fillColor: '#' + fillColor,
			strokeOpacity: strokeOpacity
		}, options));

		circle.addListener('click', function() {
			if (mapObj.callback) {
				mapObj.callback(id);
			}
		});

		mapObj.mappable = circle;

		return mapObj;
	},

	createPolygon: function(id, latLngs, options) {

		var defaults = {
			strokeWeight: 2,
			strokeColor: '#000000',
			fillOpacity: .01
		};

		if (!options) {
			options = {};
		}

		var mapObj = new gregzuro.mapping.MapObject({
			id: id,
			map: this.map
		});
		if (options.callback) {
			mapObj.callback = function() {
				options.callback();
			}
		}

		var latLngsClockwise = gregzuro.mapping.isPolyClockwise(latLngs);
		var reverseLatLngs = false; //!latLngsClockwise;

		var coords = [];
		if (reverseLatLngs) {
			_.each(latLngs, function(latLng) {
				coords.unshift({
					lat: latLng[0],
					lng: latLng[1]
				});
			});
		} else {
			_.each(latLngs, function(latLng) {
				coords.push({
					lat: latLng[0],
					lng: latLng[1]
				});
			});
		}

		var fillColor;
		var strokeOpacity;
		if (options.anti || !latLngsClockwise) {
			fillColor = '444444';
			strokeOpacity = 0;
		} else {
			fillColor = '00FF00';
			strokeOpacity = 1;
		}

		var polygon = new google.maps.Polygon(_.extend({
			map: this.map,
			paths: coords,
			fillColor: '#' + fillColor,
			strokeOpacity: strokeOpacity
		}, options));
		polygon.addListener('click', function() {
			if (mapObj.callback) {
				mapObj.callback(id);
			}
		});

		mapObj.mappable = polygon;

		return mapObj;
	},

	createSequentialSet: function(id, arr, callback) {
		var self = this;

		var colorIndex = 0;
		var colorInc = (arr.length > 1) ? (1 / (arr.length - 1)) : 0;

		var set = new gregzuro.mapping.MapObject({
			id: id,
			map: this.map
		});
		set.setChild(new gregzuro.mapping.MapObject({
			id: 'markers',
			map: this.map
		}));
		set.setChild(new gregzuro.mapping.MapObject({
			id: 'lines',
			map: this.map
		}));

		for (var i = 0; i < arr.length; ++i) {

			var item = arr[i];

			var color;
			if (colorIndex < .5) {
				color = rgb2hex(255, colorIndex * 2 * 255, 0);
			} else {
				color = rgb2hex(Math.max(255 - (((colorIndex * 2) - 1) * 255), 0), 255, 0);
			}

			// create the marker
			var markerObj = this.createMarker(item.id, [parseFloat(item.lat), parseFloat(item.lng)], {
				title: item.id,
				path: google.maps.SymbolPath.CIRCLE,
				fillColor: '#' + color,
				strokeWeight: 2,
				scale: 6
			});

			if (callback) {
				(function(obj) {
					obj.callback = function() {
						callback(obj);
					};
				})(markerObj);
			}

			// add marker to set
			this.addToObj(set.children.markers, markerObj);

			// draw line objects (even if not currently used)
			if (i > 0) {

				var lines = [];
				lines.push(new google.maps.LatLng(arr[i - 1].lat, arr[i - 1].lng));
				lines.push(new google.maps.LatLng(arr[i].lat, arr[i].lng));

				var line = new google.maps.Polyline({
					path: lines,
					geodesic: true,
					strokeColor: '#' + color,
					strokeOpacity: 1.0,
					strokeWeight: 2
				});

				var lineObj = new gregzuro.mapping.MapObject({
					map: this.map,
					id: item.id,
					mappable: line
				});

				// add the line to set
				this.addToObj(set.children.lines, lineObj);
			}

			// iterate to next color
			colorIndex += colorInc;
		}

		return set;
	},

	/*
	 *  map object functions
	 */

	translateInput: function(idOrMapObj) {
		var mapObj = null;

		if (idOrMapObj && idOrMapObj instanceof gregzuro.mapping.MapObject) {
			mapObj = idOrMapObj;
		} else if (idOrMapObj) {
			mapObj = this.objects[idOrMapObj];
		}

		if (!mapObj) {
			throw ('invalid map object');
		}

		return mapObj;
	},

	setMapObject: function(mapObj) {
		var self = this;

		// delete the previous one
		this.removeTopLevelObject(mapObj.id);

		this.objects[mapObj.id] = mapObj;

		// add the event listener (more work here!!!)
		mapObj.on('change', function() {
			self.postProcess(function() {
				//self.fitObjects();
			});
		});

		this.postProcess(function() {
			self.fitObjects();
		});
	},

	createMapObject: function(options) {
		return new gregzuro.mapping.MapObject(_.extend(options, {
			map: this.map
		}));
	},

	getMapObject: function() {
		var args = _.toArray(arguments);
		return this.objects[args[0]].getChild(args.slice(1));
	},

	addToTmp: function(child) {
		if (!this.tmpObject) {
			this.tmpObject = new gregzuro.mapping.MapObject({
				map: this.map
			});
		}
		this.addToObj(this.tmpObject, child);
	},

	addToObj: function(idOrMapObj, child) {
		var self = this;
		var mapObj = this.translateInput(idOrMapObj);

		mapObj.setChild(child);

		this.postProcess(function() {
			self.fitObjects();
		});
	},

	iterateObjects: function(func) {
		var self = this;
		_.each(this.objects, function(obj) {
			obj.iterateDown(func);
		});
	},

	removeTopLevelObject: function(id) {
		if (!this.has(id)) {
			return;
		}

		if (this.selectedObject && this.selectedObject.id == id) {
			this.selectedObject = null;
		}

		this.objects[id].destroy();
		delete this.objects[id];

		this.postProcess();
	},

	/*
	 *  bounds/global render functions
	 */

	addToBounds: function(mappable) {
		if (mappable instanceof google.maps.Marker) {
			this.bounds.extend(mappable.getPosition());
		} else if (mappable instanceof google.maps.Circle) {
			this.bounds.union(mappable.getBounds());
		} else if (mappable instanceof google.maps.Polygon) {
			this.bounds.union(mappable.getBounds());
		}
	},

	fitBounds: function(bounds) {
		this.fitBoundsThrottle(bounds);
	},

	fitObjects: function() {
		this.fitBounds();
	},

	clear: function() {
		this.iterateObjects(function(obj) {
			obj.destroy();
		});
		this.objects = {};

		if (this.tmpObject && this.tmpObject.mappable) {
			this.tmpObject.destroy();
		}
	},

	resize: function() {
		google.maps.event.trigger(this.map, 'resize');
	}
});


// distinct object for gregzuro mapping features
var gregzuro = new function() {
	this.mapping = new function() {

		this.isPolyClockwise = function(latLngs) {
			var count = 0;

			for (var i = 0; i < latLngs.length; ++i) {
				var nextIndex = (i + 1) % latLngs.length;
				count += (latLngs[nextIndex][0] - latLngs[i][0]) * (latLngs[nextIndex][1] + latLngs[nextIndex][1]);
			}

			return count > 0;
		};

		this.ClusterManager = function(map, clusterObj) {
				this.map = map;
				this.clusterObj = clusterObj;
				this.hilightedCluster = null;
			},

			this.ClusterManager.prototype = {

				hilight: function(obj) {
					if (obj.mappable instanceof google.maps.Marker) {
						obj.saveMappableState();
						obj.mappable.setIcon(obj.selectedIconOptions);
					}
				},

				lowlight: function() {
					if (this.hilightedCluster) {
						this.hilightedCluster.restoreMappableState();
					}
				},

				render: function(objs, selectedID) {
					var self = this;

					// reset clusters
					this.clusterObj.destroyChildren();

					// reset toplevel
					_.each(objs, function(obj) {
						// don't mess with this obj
						if (obj.id == self.clusterObj.id) {
							return;
						}

						obj.cluster = null;

						// don't mess with object state
						obj.visibleOn();
					});

					// don't cluster when zoomed in
					if (this.map.getZoom() > 15) {
						return;
					}

					var mapBounds = this.map.getBounds();

					var mapSpan = mapBounds.getNorthEast().lat() - mapBounds.getSouthWest().lat();
					var TR = mapBounds.getNorthEast();
					var BL = mapBounds.getSouthWest();

					// group items within 5% screen width of eachother
					var threshold = distanceInMeters(TR.lat(), TR.lng(), BL.lat(), BL.lng()) / 20;

					var groups = [];
					var used = {};

					for (var name1 in objs) {

						// don't add this object
						if (name1 == this.clusterObj.id) {
							continue;
						}

						// don't cluster objects that have children and aren't collapsed
						if (objs[name1].hasChildren() && !objs[name1].collapsed) {
							continue;
						}

						if (used[name1]) {
							continue;
						}

						var group = [];
						used[name1] = true;
						group.push(name1);

						for (var name2 in objs) {

							// don't add this object
							if (name2 == this.clusterObj.id) {
								continue;
							}

							// don't cluster objects that have children and aren't collapsed
							if (objs[name2].hasChildren() && !objs[name2].collapsed) {
								continue;
							}

							if (used[name2]) {
								continue;
							}

							var center1 = objs[name1].getCenter();
							var center2 = objs[name2].getCenter();

							var distance = distanceInMeters(center1.lat(), center1.lng(), center2.lat(), center2.lng());
							if (distance < threshold) {
								group.push(name2);
								used[name2] = true;
							}
						}
						groups.push(group);
					}

					var maxSize = _.size(objs);

					_.each(groups, function(group) {

						if (group.length > 1) {
							var imgID = Math.min(Math.ceil(group.length / maxSize * 3), 3);
							var halfSize = (imgID == 3) ? 39 : (imgID == 2) ? 33 : 26;

							var offsetX = 1 + (4 * ('' + group.length).length);
							var offsetY = 10;

							var image = {
								url: concatURLs(window.app.rootPath, 'static/imgs/map/m' + imgID + '.png'),
								origin: new google.maps.Point(0, 0),
								anchor: new google.maps.Point(halfSize, halfSize)
							};

							var marker = new MarkerWithLabel({
								position: objs[group[0]].getCenter(),
								map: self.map,
								icon: image,
								labelContent: group.length,
								labelAnchor: new google.maps.Point(offsetX, offsetY),
								labelClass: 'clusterIcon'
							});
							marker.addListener('click', function() {
								if (objs[group[0]].callback) {
									objs[group[0]].callback();
								}
							});

							var cluster = new gregzuro.mapping.MapObject({
								map: self.map,
								mappable: marker,
								selectedIconOptions: {
									url: concatURLs(window.app.rootPath, 'static/imgs/map/ms.png'),
									origin: new google.maps.Point(0, 0),
									anchor: new google.maps.Point(39, 39)
								}
							});

							_.each(group, function(id) {
								objs[id].cluster = cluster;

								// don't flat as hidden, this is a stateless final operation
								objs[id].visibleOff();
							});

							self.clusterObj.setChild(cluster);
						} else {
							objs[group[0]].cluster = null;
							objs[group[0]].visibleOn();
						}
					});
				}
			},

			this.MapObject = function(options) {
				var self = this;

				this.children = {};
				this.parent = null;

				this.collapsed = false;
				this.cluster = null;

				// special marker for when collapsed
				var marker = new google.maps.Marker();
				marker.setIcon({
					path: fontawesome.markers.MAP_MARKER,
					fillColor: '#00C1FF',
					fillOpacity: 1,
					strokeWeight: 2,
					scale: .5
				});
				marker.addListener('click', function() {
					if (self.callback) {
						self.callback(self.id);
					}
				});
				this.collapsedMarker = marker;

				if (options) {
					this.map = options.map;
					this.id = options.id || _.uniqueId();
					this.callback = options.callback;
					this.mappable = options.mappable;
					this.hidden = !!options.hidden;
					this.selectedIconOptions = options.selectedIconOptions;
				}

				// add an event handler
				_.extend(this, Backbone.Events);
			};

		this.MapObject.prototype = {
			iterateDown: function(func) {
				var recurIter = function(obj, func) {

					if (func(obj) === false) {
						return false;
					}

					if (obj.children) {
						_.each(obj.children, function(child) {
							recurIter(child, func);
						});
					}
				};

				recurIter(this, func);
			},

			iterateUp: function(func) {
				var recurIter = function(obj, func) {
					if (func(obj) === false) {
						return false;
					}
					if (obj.parent) {
						recurIter(obj.parent, func);
					}
				};

				recurIter(this, func);
			},

			iterateChildren: function(func) {
				_.each(this.children, function(child) {
					child.iterateDown(func);
				});
			},

			iterateParents: function(func) {
				if (this.parent != null) {
					this.parent.iterateUp(func);
				}
			},

			isMarkerOnly: function() {
				var markerOnly = true;
				this.iterateDown(function(obj) {
					if (obj.mappable) {
						if (!(obj.mappable instanceof google.maps.Marker)) {
							markerOnly = false;
							return false;
						}
					}
				});

				return markerOnly;
			},

			destroyChildren: function() {
				_.each(this.children, function(child) {
					child.destroy();
				});
			},

			destroy: function() {
				this.iterateDown(function(obj) {
					if (obj.mappable) {
						obj.mappable.setMap(null);
						obj.mappable = null;
					}

					if (obj.collapsedMarker) {
						obj.collapsedMarker.setMap(null);
						obj.collapsedMarker = null;
					}

					if (obj.parent) {
						delete obj.parent.children[obj.id];
					}

					// unbind all events
					obj.off();
				});

				this.id = null;
				this.children = {};
				this.callback = null;
			},

			getBounds: function() {
				var bounds = new google.maps.LatLngBounds();

				this.iterateDown(function(obj) {
					if (obj.mappable) {
						var mappable = obj.mappable;

						if (mappable instanceof google.maps.Marker) {
							bounds.extend(mappable.getPosition());
						} else if (mappable instanceof google.maps.Circle) {
							bounds.union(mappable.getBounds());
						} else if (mappable instanceof google.maps.Polygon) {
							bounds.union(mappable.getBounds());
						}
					}
				});

				if (bounds.isEmpty()) {
					bounds = null;
				}

				return bounds;
			},

			getCenter: function() {
				return this.getBounds().getCenter();
			},

			hasChildren: function() {
				return _.size(this.children);
			},

			setChild: function(child) {
				var self = this;

				if (this.children[child.id]) {
					this.children[child.id].destroy();
				}

				if (this.callback && !child.callback) {
					child.callback = function() {
						self.callback.apply(_.toArray(arguments));
					}
				}

				child.parent = this;
				this.children[child.id] = child;

				child.on('change', function() {
					self.trigger('change');
				});
			},

			getChild: function() {
				var args = _.toArray(arguments);

				var child = this;
				_.each(args, function(arg) {
					child = child.children[arg];
				});

				return child;
			},

			visibleOn: function() {
				// check up, if parent or parent of ... is hidden or collapsed
				var parentNotShown = false;
				this.iterateParents(function(obj) {
					if (obj.hidden || obj.collapsed) {
						parentNotShown = true;
						return false;
					}
				});

				if (!parentNotShown) {
					// check down
					this.iterateDown(function(obj) {
						if (!obj.hidden) {
							if (obj.collapsed) {
								obj.setMarkerMap(obj.collapsedMarker, obj.map);
								return false;
							} else if (obj.mappable) {
								obj.setMarkerMap(obj.mappable, obj.map);
							}
						} else {
							return false;
						}
					});
				}
			},

			visibleOff: function() {
				this.iterateDown(function(obj) {
					if (obj.mappable) {
						obj.setMarkerMap(obj.mappable, null);
					}
					if (obj.collapsedMarker) {
						obj.setMarkerMap(obj.collapsedMarker, null);
					}
				});
			},

			show: function() {
				this.hidden = false;
				this.visibleOn();

				this.trigger('change');
			},

			hide: function() {
				this.hidden = true;
				this.visibleOff();

				this.trigger('change');
			},

			expand: function() {
				var self = this;
				this.collapsed = false;

				this.visibleOn();

				this.setMarkerMap(this.collapsedMarker, null);
			},

			collapse: function() {
				var self = this;
				this.collapsed = true;


				this.visibleOff();

				this.collapsedMarker.setPosition(this.getBounds().getCenter());
				this.setMarkerMap(this.collapsedMarker, this.map);
			},

			setMarkerMap: function(marker, map) {
				if (marker.map != map) {
					marker.setMap(map);
				}
			},

			saveMappableState: function() {
				var mappable = this.mappable;

				if (mappable) {
					if (mappable instanceof google.maps.Marker) {
						var icon = this.mappable.getIcon();

						options = {};

						for (var name in icon) {
							options[name] = icon[name];
						}
					} else {
						options = {
							fillColor: this.mappable.get('fillColor'),
							fillOpacity: this.mappable.get('fillOpacity'),
							strokeColor: this.mappable.get('strokeColor'),
							strokeOpacity: this.mappable.get('strokeOpacity'),
							strokePosition: this.mappable.get('strokePosition'),
							strokeWeight: this.mappable.get('strokeWeight')
						};
					}

					this.savedMappableOptions = options;
				}
			},

			restoreMappableState: function() {
				var mappable = this.mappable;

				if (this.savedMappableOptions && mappable) {
					if (mappable instanceof google.maps.Marker) {
						mappable.setIcon(_.extend(mappable.getIcon(), this.savedMappableOptions));
					} else {
						mappable.setOptions(this.savedMappableOptions);
					}
				}
			}
		};
	};
};