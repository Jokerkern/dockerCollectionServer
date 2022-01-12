FROM ubuntu
MAINTAINER kern:852946650@qq.com
ADD collectionServer /collectionServer/collectionServer
ADD sources.list /etc/apt/sources.list
WORKDIR /collectionServer
RUN apt update
RUN apt install mariadb-server-10.3 -y
EXPOSE 9089
EXPOSE 8888
ADD start.sh /start.sh
RUN chmod +x /start.sh
ENTRYPOINT ["/start.sh"]
