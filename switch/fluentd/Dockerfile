FROM fluent/fluentd:latest-onbuild
RUN fluent-gem install fluent-plugin-influxdb
EXPOSE 24284
CMD fluentd -c /fluentd/etc/$FLUENTD_CONF -p /fluentd/plugins $FLUENTD_OPT