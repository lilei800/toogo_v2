## HotGo 管理后台服务端

### 测试与健康检查

- **快速健康检查（无需 DB/Redis）**：在 `hotgo_v2/server` 执行 `go test ./...`
- **集成测试（需要 DB/Redis）**：见 `INTEGRATION_TESTS.md`