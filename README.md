<p align="center">
  <img src="build/appicon.png" alt="ReSo Backup" width="128" height="128">
</p>

<h1 align="center">ReSo Backup</h1>

<p align="center">
  <a href="README_EN.md">English</a> | 中文
</p>

<p align="center">
  基于 Reed-Solomon 纠删码的加密备份工具 —— 把文件切成多份分片，分散存储，任意几份丢失也能恢复
</p>

<p align="center">
  <a href="#工作原理">工作原理</a> · <a href="#使用场景">使用场景</a> · <a href="#快速上手">快速上手</a> · <a href="#gui-模式">GUI</a> · <a href="#cli-模式">CLI</a> · <a href="#构建">构建</a>
</p>


## 界面预览

<p align="center">
  <img src="screenshot.png" alt="ReSo Backup 界面截图" width="600">
</p>

## 工作原理

ReSo Backup 使用 **Reed-Solomon 纠删码** 技术，将一个文件切分成 N 份分片（称为"分片"），并额外生成校验数据。只要拿到其中任意 K 份（K ≤ N），就能完整还原原始文件。

你可以把它想象成：

> 把一个拼图拆成 N 块，每一块都看不出完整画面。但只要有任意 K 块，就能拼出原图。即使丢掉了 N − K 块，也完全不影响恢复。

整个备份流程如下：

```
原始文件
   │
   ▼
[可选] AES-256-GCM 加密（用你的密码加密文件内容和文件名）
   │
   ▼
Reed-Solomon 编码（切成 N 份，任意 K 份可恢复）
   │
   ▼
输出：1 个元数据文件 (.rsmeta) + N 个分片文件 (.rs.001 ~ .rs.N)
```

**恢复时**，你只需要提供任意一个分片文件（或 `.rsmeta` 元数据文件），程序会自动找到同组的所有分片，凑够 K 份即可还原。

### 两个关键参数

| 参数 | 含义 |
|------|------|
| **分片数 (shares)** | 文件总共被切成几份 |
| **恢复阈值 (threshold)** | 至少需要几份才能恢复 |

分片数 − 恢复阈值 = 可以容忍丢失的份数。

## 使用场景

假设你要备份一个重要文件，希望：
- 任何一个网盘服务商都**看不到文件内容**
- 仅凭本地数据也**无法单独还原**
- 任何一个存储位置出问题都**不影响恢复**

### 推荐方案：5 份分片，3 份可恢复

设置分片数 = **5**，恢复阈值 = **3**：

```bash
rsbackup backup --input important.docx --shares 5 --threshold 3 --encrypt --password "my-secret"
```

程序会生成以下文件：

```
important.docx.rsmeta          ← 元数据（记录参数、加密信息）
important.docx.rs.001          ← 分片 1
important.docx.rs.002          ← 分片 2
important.docx.rs.003          ← 分片 3
important.docx.rs.004          ← 分片 4
important.docx.rs.005          ← 分片 5
```

然后这样分配存储：

| 存储位置 | 放哪些文件 |
|----------|-----------|
| 网盘 A | `.rs.001` |
| 网盘 B | `.rs.002` |
| 网盘 C | `.rs.003` |
| 本地硬盘 | `.rs.004` + `.rs.005` + `.rsmeta` |

这样的好处是：

- **隐私保障**：每个网盘只拿到 1 份分片，加上文件经过加密，任何单一网盘服务商都无法获取你的文件内容
- **本地安全**：本地只有 2 份分片（少于恢复阈值 3），仅凭本地数据无法还原文件，即使电脑被他人访问也是安全的
- **容灾能力**：任意一个网盘文件丢失，甚至两个网盘同时出问题，都不影响恢复

### 恢复方案一：本地文件丢失

本地硬盘坏了，但三个网盘的分片还在。从三个网盘分别下载分片，放在同一个目录下，然后：

```bash
rsbackup restore --input important.docx.rs.001 --password "my-secret"
```

程序会自动发现同组的 3 份分片（001、002、003），满足恢复阈值 3，文件完整还原。

### 恢复方案二：本地文件完好

本地文件没丢，只需要从任意一个网盘再下载一份分片就够了（本地 2 份 + 网盘 1 份 = 3 份）：

```bash
# 把从网盘下载的 .rs.001 放到本地分片所在的目录
rsbackup restore --input important.docx.rs.004 --password "my-secret"
```

程序会找到本地的 004、005 和刚下载的 001，共 3 份分片，完成恢复。

### 其他常见方案

| 场景 | 分片数 | 恢复阈值 | 可丢失份数 | 存储膨胀 |
|------|--------|----------|-----------|---------|
| 保守方案（高冗余） | 8 | 3 | 5 | 2.67x |
| 均衡方案 | 5 | 3 | 2 | 1.67x |
| 紧凑方案 | 4 | 3 | 1 | 1.33x |
| 无冗余（不推荐） | 3 | 3 | 0 | 1.00x |

> 存储膨胀 = 分片数 / 恢复阈值，即备份文件总大小约为原始文件的多少倍。

## 快速上手

### 安装

从 [Releases](https://github.com/user/Reed-Solomon-Backup/releases) 页面下载对应平台的可执行文件。

### GUI 模式

双击运行 `rsbackup`（无参数）即可打开图形界面：

```bash
rsbackup
```

<p align="center">
  <img src="screenshot.png" alt="GUI 界面" width="600">
</p>

在 GUI 中：
1. 切换到 **备份** 标签页
2. 选择要备份的文件
3. 拖动滑块设置分片数和恢复阈值
4. 可选：勾选加密并输入密码
5. 点击 **开始备份**

恢复时切换到 **恢复** 标签页，选择任意一个分片文件或 `.rsmeta` 文件即可。

### CLI 模式

#### 备份

```bash
# 基本备份（默认 8 份分片，5 份可恢复）
rsbackup backup --input ./document.pdf

# 加密备份
rsbackup backup --input ./document.pdf --encrypt --password "my-secret"

# 自定义分片参数
rsbackup backup --input ./document.pdf --shares 5 --threshold 3 --encrypt --password "my-secret"

# 同时加密文件名
rsbackup backup --input ./document.pdf --encrypt --encrypt-filename --password "my-secret"
```

如果不传密码，程序会在终端提示输入。

#### 恢复

```bash
# 从任意一个分片文件恢复
rsbackup restore --input ./document.pdf.rs.001 --password "my-secret"

# 从元数据文件恢复
rsbackup restore --input ./document.pdf.rsmeta --password "my-secret"

# 未加密的备份无需密码
rsbackup restore --input ./document.pdf.rsmeta

# 指定输出目录
rsbackup restore --input ./document.pdf.rs.001 --out-dir ./restored/
```

#### 完整参数

```text
rsbackup                          启动 GUI（无参数时）
rsbackup backup  --input <file> [--shares 8] [--threshold 5] [--password <pwd>] [--out-dir <dir>] [--encrypt] [--encrypt-filename]
rsbackup restore --input <any .rs.NNN or .rsmeta file> [--password <pwd>] [--out-dir <dir>]
```

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--input` | 源文件路径（备份）或分片/元数据路径（恢复） | 必填 |
| `--shares` | 总分片数 (3~128) | 8 |
| `--threshold` | 恢复所需最少分片数 (1~shares) | 5 |
| `--password` | 加密密码（可选，为空时交互输入） | - |
| `--out-dir` | 输出目录 | 源文件目录 / 当前目录 |
| `--encrypt` | 启用 AES-256-GCM 加密 | false |
| `--encrypt-filename` | 加密原始文件名（需配合 `--encrypt`） | false |

## 安全说明

- **加密算法**：AES-256-GCM（认证加密，防篡改）
- **密钥派生**：scrypt（抗暴力破解，参数 N=32768, r=8, p=1）
- **文件名加密**：可选，使用同一主密钥加密原始文件名
- **密码安全**：密码不会存储在任何文件中，丢失密码 = 无法恢复加密备份

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go |
| GUI 框架 | Wails v3 |
| 前端 | Svelte 5 + DaisyUI 5 + Tailwind CSS 4 |
| 加密 | AES-256-GCM + scrypt KDF |
| 纠删码 | klauspost/reedsolomon |
| 构建 | Make + Taskfile |

## 构建

```bash
# 构建当前平台
task build

# 构建全部平台
make

# 开发模式（热重载）
wails3 dev -config ./build/config.yml -port 9245
```

## 额外说明

- 元数据文件 (`.rsmeta`)：保存 KDF 参数、分片参数、原始文件大小、加密标志等恢复所需的全部信息
- 分片文件 (`.rs.NNN`)：带有 31 字节轻量头部（magic: `RSBK`），支持从任意分片自动识别所属备份集合
- 存储开销大致与 `shares / threshold` 成正比；`shares - threshold` 越大，可容忍丢失的份数越多，但占用空间也越大

## License

[LGPL-2.1](LICENSE)
