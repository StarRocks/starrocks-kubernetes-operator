FROM ubuntu:22.04

ADD /bin/sroperator /sroperator

CMD ["/sroperator"]
