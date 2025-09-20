#!/bin/bash

getProcess=$(ss -tnpl|grep -w 8080|sed 's/.*pid=\([0-9]*\).*/\1/')
if [ "$getProcess" != "" ]; then
kill -9 $getProcess
fi

getProcess=$(ss -tnpl|grep -w 8081|sed 's/.*pid=\([0-9]*\).*/\1/')
if [ "$getProcess" != "" ]; then
kill -9 $getProcess
fi
