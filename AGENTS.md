# Fitness Tracker 项目说明

## 启动流程

## 远端登录规则

需要通过 JumpServer/xssh 登录远端机器时，Google Authenticator/TOTP code
使用 `~/.bashrc` 中的 `jpcode` 命令生成。不要使用缓存 code 或手工记忆的
code。

固定命令：

```bash
JUMP_TOTP_CODE="$(bash -ic 'jpcode' 2>/dev/null | tail -n 1)"
```

如果没有登录成功，优先按“`jpcode` 没有加载或 code 已过期”处理：重新执行
固定命令生成 code，并且只做一次新的登录尝试。
不得把 TOTP secret、生成出的 code、JumpServer 密码或 xssh 密码写入仓库、提交、
日志摘要或用户可见文档。

### 本地开发

在仓库根目录运行：

```bash
./start.sh
```

`start.sh` 会执行以下操作：

1. 将根目录 `frontend/` 同步到 `backend/frontend/`。
2. 在 `backend/` 中执行 `go run -buildvcs=false .`。

本地默认监听 `0.0.0.0:8080`，默认数据库位于 `backend/data/fitness.db`。可通过 `LISTEN_ADDR`、`DATA_DIR` 等环境变量覆盖。

### 构建发布包

```bash
./package_release.sh
```

该脚本通过 `dockerbuild.sh` 构建并测试 Linux/amd64 二进制，然后生成 `dist/fitness-tracker-linux-amd64.tar.gz`。默认还会上传发布包；仅构建本地包时使用：

```bash
SKIP_UPLOAD=1 ./package_release.sh
```

发布包不得包含 `.fitness-tracker.env` 或真实 token。

### 云端首次部署

将发布包解压到云主机后启动服务：

```bash
chmod +x fitness-tracker start_cloud.sh sync_from_cloud.sh
./start_cloud.sh start
```

后续使用以下命令管理服务：

```bash
./start_cloud.sh status
./start_cloud.sh restart
./start_cloud.sh stop
```

云端默认监听 `183.36.16.116:19797`。如部署地址变化，必须通过 `LISTEN_ADDR` 覆盖，不能为新环境继续硬编码地址。

### 从云端同步数据库到本地

先停止本地服务，确保本地 SQLite 数据库没有被占用，然后运行：

```bash
./sync_from_cloud.sh
```

脚本从云端下载一致的 SQLite 快照，执行 `PRAGMA integrity_check`，备份现有本地数据库后再替换，然后将 `backend/data/fitness.db` 提交并推送到当前 Git 仓库。可用 `CLOUD_URL` 和 `LOCAL_DB` 覆盖默认地址及目标路径；可用 `GIT_COMMIT=0` 跳过提交，`GIT_PUSH=0` 跳过推送，`GIT_REMOTE` 和 `GIT_REF` 覆盖推送目标。

## 关键决策

- 同步接口不做应用层认证，任何能访问 `/api/sync/database` 的客户端都可以下载完整 SQLite 数据库。必须通过网络层限制访问范围，例如内网、VPN、SSH 隧道、防火墙或反向代理 ACL。
- 当前同步方向是云端到本地的单向覆盖，不进行双向合并。同步前会备份本地数据库，但本地未上传的数据仍可能因覆盖而从工作数据库中消失。
- 云端通过 SQLite `VACUUM INTO` 生成一致快照，以支持数据库处于 WAL 模式或服务仍在写入的场景；本地替换数据库时服务必须停止。
- 当前默认 `CLOUD_URL` 使用 HTTP，生产环境应使用 HTTPS、VPN 或 SSH 隧道；在完成传输层保护前，不应通过公网明文同步。

## 验证要求

修改启动或同步逻辑后至少运行：

```bash
bash -n start.sh start_cloud.sh sync_from_cloud.sh package_release.sh
cd backend
GOCACHE=/tmp/go-build-cache go test ./...
```
