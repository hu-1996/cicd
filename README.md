# cicd

## 环境变量

- CICD_ADMIN_USERNAME: 管理员用户名
- CICD_ADMIN_PASSWORD: 管理员密码

## 启动

```bash
go run main.go
```

## 前端

```bash
cd cicd-web
npm run dev
```

## 前端部署

代码放到 cicd-web/web下，然后启动cicd-server，dist/assets放到cicd-server/web/assets下，其余放到cicd-server/web/views下

## 打镜像(暂不可用)

```bash
# ./build-image.sh <os> <arch> <module>
# 前端没有镜像脚本

./build-image.sh linux amd64 cicd-server
```

## 打包

```bash
# cicd-server
cd cicd-server && go mod tidy && GOOS=linux GOARCH=amd64 go build -a -o cicd-server main.go

# cicd-runner
cd cicd-runner && go mod tidy && GOOS=linux GOARCH=amd64 go build -a -o cicd-runner main.go

# cicd-web
cd cicd-web && npm run build
```

## 运行

```bash
# cicd-server
执行cicd-server下start.sh

# cicd-runner
执行cicd-runner下start.sh

# cicd-web
build完成后出现dist目录
index.html、vite.svg -> cicd-server二进制文件同级目录下的web/views下
assets目录复制到与cicd-server二进制文件同级目录下的web/assets
```