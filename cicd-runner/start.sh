#!/bin/bash

server_url=http://localhost:8029/api      # 服务器地址（runner机器可以访问到的）
runner_url=http://localhost:5913          # 运行器地址（server机器可以访问到的）

# -n 运行器名称
# -s 服务器地址
# -r 运行器地址
# -l 运行器标签，可多个，执行ci任务时，会根据标签匹配runner机器并执行任务
nohup /home/devops/ci/cicd-runner -n shanghai01 -s $server_url -r $runner_url -l sh_01 > runner.log 2>&1 &