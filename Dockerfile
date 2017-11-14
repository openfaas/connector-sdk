FROM alpine:3.6
WORKDIR /root/
COPY connector    connector

CMD ["./connector"]