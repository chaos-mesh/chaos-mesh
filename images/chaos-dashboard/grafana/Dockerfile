FROM grafana/grafana:6.5.2

# enable anonymous
COPY grafana.ini /etc/grafana

# datasource
COPY event.yaml /etc/grafana/provisioning/datasources/
COPY metric.yaml /etc/grafana/provisioning/datasources/

# dashboard
RUN mkdir /home/grafana/templates
RUN mkdir /home/grafana/dashboards
COPY monitor.json /home/grafana/templates/
COPY monitor.yaml /etc/grafana/provisioning/dashboards/

# bootstrap script

USER root
COPY bootstrap.sh /
RUN chmod 755 /bootstrap.sh
USER grafana

ENTRYPOINT [ "/bootstrap.sh" ]