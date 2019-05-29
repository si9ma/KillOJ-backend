FROM 1.12-alpine

RUN apk add \
    # for docker-entrypoint.sh
    bash \
    # for go gete
    git

RUN mkdir /app

ADD ./docker-entrypoint.sh /

RUN go get -v github.com/si9ma/KillOJ-backend
RUN go build -o /app/backend github.com/si9ma/KillOJ-backend 

RUN cp -r $GOPATH/src/github.com/si9ma/KillOJ-backend/conf /app

WORKDIR /app

EXPOSE 8080
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["backend"]