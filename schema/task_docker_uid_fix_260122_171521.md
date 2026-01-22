# Task: Docker容器UID不匹配问题修复

**任务ID**: task_docker_uid_fix_260122_171521
**创建时间**: 2026-01-22
**状态**: 进行中
**目标**: 解决宿主机UID 1000与容器内UID 1001不匹配导致的证书文件读取问题

## 最终目标
1. 分析当前Dockerfile和docker-compose.yml配置
2. 评估三种解决方案的适用性
3. 执行选定的方案，解决证书文件读取权限问题
4. 验证修复效果

## 拆解步骤
### 1. 分析当前配置
- [x] 查看Dockerfile中appuser的UID设置：UID=1001, GID=1001
- [x] 查看docker-compose.yml中user映射配置：第26行注释了user映射
- [ ] 检查宿主机证书文件权限和所有者

### 2. 方案评估与决策
- [x] 方案A: 修改Dockerfile，将appuser uid从1001改为1000（需要重新构建镜像）
- [x] 方案B: 启用docker-compose.yml第26行的user映射（推荐：快速、安全、灵活）
- [x] 方案C: 修改宿主机证书权限chmod -R a+r（安全性最低）
- [x] 推荐最优方案并获取用户确认：**方案B**

### 3. 执行修复
- [x] 根据选定方案实施修改：已启用docker-compose.yml第26行的user映射
- [x] 分析Dockerfile用户创建的影响：保留作为安全回退机制
- [ ] 调整Dockerfile权限设置（如果需要）：待测试后决定
- [ ] 测试容器启动和证书读取

### 4. 验证与测试
- [ ] 构建/启动容器测试权限问题是否解决
- [ ] 验证应用功能正常

## 当前进度
### 正在进行: 验证修复效果与配置意义分析
已启用docker-compose.yml的user映射，正在分析Dockerfile中UID=1001配置的实际意义和价值

## 讨论点
1. **Dockerfile中UID=1001的意义**：虽然运行时被docker-compose的user映射覆盖，但作为安全回退机制
2. **文件所有权冲突**：`chown -R appuser:appgroup /app`导致文件属于UID=1001，但进程以UID=501运行
3. **多环境适配**：`user: "${UID:-1000}:${GID:-1000}"`自动适配Mac(501)和AWS(1000)环境

## 下一步行动
1. 启动容器测试权限问题是否解决
2. 如果需要，调整Dockerfile权限设置（chmod替代chown）
3. 验证应用功能正常