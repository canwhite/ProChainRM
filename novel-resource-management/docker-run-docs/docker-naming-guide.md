# Docker镜像和容器命名详解

## 1. 为什么不需要定义image名称？

**Docker Compose的自动机制：**

当你写：
```yaml
services:
  novel-api:
    build:
      context: .
      dockerfile: Dockerfile
```

Docker Compose会自动做这些事：
1. **自动生成镜像名称**：`novel-resource-management_novel-api`
2. **自动构建镜像**：如果镜像不存在，会根据build配置构建
3. **自动创建容器**：使用生成的镜像创建容器

**完整的自动命名规则：**
- 镜像名称：`项目目录名_服务名` → `novel-resource-management_novel-api`
- 容器名称：由`container_name`指定 → `novel-api`

## 2. 手动指定镜像名称（可选）

如果你想自定义镜像名称，可以这样：
```yaml
services:
  novel-api:
    build:
      context: .
      dockerfile: Dockerfile
    image: my-custom-novel-api:latest  # ← 手动指定镜像名称
    container_name: novel-api
```

## 3. Pull和Push操作

**当前配置下的操作：**

```bash
# 查看自动生成的镜像名称
docker images | grep novel-api

# 重新打tag为你想要的名称
docker tag novel-resource-management_novel-api:latest your-username/novel-api:v1.0

# 推送到镜像仓库
docker push your-username/novel-api:v1.0

# 从镜像仓库拉取
docker pull your-username/novel-api:v1.0
```

**修改配置使用外部镜像：**
```yaml
services:
  novel-api:
    image: your-username/novel-api:v1.0  # ← 使用外部镜像
    container_name: novel-api
    build:  # ← 可以保留build，也可以删除
      context: .
      dockerfile: Dockerfile
```

## 4. 实际操作演示

查看当前项目中的实际情况：

```bash
# 查看自动生成的镜像名称
docker images | grep novel

# 输出示例：
# novel-resource-management-novel-api   latest   e3142f4a7f00   4 hours ago   89.6MB
```

注意：Docker Compose实际生成的镜像名称是 `novel-resource-management-novel-api`（用连字符代替下划线）。

## 5. 完整的Pull/Push工作流程

**如果你想推送到Docker Hub：**

```bash
# 1. 重新打tag
docker tag novel-resource-management-novel-api:latest your-dockerhub-username/novel-api:v1.0

# 2. 登录Docker Hub
docker login

# 3. 推送
docker push your-dockerhub-username/novel-api:v1.0

# 4. 其他地方拉取
docker pull your-dockerhub-username/novel-api:v1.0

# 5. 修改docker-compose.yml使用外部镜像
# image: your-dockerhub-username/novel-api:v1.0
```

## 📝 核心概念总结

- **镜像名称**：Docker Compose自动生成，也可以手动指定
- **容器名称**：通过`container_name`指定
- **Pull/Push的对象**：是镜像（image），不是容器
- **服务名称**：只是Docker Compose内部的逻辑标识符

记住：**Pull和Push的都是镜像（Image），不是容器（Container）！**