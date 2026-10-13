# Go 语言学习示例

这是一个基于 cex-exchange 交易所项目整理的 Go 学习示例集合。

## 目录

| 文件 | 主题 | 说明 |
|------|------|------|
| [01_struct_and_method.go](01_struct_and_method.go) | 结构体和方法 | Go 最基本的数据结构 |
| [02_interface.go](02_interface.go) | 接口和依赖注入 | 解耦和面向接口编程 |
| [03_context.go](03_context.go) | Context 用法 | 请求上下文和超时控制 |
| [04_concurrency.go](04_concurrency.go) | 并发安全 | sync.Mutex 和 sync.WaitGroup |
| [05_error_handling.go](05_error_handling.go) | 错误处理 | Go 的显式错误处理模式 |
| [06_json.go](06_json.go) | JSON 处理 | 序列化/反序列化 |
| [07_cache.go](07_cache.go) | 缓存模式 | Redis 缓存实现 |
| [08_string.go](08_string.go) | 字符串处理 | 常用字符串操作 |
| [09_sort.go](09_sort.go) | 排序和查找 | 订单簿排序和聚合 |
| [10_modules.go](10_modules.go) | 模块化和项目结构 | Go 项目最佳实践 |

## 运行示例

```bash
cd examples/go_learn

# 运行单个示例
go run 01_struct_and_method.go

# 运行所有示例
for f in *.go; do
    echo "=== 运行 $f ==="
    go run "$f"
    echo ""
done
```

## 与项目代码的对应关系

这些示例基于 `cex-exchange` 项目中的实际代码：

| 示例 | 项目中对应的代码 |
|------|----------------|
| 结构体和方法 | [internal/engine/types.go](https://github.com/example/cex-exchange/blob/main/internal/engine/types.go) |
| 接口设计 | [internal/dao/user.go](https://github.com/example/cex-exchange/blob/main/internal/dao/user.go) |
| 缓存模式 | [internal/service/order.go](https://github.com/example/cex-exchange/blob/main/internal/service/order.go) |
| 并发安全 | [internal/engine/matching.go](https://github.com/example/cex-exchange/blob/main/internal/engine/matching.go) |
| 错误处理 | [internal/service/order.go](https://github.com/example/cex-exchange/blob/main/internal/service/order.go) |

## 学习路径建议

1. **入门**: 结构体和方法 → 字符串处理 → JSON 处理
2. **进阶**: 接口和依赖注入 → 错误处理 → 缓存模式
3. **高级**: Context 用法 → 并发安全 → 排序和查找
4. **工程化**: 模块化和项目结构

## 参考资料

- [Go 官方文档](https://go.dev/doc/)
- [Go 高级实战课](https://github.com/chai2010/advanced-go-programming-book)
- [Go Frame 框架文档](https://goframe.org/)