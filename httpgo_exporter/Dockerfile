FROM veknet/httpgo
  
USER root

COPY ./process_exporter /bin/process_exporter
COPY ./process_monitor_new.sh bin/process_monitor_new.sh
COPY ./entrypoint.sh bin/entrypoint.sh

RUN chmod 700 bin/process_monitor_new.sh
RUN chmod 700 bin/entrypoint.sh

CMD bin/entrypoint.sh
