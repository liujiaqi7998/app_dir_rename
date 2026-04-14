# app_dir_rename

一个面向 Windows 的小工具软件。

把 `exe` 或文件夹拖到本工具上后，会自动在目标目录生成 `desktop.ini`，用于给文件夹设置显示名称和图标，而不真正修改文件夹名。

## 功能说明

### 1. 拖入 `exe`

当用户把一个 `exe` 拖到本工具上时，程序会在这个 `exe` 所在目录生成 `desktop.ini`：

```ini
[.ShellClassInfo]
LocalizedResourceName=exe 文件名（不含 .exe）
IconResource=该 exe 路径,0
[ViewState]
Mode=
Vid=
FolderType=Generic
```

例如拖入 `D:\Games\Launcher.exe`，会在 `D:\Games` 下生成 `desktop.ini`，其中：

- `LocalizedResourceName=Launcher`
- `IconResource=D:\Games\Launcher.exe,0`

### 2. 拖入文件夹

当用户把一个文件夹拖到本工具上时，程序会进入交互模式：

1. 弹出输入框，询问文件夹显示名称
2. 弹出系统文件选择框，选择一个 `exe` 作为文件夹图标
3. 在该文件夹内生成 `desktop.ini`

这里不会真实重命名文件夹，而是通过 `LocalizedResourceName` 改变资源管理器里的显示名称。

## 使用方法

### 运行方式

1. 先构建出 `app_dir_rename.exe`
2. 把目标 `exe` 或文件夹拖到 `app_dir_rename.exe` 上
3. 程序会自动处理并弹窗提示结果

### 处理结果

程序会自动：

- 写入 `desktop.ini`
- 为 `desktop.ini` 设置隐藏和系统属性
- 为目标文件夹设置只读属性，便于 Windows 识别文件夹自定义配置

### 注意事项

- 仅支持 Windows
- 仅支持拖入 `exe` 文件或文件夹
- 文件夹模式下，图标文件必须选择 `exe`
- 如果资源管理器没有立刻刷新显示效果，可以尝试刷新目录，或重新打开该目录

## 构建方法

### 开发环境要求

- Go 1.20 或更高版本
- Windows

### 本地构建

在项目根目录执行：

```powershell
go build -o app_dir_rename.exe .
```

### 运行测试

```powershell
go test ./...
```

## 项目结构

```text
.
├─ main.go
├─ main_test.go
├─ go.mod
└─ Readme.md
```

## 实现说明

项目当前使用 Go 标准库配合 PowerShell 原生能力实现：

- 通过命令行参数接收拖拽目标路径
- 通过 PowerShell 弹出输入框、文件选择框和提示框
- 通过 `attrib` 设置 `desktop.ini` 和文件夹属性

## 适用场景

适合用来：

- 给游戏启动器目录生成自定义图标
- 给工具目录设置更友好的显示名称
- 快速把某个 `exe` 作为文件夹图标来源