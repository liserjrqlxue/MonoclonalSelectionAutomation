# 单克隆挑选自动化 Monoclonal Selection Automation

## 实现功能

- 输入一个 URL；
- 解析 path 和 query 参数；
- 提取参数：order_id、user_email、password、key；
- 构造新的下载 URL；
- 下载 zip 文件保存到 {order_id}/{order_id}.os_all_file.zip；
- 解压 zip 到 {order_id}/{order_id}.os_all_file/。

## 封装为一个可安装的 CLI 工具

- 命令行参数解析（使用 cobra）
- 重试机制（带最大重试次数）
- 打包为可执行文件（通过 go install 安装）

## ✅ 功能总览

- [x] 单线程安全下载
- [x] 重试机制
- [x] 实时进度条
- [x] 可选 SHA256 校验
- [x] 解压 ZIP

## ✅ 目录结构建议：

```
MonoclonalSelectionAutomation/
├── cmd/
│   └── root.go         // cobra 入口命令
├── downloader/
│   └── downloader.go   // 下载逻辑，多线程、重试
├── unzip/
│   └── unzip.go        // zip 解压逻辑
├── main.go
├── go.mod

```

## 🛠️ 构建和安装

```
go install
```




