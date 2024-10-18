# ai-api-proxy

**ai-api-proxy** 是一个高性能、可扩展的 API 代理服务器，旨在为国内用户提供稳定、快速的 OpenAI 和 Anthropic 等外部 API 服务的直连访问。通过反向代理、限流、安全认证等功能，**ai-api-proxy** 简化了与外部 API 的交互，提升了安全性和系统稳定性，适用于各种规模的应用场景。

## 目录

- [功能特点](#功能特点)
- [技术栈](#技术栈)
- [安装与运行](#安装与运行)
  - [前置条件](#前置条件)
  - [使用 Docker 部署](#使用-docker-部署)
  - [源码运行](#源码运行)
- [配置说明](#配置说明)
- [使用指南](#使用指南)
  - [API 认证](#api-认证)
  - [限流设置](#限流设置)
  - [请求示例](#请求示例)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
  - [环境搭建](#环境搭建)
  - [代码规范](#代码规范)
  - [测试](#测试)
  - [贡献指南](#贡献指南)
- [部署与运维](#部署与运维)
  - [持续集成/持续部署 (CI/CD)](#持续集成持续部署-cicd)
  - [容器化与编排](#容器化与编排)
- [安全性](#安全性)
- [常见问题](#常见问题)
- [许可证](#许可证)
- [致谢](#致谢)

## 功能特点

- **多 API 支持**：支持 OpenAI、Anthropic 等多种外部 API 的代理。
- **高性能反向代理**：基于 Go 语言的高效反向代理实现，确保低延迟和高吞吐量。
- **安全认证**：支持 API 密钥认证，确保只有授权用户才能访问代理服务。
- **限流机制**：内置灵活的限流策略，防止滥用和保护后端服务。
- **请求体限制**：控制请求体的最大大小，防止过大请求影响服务稳定性。
- **日志管理**：集成日志记录功能，支持日志文件和标准输出记录。
- **自动化部署**：提供 Dockerfile，简化部署流程，支持容器化部署。
- **健康检查**：内置健康检查接口，方便监控和运维。
- **配置灵活**：通过配置文件轻松调整服务参数，适应不同的使用场景。

## 技术栈

- **编程语言**：Go 1.23.2
- **Web 框架**：Go 标准库 `net/http`，Gin
- **限流库**：[ulule/limiter](https://github.com/ulule/limiter)
- **日志库**：Logrus
- **配置管理**：Viper
- **容器化**：Docker
- **持续集成**：GitHub Actions

## 安装与运行

### 前置条件

- [Docker](https://www.docker.com/get-started)（如果选择使用 Docker 部署）
- [Go 1.23.2](https://golang.org/dl/)（如果选择源码运行）
- Git

### 使用 Docker 部署

1. **克隆仓库**

   ```shell
   git clone https://github.com/你的用户名/ai-api-proxy.git
   cd ai-api-proxy
   ```

2. **配置 Docker 环境**

   确保 `config.yaml` 已正确配置。可以参考以下示例：

   ```yaml
   server_port: "3001"
   rate_limit: "100-M"
   fixed_request_ip: ""
   max_request_body_size_mb: 50
   log_dir: "logs"
   log_name: "app.log"
   log_level: "info"
   path_map: {
     "/openai/": "https://api.openai.com",
     "/anthropic/": "https://api.anthropic.com",
   }
   ```

3. **构建并运行 Docker 容器**

   ```shell
   docker build -t ai-api-proxy:latest .
   docker run -d --restart=always \
     --name ai-api-proxy \
     -v $(pwd)/config.yaml:/app/config.yaml \
     -v $(pwd)/logs:/app/logs \
     --network my-network \
     ai-api-proxy:latest
   ```

### 源码运行

1. **克隆仓库**

   ```shell
   git clone https://github.com/你的用户名/ai-api-proxy.git
   cd ai-api-proxy
   ```

2. **安装依赖**

   ```shell
   go mod download
   ```

3. **配置应用**

   编辑 `config.yaml`，参考上文 Docker 部署中的示例。

4. **运行应用**

   ```shell
   go run ./cmd/proxy
   ```

## 配置说明

`config.yaml` 是应用的主要配置文件，包含以下字段：

- `server_port`: 服务器监听的端口号（默认 `"3001"`）。
- `rate_limit`: 请求限流策略，格式为 `"100-M"` 表示每分钟最多100次请求。
- `fixed_request_ip`: 固定请求 IP，用于隐藏原始客户端 IP（默认为空，表示不固定）。
- `max_request_body_size_mb`: 请求体最大允许大小，单位为 MB（默认 `50`）。
- `log_dir`: 日志文件存放目录（默认 `"logs"`）。
- `log_name`: 日志文件名称（默认 `"app.log"`）。
- `log_level`: 日志记录级别（支持 `debug`, `info`, `warn`, `error`，默认 `"info"`）。
- `path_map`: 路径映射，定义了代理的前缀路径与目标 API 的对应关系。

示例：

```yaml
server_port: "3001"
rate_limit: "100-M"
fixed_request_ip: ""
max_request_body_size_mb: 50
log_dir: "logs"
log_name: "app.log"
log_level: "info"
path_map: {
  "/openai/": "https://api.openai.com",
  "/anthropic/": "https://api.anthropic.com",
}
```

## 使用指南

### API 认证

所有请求必须包含有效的 API 密钥，可以通过 `Authorization` 或 `x-api-key` 头进行传递。

示例：

```http
GET /openai/v1/engines HTTP/1.1
Host: your-domain.com
Authorization: your-api-key
```

### 限流设置

应用内置了限流机制，默认每分钟最多允许 100 次请求。可在 `config.yaml` 中通过 `rate_limit` 字段进行调整。

示例：

```yaml
rate_limit: "200-M" # 每分钟最多200次请求
```

### 请求示例

假设你已经配置好 `path_map` 并运行了代理服务器，以下是如何通过代理访问 OpenAI API 的示例：

```shell
curl -H "Authorization: your-api-key" https://your-domain.com/openai/v1/engines
```

## 项目结构

```plaintext
ai-api-proxy/
├── cmd/
│   └── proxy/
│       └── main.go           # 应用入口
├── internal/
│   ├── config/
│   │   ├── config.go         # 配置加载
│   │   └── config_test.go    # 配置测试
│   ├── middleware/
│   │   └── middleware.go     # 中间件实现
│   └── proxy/
│       ├── reverse_proxy.go  # 反向代理逻辑
│       └── reverse_proxy_test.go # 反向代理测试
├── pkg/
│   └── logger/
│       └── logger.go         # 日志初始化
├── config.yaml               # 配置文件
├── Dockerfile                # Docker 构建文件
├── entrypoint.sh             # Docker 入口脚本
├── Makefile                  # 构建脚本
├── go.mod
├── go.sum
├── README.md                 # 项目说明
├── .gitignore
├── .dockerignore
└── .github/
    └── workflows/
        └── docker-image.yml  # GitHub Actions CI 配置
```

## 开发指南

### 环境搭建

1. **安装 Go**

   下载并安装 Go 1.23.2 或更高版本。参考 [Go 官方安装指南](https://golang.org/doc/install)。

2. **克隆仓库**

   ```shell
   git clone https://github.com/你的用户名/ai-api-proxy.git
   cd ai-api-proxy
   ```

3. **安装依赖**

   ```shell
   go mod download
   ```

### 代码规范

- 遵循 [Effective Go](https://golang.org/doc/effective_go) 规范。
- 代码风格统一，建议使用 `gofmt` 格式化代码。
- 使用有意义的变量和函数命名，注重代码可读性。
- 遵循 Go 的最佳实践，避免常见的陷阱和反模式。
- 编写充分的注释，特别是在复杂或不直观的逻辑部分。

### 测试

1. **运行所有测试**

   ```shell
   go test ./...
   ```

2. **运行特定测试**

   ```shell
   go test ./internal/proxy -run TestNewOpenAIReverseProxy_InvalidURL
   ```

3. **测试覆盖率**

   ```shell
   go test ./... -cover
   ```

### 贡献指南

欢迎各类贡献，无论是提出问题、报告 bug、建议新功能，还是提交代码。请按照以下步骤进行：

1. **Fork 本仓库**

   在 GitHub 上点击 "Fork" 按钮，将仓库复制到你的账户下。

2. **创建新分支**

   ```shell
   git checkout -b feature/你的功能
   ```

3. **提交更改**

   ```shell
   git commit -m "添加了新的功能"
   ```

4. **推送到远程分支**

   ```shell
   git push origin feature/你的功能
   ```

5. **创建 Pull Request**

   在 GitHub 上发起 Pull Request，描述你的更改内容和原因。

请确保所有新功能都配有相应的测试，并且通过现有测试。我们会对 Pull Request 进行审查，并在必要时提供反馈。

## 部署与运维

### 持续集成/持续部署 (CI/CD)

本项目使用 GitHub Actions 进行持续集成和持续部署。CI/CD 流程包括：

- **代码检出**：使用 `actions/checkout` 检出代码。
- **构建 Docker 镜像**：使用 Dockerfile 构建镜像。
- **运行测试**：执行所有单元测试，确保代码质量。
- **推送镜像**：将构建好的镜像推送到 Docker Hub 或其他镜像仓库。

你可以在 `.github/workflows/docker-image.yml` 中查看和自定义 CI/CD 配置。

### 容器化与编排

项目提供了 `Dockerfile` 以支持容器化部署。你可以使用 Docker Compose 或 Kubernetes 进行容器编排，进一步提升系统的可伸缩性和可靠性。

**示例 Kubernetes 部署配置：**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ai-api-proxy
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ai-api-proxy
  template:
    metadata:
      labels:
        app: ai-api-proxy
    spec:
      containers:
      - name: ai-api-proxy
        image: your-docker-username/ai-api-proxy:latest
        ports:
        - containerPort: 3001
        volumeMounts:
        - name: config
          mountPath: /app/config.yaml
          subPath: config.yaml
        - name: logs
          mountPath: /app/logs
      volumes:
      - name: config
        configMap:
          name: ai-api-proxy-config
      - name: logs
        persistentVolumeClaim:
          claimName: ai-api-proxy-logs
---
apiVersion: v1
kind: Service
metadata:
  name: ai-api-proxy-service
spec:
  type: LoadBalancer
  selector:
    app: ai-api-proxy
  ports:
    - protocol: TCP
      port: 80
      targetPort: 3001
```

## 安全性

- **认证与授权**：仅允许拥有有效 API 密钥的用户访问代理服务，确保服务的安全性。
- **输入验证**：对所有输入进行严格验证，防止注入攻击和其他恶意请求。
- **日志监控**：记录所有请求和错误，便于监控和审计。
- **依赖管理**：定期更新依赖库，修复已知的安全漏洞。
- **HTTPS 支持**：建议在生产环境中使用 HTTPS 保障数据传输的安全性。

## 常见问题

### 如何修改 API 代理的目标地址？

编辑 `config.yaml` 中的 `path_map` 字段，添加或修改路径前缀与目标 API 地址的映射关系。例如：

```yaml
path_map: {
  "/openai/": "https://api.openai.com",
  "/anthropic/": "https://api.anthropic.com",
  "/newapi/": "https://api.newservice.com",
}
```

### 如何增加新的限流策略？

在 `config.yaml` 中修改 `rate_limit` 字段，按照 **ulule/limiter** 的格式进行配置。例如，将限流调整为每分钟200次请求：

```yaml
rate_limit: "200-M"
```

### 日志文件在哪里？

默认情况下，日志文件保存在项目根目录下的 `logs` 文件夹中，文件名为 `app.log`。你可以在 `config.yaml` 中通过 `log_dir` 和 `log_name` 字段自定义日志路径和文件名。

## 许可证

本项目采用 [MIT 许可证](LICENSE) 进行许可。详情请参阅 [LICENSE](LICENSE) 文件。

## 致谢

感谢所有为 **ai-api-proxy** 项目贡献代码、报告问题和提出建议的开发者和用户。您的支持是我们不断前进的动力！

---
如果您在使用过程中有任何问题或建议，欢迎提交 [Issue](https://github.com/你的用户名/ai-api-proxy/issues) 或参与讨论。
