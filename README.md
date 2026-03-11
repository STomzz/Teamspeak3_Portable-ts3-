# TS3 Portable Launcher

这个仓库提供一个 `TeamSpeak 3` 的 Windows 便携启动器方案。

## 封装便携版 EXE 前要准备什么

如果你的目标是封装出一个可直接双击运行的便携版 `TS3Portable.exe`，你需要先准备下面这些材料。

### 1. 必须准备的材料

- `TeamSpeak 3` Windows 客户端文件
  你需要自己从官方渠道获取，并确保里面包含 `ts3client_win64.exe` 或 `ts3client_win32.exe`。
- `Go`
  用来编译这个启动器。
- `goversioninfo`
  用来给最终的 Windows 可执行文件写入图标、版本信息和 manifest。

### 2. 这些材料要放在哪里

你需要把 `TeamSpeak 3` 客户端文件(安装好后的整个目录)打成一个 zip 压缩包，然后放到这个位置：

```text
payload/ts3-client-win64.zip
```

也就是说，在你的项目目录里最终应该是这样：

```text
teamSpeaker/
  payload/
    ts3-client-win64.zip(关键)
```

这个 zip 内部必须能找到下面至少一个文件：

- `ts3client_win64.exe`
- `ts3client_win32.exe`

### 3. 构建时脚本会自动处理什么

当你运行构建脚本时，它会自动完成这些事情：

- 读取 `payload/ts3-client-win64.zip`
- 复制到嵌入目录 `internal/payload/assets/ts3-client-win64.zip`
- 生成 Windows 资源文件
- 把图标、版本信息、manifest 和 TeamSpeak payload 一起打进最终的 `TS3Portable.exe`

### 4. 最终你需要手动准备的其实只有一个核心文件

真正必须由你自己准备、并放对位置的核心文件只有这一个：

```text
payload/ts3-client-win64.zip
```

其他像下面这些文件：

- `assets/windows/app.ico`
- `assets/windows/app.manifest`
- `VERSION`
- `scripts/build-windows.sh`

仓库里已经提供好了，你只需要按要求放入 TeamSpeak 客户端 zip，然后执行构建。

目标是做成适合个人使用的便携版：

- 构建出单个 `exe`
- 把官方 `TeamSpeak 3` 客户端压缩包嵌入这个 `exe`
- 第一次运行时自动解压到程序旁边
- 把用户数据尽量保存到本地 `data/` 目录，而不是 `%APPDATA%`
- 出错时给出可见错误弹窗，并把信息写到 `launcher.log`

## 这个方案解决什么问题

- 目标电脑不需要正常安装
- 配置和运行文件跟着程序目录走
- 方便整体复制到另一台 Windows 电脑
- 丢给你的好兄弟开黑,好兄弟懒得下载?丢给他一个单个的文件,直接运行!

## 重要限制

这个方案不能做到“永远只有一个文件”。

第一次运行后，程序会生成：

- `runtime/`：解压后的 TeamSpeak 运行文件
- `data/`：用户配置、缓存和本地数据

所以如果你想把完整的便携版搬到另一台电脑，应该复制整个目录，而不是只复制启动器 `exe`。

另外，这个启动器做的是“可交付的便携打包”，不是“完全沙箱化”。
如果 TeamSpeak 自身还写入了别的系统位置，这部分行为未必能被完全拦住。

## 法律和许可说明

本仓库不包含 `TeamSpeak 3` 官方二进制文件。你需要自己获得官方客户端，并仅在你确认许可条款允许的前提下进行打包和分发。

本仓库的开源许可证只适用于仓库中的启动器、构建脚本和相关辅助代码，不适用于 `TeamSpeak 3` 官方客户端本体，也不改变 `TeamSpeak 3` 自身的许可条款。

上传到 GitHub 时，不应提交这些内容：

- `payload/ts3-client-win64.zip`
- `dist/TS3Portable.exe`
- 任何解压后的 `TeamSpeak 3` 客户端文件
- 运行过程中生成的 `runtime/`、`data/`、`launcher.log`

## 目录说明

- `cmd/ts3-portable-launcher/main.go`：启动器入口
- `internal/payload`：处理嵌入或外部 payload 的加载逻辑
- `scripts/build-windows.sh`：适合 WSL2/Linux 的构建脚本
- `scripts/build-windows.ps1`：适合 Windows PowerShell 的构建脚本
- `assets/windows/app.ico`：默认应用图标，可替换
- `assets/windows/app.manifest`：Windows manifest
- `VERSION`：程序版本号
- `payload/ts3-client-win64.zip`：构建前由你自己提供的 TeamSpeak 3 客户端压缩包

## 准备 TeamSpeak 3 客户端压缩包

1. 准备官方 `TeamSpeak 3` Windows 客户端文件。
2. 把客户端文件打成一个 zip 压缩包。
3. 确保压缩包里包含真正的客户端可执行文件，例如 `ts3client_win64.exe`。
4. 把压缩包保存到：

```text
payload/ts3-client-win64.zip
```

这个 zip 可以是两种结构之一：

- 客户端文件直接位于压缩包根目录
- 客户端文件位于一个顶层文件夹里

启动器会在解压后的目录里自动查找：

- `ts3client_win64.exe`
- `ts3client_win32.exe`

## 构建

在安装了 Go 和 `goversioninfo` 的机器上执行：

```bash
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
```

然后执行：

```powershell
pwsh -File .\scripts\build-windows.ps1
```

构建结果输出到：

```text
dist\TS3Portable.exe
```

构建脚本会做这些事情：

- 把 payload zip 复制到嵌入目录
- 生成 Windows 版本信息资源
- 嵌入应用图标和 manifest
- 使用 `bundle` tag 进行编译

## 在 WSL2 中构建

可以。这个项目没有使用 `cgo`，所以可以在 `WSL2` 中直接交叉编译 Windows 可执行文件。

### 1. 在 WSL2 安装 Go

可以使用系统包，或者自己安装官方 Go。

例如在 Ubuntu / Debian 系的 WSL2 中：

```bash
sudo apt update
sudo apt install -y golang
```

安装后确认：

```bash
go version
```

### 2. 安装 `goversioninfo`

```bash
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
```

### 3. 在 WSL2 中构建 Windows EXE

把官方 TeamSpeak 3 客户端压缩包放到：

```text
payload/ts3-client-win64.zip
```

然后在项目根目录执行：

```bash
./scripts/build-windows.sh
```

构建完成后，产物位于：

```text
dist/TS3Portable.exe
```

## 运行方式

双击 `TS3Portable.exe` 即可。

第一次运行时它会：

1. 创建 `runtime\client`
2. 把内嵌的 TeamSpeak 3 客户端解压到该目录
3. 创建本地 `data\` 相关目录
4. 使用本地 profile 环境变量启动 TeamSpeak

如果启动失败：

- 会弹出错误对话框
- 会在程序同级目录写入 `launcher.log`

## 便携数据目录

启动器会为 TeamSpeak 进程设置这些环境变量：

- `APPDATA` -> `data\Roaming`
- `LOCALAPPDATA` -> `data\Local`
- `USERPROFILE` -> `data\Home`
- `HOME` -> `data\Home`

这样做是为了尽量把 TeamSpeak 常见的用户数据目录引导到程序本地目录中，例如 `%APPDATA%\TS3Client`。

## Windows 测试建议

建议至少做下面这些测试：

1. 在你的开发电脑上双击 `TS3Portable.exe`，确认首次解压和启动都正常。
2. 退出后再次启动，确认第二次不会重复异常解压。
3. 检查程序目录下是否生成了 `runtime/`、`data/`，并确认配置确实落在本地目录。
4. 把整个目录复制到另一台 Windows 电脑，再次双击测试。
5. 在非管理员权限下测试。
6. 把程序放到带中文、空格的路径里测试。
7. 故意删掉 `runtime/client` 中的部分文件，再启动，确认它能自动重新解压恢复。

## 图标和版本信息

当前构建会把这些信息写入最终的 Windows 可执行文件：

- 应用图标：`assets/windows/app.ico`
- 文件版本：来自 `VERSION`
- 产品名：`TS3 Portable Launcher`
- manifest：`assets/windows/app.manifest`

如果你要换成自己的正式品牌：

1. 替换 `assets/windows/app.ico`
2. 修改 `VERSION`
3. 按需调整构建脚本里生成的 `FileDescription`、`ProductName`、`CompanyName`

## 不嵌入 payload 的开发模式

如果你只是本地调试，可以不使用 `bundle`：

```powershell
go build -o .\dist\TS3Portable-dev.exe .\cmd\ts3-portable-launcher
```

这种模式下，程序会从可执行文件同级目录下读取：

```text
dist\payload\ts3-client-win64.zip
```

## 已知风险

- 如果 TeamSpeak 还会写入别的 Windows 系统位置，并且这些位置不受当前环境变量影响，那么仍然可能留下非便携数据
- 插件、协议关联、自动更新等功能可能默认假设它是正常安装版
- 当前方案没有处理自动更新，如需更新版本，需要重新替换 payload 并重新构建
- 如果杀毒软件或系统策略拦截未知 EXE，第一次运行的解压过程可能被阻止
