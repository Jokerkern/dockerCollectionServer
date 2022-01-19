#!/bin/bash
sudo docker build -t collectionserver:v1 . 
sudo docker run -it --privileged=true --mount type=bind,src=$(cd `dirname $0`; pwd)/data,dst=/var/lib/mysql -p 9091:9089 -p 8890:8888 collectionserver:v1
