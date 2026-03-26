## 什么是延迟队列
消息不是立刻被消费，而是要等到指定时间后才允许被处理。
## Kafka 为什么不原生支持延迟消息
Kafka 的核心能力是：

* 高吞吐写入
* 分区有序
* 消费组分摊
* offset 管理

但 Kafka **没有原生的“延迟到某个时间再投递”**能力。

也就是说，你不能像某些 MQ 一样直接说：
> delay = 5s

然后 Kafka 自动 5 秒后给你消费。

所以：

Kafka 做延迟，一般都要靠“应用层补一层调度逻辑”。
## Kafka 常见延迟实现方案有哪些
### 方案 1：多层 retry topic + 不同消费者延迟处理

#### 优点
* 简单
* 容易理解
* 很适合学习和中小项目

#### 缺点
* consumer 会被 sleep 占住
* 吞吐差
* 不适合大量延迟任务
### 方案 2：Kafka 作为存储，应用层调度器决定什么时候转发

思路是：

* 消息先进入一个 retry topic
* 消息体里带 next_retry_at
* 有一个 scheduler/dispatcher 消费 retry topic
* 如果时间还没到，不转发
* 时间到了，再投递到正式执行 topic
#### 优点
* 更工程化
* 不阻塞业务消费者
* 可以统一管理重试时间
#### 缺点
* 实现复杂
* 需要额外调度器
### 方案 3：Redis ZSet / 时间轮 做延迟层，Kafka 做正式消费层
思路：

*  重试失败后，不直接进 Kafka retry topic
*  先把消息放进 Redis ZSet
*  score = 下次执行时间戳
*  调度 worker 定时扫描 ZSet
*  到时间后，把消息重新投递到 Kafka 主 topic 或 retry topic
#### 优点
* Redis 非常适合按时间调度
* 你项目里本来就有 Redis
* 实现延迟很自然
* 比 Kafka 硬做延迟更顺手
缺点
* 架构变成 Redis + Kafka 双组件协作
* 需要额外扫描器
## Redis ZSet 为什么适合做延迟任务
Redis ZSet 擅长“定时调度”
## 在 go-product 里，为什么“Redis 延迟层 + MQ 正式消费”更合理
原因很简单：

* 项目里 Redis 已经是核心组件
* 本来就有 worker 架构
* Redis 做“到时间再取出”非常顺手
* 这也更符合很多实际业务系统的做法

也就是说：

> Kafka 擅长“消息流转”\
> Redis ZSet 擅长“定时调度”

这两个组合起来，比强行用 Kafka 单独扛延迟更合理。
## D44 核心设计图
```mermaid
业务服务
    ↓
写入延迟任务（Redis ZSet）
    ↓
delay-worker 定时扫描
    ↓
到期任务重新投递
    ↓
Kafka / Redis Stream 主消费链路
    ↓
真正执行业务
```