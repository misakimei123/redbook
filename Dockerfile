FROM ubuntu:22.04
COPY redbook /app/redbook
WORKDIR /app
CMD ["/app/redbook"]