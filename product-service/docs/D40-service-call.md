## 1. 当前为什么不直接做 HTTP/gRPC 调用

因为当前项目仍然是：

* 单 `module`
* 单入口
* 同进程运行

此时先抽 `client` 接口，能先解决服务间直接 `import` 的耦合问题。

## 2. 当前调用模式是什么

当前采用：

* productclient.Client

* userclient.Client

在 `bootstrap` 中注入本地实现 `LocalClient`，内部仍调用当前进程内的 `service`。

## 3. 这样做的价值

这样 `order-service` 不再直接依赖 `product-service` 的实现，而是依赖调用抽象。
后续如果切 `HTTP/gRPC`，只需替换 `client` 实现，不用改业务逻辑。

## 4. 后续如何演进

后续到真正独立服务阶段，可将：

* `LocalProductClient` 替换为 `HTTPProductClient` 或 `GRPCProductClient`

* `LocalUserClient` 替换为 `HTTPUserClient` 或 `GRPCUserClient`