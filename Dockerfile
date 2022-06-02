FROM centos:7

ADD /bin/manager /manager

CMD ["/manager"]
