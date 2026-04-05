# Reed-Solomon Backup

> [!WARNING]
> 这是一个可用的最小版本，适合本地加密备份与恢复

一个使用 Go 编写的命令行工具，用于：

- 使用密码加密文件内容与文件名
- 用 Reed-Solomon 生成冗余分片
- 在恢复时只需提供任意一个分片文件（或 `.rsmeta`）和密码

## 特性

- 支持 `backup` / `restore` 两个子命令
- 密码输入支持：命令行参数优先，未提供时交互输入
- 强制要求：
  - `shares > 20`
  - `threshold > 80% * shares`
- 输出分片格式：`.rs.001`、`.rs.002` ...
- 元数据文件：`.rsmeta`

## 安装/构建

```bash
go build -o rsbackup .
```

## 备份

```bash
rsbackup backup --input ./example.txt --shares 24 --threshold 20 --password my-secret
```

如果不传密码：

```bash
rsbackup backup --input ./example.txt --shares 24 --threshold 20
```

程序会提示交互输入密码。

输出示例：

```text
<encrypted-prefix>.rsmeta
<encrypted-prefix>.rs.001
<encrypted-prefix>.rs.002
...
```

## 恢复

只需要任意一个分片文件或 `.rsmeta` 文件：

```bash
rsbackup restore --input ./<encrypted-prefix>.rs.007 --password my-secret
```

或：

```bash
rsbackup restore --input ./<encrypted-prefix>.rsmeta
```

## 实现说明

- 文件内容：先使用 AES-256-GCM 加密，再进行 Reed-Solomon 切片编码
- 文件名：使用同一主密钥加密，并转换为文件系统安全前缀
- 元数据：保存 KDF 参数、阈值、原始文件大小、加密后的文件名等
- 分片文件：带有轻量头部，支持从任意分片识别备份集合
