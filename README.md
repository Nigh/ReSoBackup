# ReSo Backup

> [!WARNING]
> 这是一个实现 [这个帖子的需求](https://meta.appinn.net/t/topic/83595) 的可用的最小版本，适合本地加密备份与恢复

一个使用 Go 编写的命令行工具，用于：

- 使用密码加密文件内容与文件名（可选）
- 用 Reed-Solomon 生成冗余分片
- 在恢复时只需提供任意一个分片文件（或 `.rsmeta`）

## 特性

- 支持 `backup` / `restore` 两个子命令
- 加密为可选项：可选择是否加密文件内容与文件名
- 密码输入支持：命令行参数优先，未提供时交互输入
- 备份参数硬限制：
  - `3 <= shares <= 128`
  - `1 <= threshold <= shares`
- 默认参数：
  - `shares = 8`
  - `threshold = 5`
- 当参数组合存在风险时，会要求交互二次确认：
  - 冗余过低：分片稍有丢失就可能无法恢复
  - 冗余过高：总存储数据量会明显增大
- 输出分片格式：`.rs.001`、`.rs.002` ...
- 元数据文件：`.rsmeta`
- 恢复时自动检测备份是否加密，无需手动指定

## 使用


### 1. 备份

```bash
rsbackup backup --input ./example.txt --shares 8 --threshold 5
```

启用加密：

```bash
rsbackup backup --input ./example.txt --shares 8 --threshold 5 --encrypt --password my-secret
```

启用加密并加密文件名：

```bash
rsbackup backup --input ./example.txt --encrypt --encrypt-filename --password my-secret
```

如果不传密码：

```bash
rsbackup backup --input ./example.txt --encrypt
```

程序会提示交互输入密码。

如果参数组合触发风险提示，程序会显示 warning 并要求再次确认；默认回答为 `No`，不确认则本次备份会被取消。

输出示例：

```text
<file-prefix>.rsmeta
<file-prefix>.rs.001
<file-prefix>.rs.002
...
```

### 2. 恢复

只需要任意一个分片文件或 `.rsmeta` 文件：

```bash
rsbackup restore --input ./<file-prefix>.rs.007 --password my-secret
```

或：

```bash
rsbackup restore --input ./<file-prefix>.rsmeta
```

如果备份未加密，无需提供密码：

```bash
rsbackup restore --input ./<file-prefix>.rsmeta
```

## 构建

```bash
make
```

默认会构建全部发布版本

## 额外说明

- 文件内容：可选择使用 AES-256-GCM 加密，再进行 Reed-Solomon 切片编码
- 文件名：可选择加密，使用同一主密钥加密并转换为文件系统安全前缀
- 元数据：保存 KDF 参数、阈值、原始文件大小、加密标志等
- 分片文件：带有轻量头部，支持从任意分片识别备份集合
- 存储开销大致与 `shares / threshold` 成正比；`shares - threshold` 越小，可容忍的分片丢失数量越少
