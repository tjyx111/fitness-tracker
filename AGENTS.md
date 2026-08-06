# 助手项目说明（仓库 fitness-tracker）

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

该脚本通过 `dockerbuild.sh` 构建并测试 Linux/amd64 二进制，然后生成 `dist/assistant-linux-amd64.tar.gz`。默认还会上传发布包；仅构建本地包时使用：

```bash
SKIP_UPLOAD=1 ./package_release.sh
```

发布包不得包含 `.fitness-tracker.env` 或真实 token。

### 构建 Android APK

Android 客户端位于 `android/`，是加载当前 HTTPS 前端的 WebView 应用。运行：

```bash
./scripts/build_android.sh
```

签名密钥固定保存在仓库外的 `/root/.config/fitness-tracker/android/fitness-release.jks`，不得提交、打包、上传或重新生成替换。构建产物为 `dist/assistant.apk`。云端文件路径为 `/root/lbs/fitness/downloads/assistant.apk`，下载地址为 `https://111.230.63.109:19797/downloads/assistant.apk`。

Android WebView 会通过 `frontend/service-worker.js` 缓存最近成功加载的页面和 GET API 响应，离线模式仅供查看，所有写操作仍需联网。每日训练提醒默认关闭，由用户在 App 内主动设置；提醒配置会在设备重启后恢复。

### 上传 HTML 分析报告

服务将 `.htm`/`.html` 报告保存在 `DATA_DIR/reports`，可通过 `REPORT_DIR`
覆盖。App 的“报告”页面会读取 `/api/reports` 列表并展示报告内容。

上传接口必须配置仓库外的 `REPORT_UPLOAD_TOKEN`，然后在开发机运行：

```bash
REPORT_UPLOAD_TOKEN=... ./scripts/upload_report.sh report.htm
```

也可以直接请求 `PUT /api/reports/{name.htm}`，请求体为 HTML，认证头为
`Authorization: Bearer ...`。云端 systemd 从仓库外的
`/root/.config/fitness-tracker/fitness-tracker.env` 读取 token。不得把真实 token
写入仓库、发布包或日志。

### 云端首次部署

将发布包解压到云主机后启动服务：

```bash
chmod +x assistant start_cloud.sh sync_from_cloud.sh
./start_cloud.sh start
```

后续使用以下命令管理服务：

```bash
./start_cloud.sh status
./start_cloud.sh restart
./start_cloud.sh stop
```

新云主机部署目录为 `/root/lbs/fitness`，systemd 服务为 `assistant.service`，监听 `0.0.0.0:19797`，公网地址为 `https://111.230.63.109:19797/`。云端通过 `TLS_CERT_FILE` 和 `TLS_KEY_FILE` 启用 HTTPS，不请求或校验客户端证书；证书文件位于 `/root/lbs/fitness/tls`。如部署地址变化，必须通过 `LISTEN_ADDR` 覆盖，不能为新环境继续硬编码地址。

### 从云端同步数据库到本地

先停止本地服务，确保本地 SQLite 数据库没有被占用，然后运行：

```bash
./sync_from_cloud.sh
```

脚本从云端下载一致的 SQLite 快照，执行 `PRAGMA integrity_check`，备份现有本地数据库后再替换，然后将 `backend/data/fitness.db` 提交并推送到当前 Git 仓库。可用 `CLOUD_URL` 和 `LOCAL_DB` 覆盖默认地址及目标路径；可用 `GIT_COMMIT=0` 跳过提交，`GIT_PUSH=0` 跳过推送，`GIT_REMOTE` 和 `GIT_REF` 覆盖推送目标。

## 关键决策

- Git 仓库、源码目录、SQLite 数据库名、仓库外配置目录、Android `applicationId` 和签名身份继续使用 `fitness-tracker` 兼容命名；面向用户及发布运行的产品名、二进制、发布包、APK 和 systemd 服务使用“助手”或 `assistant`。
- 同步接口不做应用层认证，任何能访问 `/api/sync/database` 的客户端都可以下载完整 SQLite 数据库。必须通过网络层限制访问范围，例如内网、VPN、SSH 隧道、防火墙或反向代理 ACL。
- 当前同步方向是云端到本地的单向覆盖，不进行双向合并。同步前会备份本地数据库，但本地未上传的数据仍可能因覆盖而从工作数据库中消失。
- 云端通过 SQLite `VACUUM INTO` 生成一致快照，以支持数据库处于 WAL 模式或服务仍在写入的场景；本地替换数据库时服务必须停止。
- 默认 `CLOUD_URL` 使用新云主机的 HTTPS 地址，同步脚本从仓库外的 `/root/.config/fitness-tracker/tls` 读取 `ca.crt` 校验服务端证书，不使用客户端证书。

## 验证要求

修改启动或同步逻辑后至少运行：

```bash
bash -n start.sh start_cloud.sh sync_from_cloud.sh package_release.sh scripts/generate_tls_certs.sh
cd backend
GOCACHE=/tmp/go-build-cache go test ./...
```

修改 Android 工程后还必须运行 `./scripts/build_android.sh`，确认 Android Lint、release 构建和 `apksigner` 验证成功。
