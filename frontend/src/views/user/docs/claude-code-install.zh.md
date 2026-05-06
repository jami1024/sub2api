# Claude CLI / Claude Code 安装与 API 登录指南

> 适用系统：macOS、Linux、Windows  
> 适用工具：Claude Code CLI，也就是终端里的 `claude` 命令。

## 1. 准备工作

安装前请确认环境满足以下要求：

| 项目 | 要求 |
|---|---|
| macOS | macOS 10.15 或更高版本 |
| Linux | Ubuntu 20.04+ / Debian 10+，或兼容发行版 |
| Windows | Windows 10+，推荐使用 WSL；也可使用 Git Bash |
| Node.js | Node.js 18 或更高版本 |
| 内存 | 建议 4GB 以上 |
| 网络 | 需要能访问 Claude 或你的 API 服务地址 |

检查 Node.js 和 npm：

```bash
node -v
npm -v
```

如果没有安装 Node.js，建议先安装 LTS 版本。

---

## 2. 安装 Node.js

Claude CLI 需要 Node.js 18 或更高版本。推荐安装 Node.js LTS 版本。

安装完成后，请用下面的命令确认是否安装成功：

```bash
node -v
npm -v
```

如果能看到版本号，例如 `v20.x.x` 或 `v22.x.x`，就说明 Node.js 和 npm 已经可以正常使用。

### 2.1 macOS 安装 Node.js

#### 方式一：官网下载

1. 打开 [Node.js 官网](https://nodejs.org/)。
2. 下载 **LTS** 版本。
3. 双击 `.pkg` 安装包并按提示安装。
4. 安装完成后重新打开终端，执行：

```bash
node -v
npm -v
```

#### 方式二：使用 Homebrew

如果你已经安装 Homebrew，可以执行：

```bash
brew install node
```

验证安装：

```bash
node -v
npm -v
```

### 2.2 Linux 安装 Node.js

Linux 推荐使用 NodeSource 安装 Node.js LTS。下面以 Node.js 22 为例：

```bash
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt-get install -y nodejs
```

验证安装：

```bash
node -v
npm -v
```

如果你使用的不是 Ubuntu / Debian 系发行版，也可以从 [Node.js 官网](https://nodejs.org/) 下载对应系统的安装包，或使用系统自带的软件包管理器安装。

### 2.3 Windows 安装 Node.js

#### 方式一：官网下载

1. 打开 [Node.js 官网](https://nodejs.org/)。
2. 下载 Windows 版 **LTS** 安装包。
3. 双击 `.msi` 文件安装。
4. 安装完成后重新打开 PowerShell，执行：

```powershell
node -v
npm -v
```

#### 方式二：WSL 内安装

如果你使用 WSL，请在 WSL 的 Ubuntu 终端里安装 Node.js，而不是只在 Windows 里安装。

```bash
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt-get install -y nodejs
```

验证安装：

```bash
node -v
npm -v
```

> 提醒：如果在 WSL 中运行 `which node` 后路径出现在 `/mnt/c/` 下，说明调用的是 Windows 里的 Node.js。建议在 WSL 内重新安装 Node.js，避免后续权限或路径问题。

---

## 3. macOS 安装 Claude CLI

### 方式一：使用 npm 安装

```bash
npm install -g @anthropic-ai/claude-code
```

安装完成后检查版本：

```bash
claude --version
```

进入项目目录启动：

```bash
cd /path/to/your/project
claude
```

### 方式二：使用官方原生安装脚本

```bash
curl -fsSL https://claude.ai/install.sh | bash
```

安装后重新打开终端，再执行：

```bash
claude --version
```

---

## 4. Linux 安装 Claude CLI

### 方式一：使用 npm 安装

```bash
npm install -g @anthropic-ai/claude-code
```

验证安装：

```bash
claude --version
```

启动 Claude：

```bash
cd /path/to/your/project
claude
```

### 方式二：使用官方原生安装脚本

```bash
curl -fsSL https://claude.ai/install.sh | bash
```

如果安装后提示找不到 `claude`，请重新打开终端，或检查 `PATH` 是否包含安装目录。

---

## 5. Windows 安装 Claude CLI

Windows 推荐使用 WSL，也就是在 Windows 中运行 Linux 环境。

### 方式一：通过 WSL 安装（推荐）

先安装 WSL，例如 Ubuntu：

```powershell
wsl --install
```

安装完成后打开 Ubuntu 终端，检查 Node.js：

```bash
node -v
npm -v
```

然后安装 Claude CLI：

```bash
npm install -g @anthropic-ai/claude-code
```

验证安装：

```bash
claude --version
```

进入项目目录并启动：

```bash
cd /path/to/your/project
claude
```

### 方式二：Windows PowerShell 原生安装

也可以使用官方 PowerShell 安装脚本：

```powershell
irm https://claude.ai/install.ps1 | iex
```

安装完成后重新打开 PowerShell，检查：

```powershell
claude --version
```

### 方式三：使用 Git Bash

如果你使用 Git for Windows，可以在 Git Bash 中运行 Claude CLI。

如果 Claude 找不到 Git Bash，可以在 PowerShell 中指定 Git Bash 路径：

```powershell
$env:CLAUDE_CODE_GIT_BASH_PATH="C:\Program Files\Git\bin\bash.exe"
```

---

## 6. 使用 API Key 登录 / 认证

如果你使用的是第三方 API 网关或中转服务，例如企业代理、LiteLLM、AIGO 或其他 Anthropic 兼容服务，可以通过环境变量配置 API 登录。

> 注意：这里的“登录”不是网页登录，而是通过 API Key 进行认证。

通常需要同时设置两个环境变量：

| 变量名 | 作用 |
|---|---|
| `ANTHROPIC_BASE_URL` | API 服务地址 |
| `ANTHROPIC_AUTH_TOKEN` | API Key 或访问令牌 |

### macOS / Linux 临时配置

```bash
export ANTHROPIC_BASE_URL="https://your-api-host.example.com"
export ANTHROPIC_AUTH_TOKEN="YOUR_API_KEY"
claude
```

### macOS 永久配置（Zsh）

```bash
echo 'export ANTHROPIC_BASE_URL="https://your-api-host.example.com"' >> ~/.zshrc
echo 'export ANTHROPIC_AUTH_TOKEN="YOUR_API_KEY"' >> ~/.zshrc
source ~/.zshrc
```

### Linux 永久配置（Bash）

```bash
echo 'export ANTHROPIC_BASE_URL="https://your-api-host.example.com"' >> ~/.bashrc
echo 'export ANTHROPIC_AUTH_TOKEN="YOUR_API_KEY"' >> ~/.bashrc
source ~/.bashrc
```

### Windows PowerShell 临时配置

```powershell
$env:ANTHROPIC_BASE_URL="https://your-api-host.example.com"
$env:ANTHROPIC_AUTH_TOKEN="YOUR_API_KEY"
claude
```

### Windows PowerShell 永久配置

```powershell
[System.Environment]::SetEnvironmentVariable("ANTHROPIC_BASE_URL", "https://your-api-host.example.com", [System.EnvironmentVariableTarget]::User)
[System.Environment]::SetEnvironmentVariable("ANTHROPIC_AUTH_TOKEN", "YOUR_API_KEY", [System.EnvironmentVariableTarget]::User)
```

重新打开终端后生效。

### AIGO API 示例

如果你的 API 服务地址是 AIGO，可以这样配置。API Key 可登录 [AIGO 控制台](https://www.aigo.run) 后，在密钥或 API Key 管理页面创建并复制。

```bash
export ANTHROPIC_BASE_URL="https://www.aigo.run"
export ANTHROPIC_AUTH_TOKEN="YOUR_AIGO_API_KEY"
claude
```

Windows PowerShell：

```powershell
$env:ANTHROPIC_BASE_URL="https://www.aigo.run"
$env:ANTHROPIC_AUTH_TOKEN="YOUR_AIGO_API_KEY"
claude
```

---

## 7. 验证是否配置成功

### 检查 Claude CLI

```bash
claude --version
```

### 检查安装状态

```bash
claude doctor
```

### 启动测试

进入任意项目目录：

```bash
cd /path/to/your/project
claude
```

然后输入一个简单问题，例如：

```text
请帮我总结这个项目的目录结构
```

如果 Claude 能正常回复，说明安装和认证已经完成。

---

## 8. 常见问题

### 1. `claude: command not found` 怎么办？

通常是安装目录没有加入 `PATH`。

可以先检查：

```bash
which claude
```

如果没有输出，请重新打开终端，或检查 npm 全局安装目录。

### 2. npm 安装时权限不足怎么办？

不建议直接使用 `sudo npm install -g`。更推荐：

- 使用官方原生安装脚本；或
- 调整 npm 全局安装目录到用户目录；或
- 执行迁移命令：

```bash
claude migrate-installer
```

### 3. Windows WSL 里安装失败怎么办？

先确认 `node` 和 `npm` 来自 WSL 内部，而不是 Windows 路径：

```bash
which node
which npm
```

正常情况下路径应类似：

```text
/usr/bin/node
/usr/bin/npm
```

如果路径在 `/mnt/c/` 下，说明 WSL 正在调用 Windows 的 Node.js，建议在 WSL 内重新安装 Node.js。

### 4. API Key 配好了但还是 401 怎么办？

请检查：

- API Key 是否复制完整；
- API Key 是否过期或被删除；
- `ANTHROPIC_BASE_URL` 是否写错；
- 确认服务商要求使用的认证变量是否为 `ANTHROPIC_AUTH_TOKEN`；
- 如果服务商要求其他变量名，请以服务商文档为准。

### 5. 如何切换账号或 API Key？

网页登录方式：

```text
/logout
/login
```

环境变量方式：重新设置对应环境变量即可。

例如：

```bash
export ANTHROPIC_AUTH_TOKEN="NEW_API_KEY"
claude
```

---

## 9. 推荐使用流程

第一次安装建议按这个顺序：

1. 安装 Node.js 18+；
2. 安装 Claude CLI；
3. 执行 `claude --version`；
4. 执行 `claude doctor`；
5. 配置账号登录或 API Key；
6. 进入项目目录运行 `claude`；
7. 用一个简单问题测试是否能正常回复。

---

## 参考资料

- [Claude Code 官方安装文档](https://docs.anthropic.com/en/docs/claude-code/getting-started)
- [Claude Code 官方快速开始](https://docs.anthropic.com/en/docs/claude-code/quickstart)
- [Claude Code 设置与环境变量](https://docs.anthropic.com/en/docs/claude-code/settings)
- [Claude Code LLM Gateway 配置](https://docs.anthropic.com/en/docs/claude-code/llm-gateway)
- [Claude Code 故障排查](https://docs.anthropic.com/en/docs/claude-code/troubleshooting)
