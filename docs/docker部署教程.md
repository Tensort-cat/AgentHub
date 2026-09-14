我已经按这个项目整理好了 Docker 相关文件，关键是：

- Dockerfile：用来把 Go 后端编译成镜像
- docker-compose.yml：同时起 `agenthub` + `MySQL 8.0` + `Redis 7.4`
- config.toml：已改成容器内网络地址 `mysql` / `redis`
- tables.sql：数据库初始化脚本

下面给你一份“保姆版 Windows + Docker Desktop”手动部署教程，照着做就能完成“先打包，再运行容器”的全流程。

1. 先确认环境

在 Windows 的 PowerShell / CMD 里执行：

```powershell
docker version
docker compose version
```

如果能正常输出版本号，说明 Docker Desktop 安装正常。

如果你没装 Docker Desktop，请先装好，然后重启电脑。

2. 进入项目目录

在 PowerShell 里，切到项目根目录：

```powershell
cd D:\dev_soft\AgentHub
```

3. 先看一下项目现在的 Docker 配置

你当前这个项目，后端服务依赖：

- Go 1.26.7
- MySQL 8.0
- Redis 7.4-alpine

并且在 Docker 中，后端、MySQL、Redis 是三个独立容器。

项目里的关键配置已被改成适配容器网络：

- 后端访问 MySQL：`mysql:3306`
- 后端访问 Redis：`redis:6379`
- 后端监听：`0.0.0.0:8000`

对应的配置文件是 config.toml，内容会类似这样：

```toml
[mainConfig]
appName = "AgentHub"
host = "0.0.0.0"
port = 8000

[mysqlConfig]
host = "mysql"
port = 3306
user = "root"
password = "123456"
database = "agenthub"

[redisConfig]
addr = "redis:6379"
password = "123456"
db = 0
dimension = 2048
```

这里要注意：

- `mysql` 和 `redis` 是容器名，不是你电脑本机 IP
- 容器之间能通过 Docker 网络自动通信
- 这样后端在容器里就能找到数据库和缓存

4. 先构建镜像（打包）

在项目根目录执行：

```powershell
docker build -t agenthub:latest .
```

这一步的意思是：

- 读取 Dockerfile
- 用 Go 官方镜像编译项目
- 生成一个最终的应用镜像，名字是 `agenthub:latest`

如果没报错，就说明镜像打包成功。

常见问题：

- 报 `permission denied`：通常 Docker Desktop 没启动，重启它
- 报 Go 版本问题：本机 Go 版本和 go.mod 里要求不一致，但 Docker 镜像里是 `golang:1.26.7`，所以容器内编译没问题
- 报 `COPY failed`：检查是否在项目根目录，`Dockerfile` 是否存在

5. 启动 MySQL 和 Redis 容器

先单独启动数据库和缓存：

```powershell
docker compose up -d mysql redis
```

这会启动两个容器：

- `agenthub-mysql`
- `agenthub-redis`

然后可以查看是否正常启动：

```powershell
docker ps
```

你应该能看到类似：

- `agenthub-mysql`
- `agenthub-redis`

6. 初始化 MySQL 数据库结构

项目里有 tables.sql，用于建表。

先把 SQL 文件拷进 MySQL 容器：

```powershell
docker cp .\tables.sql agenthub-mysql:/tmp/tables.sql
```

然后执行：

```powershell
docker exec -it agenthub-mysql mysql -uroot -p123456 -e "CREATE DATABASE IF NOT EXISTS agenthub;"
docker exec -it agenthub-mysql mysql -uroot -p123456 agenthub < /tmp/tables.sql // Linux 环境用这个
Get-Content .\tables.sql | docker exec -i agenthub-mysql mysql -uroot -p123456 agenthub // powershell 用这个
```

这一步的意思是：

- 创建数据库 `agenthub`
- 把 tables.sql 中的表结构导入进 MySQL

如果你想确认有没有建表，可以执行：

```powershell
docker exec -it agenthub-mysql mysql -uroot -p123456 -e "SHOW DATABASES;"
docker exec -it agenthub-mysql mysql -uroot -p123456 agenthub -e "SHOW TABLES;"
```

7. 启动后端容器

现在启动后端容器：

```powershell
docker compose up -d agenthub
```

这一步会：

- 从 Dockerfile 创建/重建后端镜像
- 启动 `agenthub-app`
- 自动等待 MySQL / Redis 健康检查通过后再启动

8. 检查服务是否可访问

在浏览器访问：

```text
http://localhost:8000
```

如果你看到返回404且容器日志有访问记录就是部署好了。

