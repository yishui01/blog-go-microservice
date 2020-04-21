#!/bin/sh

#开启mysql远程连接
USERNAME=root
MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD
MYSQL_PORT=$MYSQL_PORT

remote_connect="grant all privileges on *.*  to 'root'@'%'"
update_root_pass="ALTER USER 'root'@'%' IDENTIFIED WITH mysql_native_password BY '${MYSQL_ROOT_PASSWORD}'"
flush_pri="flush privileges"

/usr/bin/mysql -P${MYSQL_PORT} -u${USERNAME} -p${MYSQL_ROOT_PASSWORD} -e "${remote_connect}"
/usr/bin/mysql -P${MYSQL_PORT} -u${USERNAME} -p${MYSQL_ROOT_PASSWORD} -e "${update_root_pass}"
/usr/bin/mysql -P${MYSQL_PORT} -u${USERNAME} -p${MYSQL_ROOT_PASSWORD} -e "${flush_pri}"