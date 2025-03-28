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

## 打镜像

```bash
# ./build-image.sh <os> <arch> <module>
# 前端没有镜像脚本

./build-image.sh linux amd64 cicd-server
```

## 打包

```bash
GOOS=linux GOARCH=amd64 go build -a -o cicd-server cicd-server/main.go
```