FROM debian

WORKDIR /app
COPY echoservice /app/echoservice

CMD ["/app/echoservice"]