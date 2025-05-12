FROM 192.168.0.114:37071/zenith/alpine:latest
LABEL maintainer="zenit-yfzx-jsyfb"
# 新建目录
RUN mkdir -p /data/work/
COPY engine/dist/* /data/work/
WORKDIR /data/work
EXPOSE 8080
CMD ["/data/work/rule-server", "-g", "daemon off;"]
