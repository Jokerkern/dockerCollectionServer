#!/bin/bash
chown -R mysql /var/lib/mysql
chgrp -R mysql /var/lib/mysql
service mysql start
/collectionServer/collectionServer -m=1000000
