#!/bin/sh

#创建数据库并导入表结构
CONTAINER_NAME="mysql"
HOSTNAME="localhost"
PORT="3306"
USERNAME="root"
PASSWORD="123456"

DBNAME="blog"
SRC_SQLFILE="./db.sql"
DIST_SQLFILE="/var/db.sql"
create_db_sql="create database IF NOT EXISTS ${DBNAME}"
remote_connect="grant all privileges on *.*  to 'root'@'%'"
update_root_pass="ALTER USER 'root'@'%' IDENTIFIED WITH mysql_native_password BY '${PASSWORD}'"
flush_pri="flush privileges"
import_sql="source ${DIST_SQLFILE}"

docker cp ${SRC_SQLFILE} ${CONTAINER_NAME}:${DIST_SQLFILE}

docker exec -it ${CONTAINER_NAME} /usr/bin/mysql -h${HOSTNAME} -P${PORT} -u${USERNAME} -p${PASSWORD} -e "${create_db_sql}"
docker exec -it ${CONTAINER_NAME} /usr/bin/mysql -h${HOSTNAME} -P${PORT} -u${USERNAME} -p${PASSWORD} ${DBNAME} -e "${import_sql}"
docker exec -it ${CONTAINER_NAME} /usr/bin/mysql -h${HOSTNAME} -P${PORT} -u${USERNAME} -p${PASSWORD} -e "${remote_connect}"
docker exec -it ${CONTAINER_NAME} /usr/bin/mysql -h${HOSTNAME} -P${PORT} -u${USERNAME} -p${PASSWORD} -e "${update_root_pass}"
docker exec -it ${CONTAINER_NAME} /usr/bin/mysql -h${HOSTNAME} -P${PORT} -u${USERNAME} -p${PASSWORD} -e "${flush_pri}"

cd /app/app/service/article/cmd && ./article >> ./output.log &

cd /app/app/service/poems/cmd && ./poems >> ./output.log &

cd /app/app/service/webinfo/cmd && ./webinfo>> ./output.log &

cd /app/app/interface/main/cmd && ./main>> ./output.log