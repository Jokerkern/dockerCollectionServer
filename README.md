# dockerCollectionServer
sudo docker build -t collectionserver:v1 .  
sudo docker run -it --privileged=true --mount type=bind,src=???/data,dst=/var/lib/mysql -p 9091:9089 -p 8890:8888 collectionserver:v1
