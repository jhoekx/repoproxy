FROM golang:1 as builder

WORKDIR /go/src/github.com/jhoekx/repoproxy
COPY . .

RUN go install -v ./...

FROM centos:7
COPY --from=builder /go/bin/repoproxy /usr/bin/repoproxy

ENV CENTOS_MIRROR "http://centos.mirror.nucleus.be"
ENV RPM_DIR /var/lib/repoproxy/rpms
VOLUME ${RPM_DIR}
EXPOSE 8080

CMD ["repoproxy"]
