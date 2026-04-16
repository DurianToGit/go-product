### 为什么 Gin gateway 只做协议转换
 因为这是职责边界
 * ateway handler 不做业务
 * gateway handler 只做协议转换
 * 业务仍然在 product service 里
### 为什么 gateway 依赖 client 而不是 service
因为依赖service的话，你只是把 handler 挪了个地方，本质没变。
### 为什么这节先只改 product 读链路
读链路最稳，最适合先做网关化