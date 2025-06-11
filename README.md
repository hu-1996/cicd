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

## 打镜像(暂不可用)

```bash
# ./build-image.sh <os> <arch> <module>
# 前端没有镜像脚本

./build-image.sh linux amd64 cicd-server
```

## 打包

```bash
# !!!如果是arm64机器把GOARCH换成arm64即可
# cicd-server
cd cicd-server && go mod tidy && GOOS=linux GOARCH=amd64 go build -a -ldflags="-s -w" -o cicd-server main.go

# cicd-runner
cd cicd-runner && go mod tidy && GOOS=linux GOARCH=amd64 go build -a -ldflags="-s -w" -o cicd-runner main.go

# cicd-web
cd cicd-web && npm run build
```

## 部署

```bash
# cicd-web
build完成后出现dist目录
index.html、vite.svg -> cicd-server二进制文件同级目录下的web/views下
assets目录复制到与cicd-server二进制文件同级目录下的web/assets

# cicd-server
执行cicd-server下start.sh，需要先把cicd-web项目复制完成再启动

# cicd-runner
执行cicd-runner下start.sh
```