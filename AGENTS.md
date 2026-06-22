# Fitness Tracker 项目说明

## 启动流程

### 本地开发

在仓库根目录运行：

```bash
./start.sh
```

`start.sh` 会执行以下操作：

1. 如果根目录的 `.fitness-tracker.env` 不存在，则生成一次 32 字节随机 `SYNC_TOKEN` 并以 `600` 权限保存。
2. 如果 token 文件已经存在，则读取并复用，启动时不得重新生成。
3. 将根目录 `frontend/` 同步到 `backend/frontend/`。
4. 在 `backend/` 中执行 `go run -buildvcs=false .`。

本地默认监听 `0.0.0.0:8080`，默认数据库位于 `backend/data/fitness.db`。可通过 `LISTEN_ADDR`、`DATA_DIR` 等环境变量覆盖。

### 构建发布包

```bash
./package_release.sh
```

该脚本通过 `dockerbuild.sh` 构建并测试 Linux/amd64 二进制，然后生成 `dist/fitness-tracker-linux-amd64.tar.gz`。默认还会上传发布包；仅构建本地包时使用：

```bash
SKIP_UPLOAD=1 ./package_release.sh
```

发布包包含 token 读取脚本，但绝不能包含 `.fitness-tracker.env` 或真实 token。

### 云端首次部署

将发布包解压到云主机后，先把本地生成的同一个 token 配置到云端：

```bash
chmod +x fitness-tracker start_cloud.sh sync_from_cloud.sh
SYNC_TOKEN='本地生成的token' ./start_cloud.sh configure-token
./start_cloud.sh start
```

`configure-token` 会在云端脚本目录生成权限为 `600` 的 `.fitness-tracker.env`。后续使用以下命令管理服务，无需重新配置 token：

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

脚本自动读取根目录 `.fitness-tracker.env`，从云端下载一致的 SQLite 快照，执行 `PRAGMA integrity_check`，备份现有本地数据库后再替换。可用 `CLOUD_URL` 和 `LOCAL_DB` 覆盖默认地址及目标路径。

## 关键决策

- `SYNC_TOKEN` 是长期共享密钥：本地只生成一次，云端服务端与本地同步客户端必须使用同一个值。仅在泄露或主动轮换时重新生成。
- 环境变量 `SYNC_TOKEN` 优先于 token 文件；正常持久化使用脚本目录下的 `.fitness-tracker.env`。该文件权限必须为 `600`，已被 `.gitignore` 排除，严禁提交、打包、写入日志或文档。
- 云端启动必须存在 `SYNC_TOKEN`；未配置时 `start_cloud.sh` 应拒绝启动，避免同步接口在错误配置下运行。
- token 只授权下载完整数据库，因此泄露应按数据库泄露处理并立即轮换。本文件和示例中不得记录真实 token。
- 当前同步方向是云端到本地的单向覆盖，不进行双向合并。同步前会备份本地数据库，但本地未上传的数据仍可能因覆盖而从工作数据库中消失。
- 云端通过 SQLite `VACUUM INTO` 生成一致快照，以支持数据库处于 WAL 模式或服务仍在写入的场景；本地替换数据库时服务必须停止。
- 当前默认 `CLOUD_URL` 使用 HTTP，Bearer token 在不可信网络上不安全。生产环境应使用 HTTPS、VPN 或 SSH 隧道；在完成传输层保护前，不应通过公网明文同步。
- 不要把 `export SYNC_TOKEN=...` 作为云端唯一配置方式，因为进程或主机重启后环境变量可能丢失；使用 `configure-token` 持久化。

## 验证要求

修改启动、同步或 token 逻辑后至少运行：

```bash
bash -n start.sh start_cloud.sh sync_from_cloud.sh sync_token_env.sh package_release.sh
cd backend
GOCACHE=/tmp/go-build-cache go test ./...
```
