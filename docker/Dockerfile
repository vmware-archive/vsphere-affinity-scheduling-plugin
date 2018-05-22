FROM ubuntu
MAINTAINER Yongkun Anfernee Gui<agui@vmware.com>

RUN apt update && apt -y install ca-certificates

ADD vsphere-affinity-scheduling-plugin /usr/bin/

EXPOSE 12346
ENTRYPOINT ["/usr/bin/vsphere-affinity-scheduling-plugin"]

