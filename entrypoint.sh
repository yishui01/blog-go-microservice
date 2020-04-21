#!/bin/sh

cd /app/app/service/article/cmd && ./article >> ./output.log &

cd /app/app/service/poems/cmd && ./poems >> ./output.log &

cd /app/app/service/webinfo/cmd && ./webinfo>> ./output.log &

cd /app/app/interface/main/cmd && ./main>> ./output.log