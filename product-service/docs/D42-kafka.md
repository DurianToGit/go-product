
## 任务 1：整理 Kafka 概念笔记

用自己的话写清楚：

* Producer：他是生产者，代表需要处理的数据，在go-product中，他可以是待处理的库存扣减任务，库存恢复任务
* Topic：消息类别，就像是不同的任务组，在go-product中，比如一个库存扣减任务就可以创建一个Topic
* Partition：消息类别的分区，提高并发用的，kafka可以保证分区顺序一致，但是不能让topic的顺序一致，在go-product中，两个库存扣减任务，他是可能被分配到不同的partition中去的
* Consumer：消费者，主要是负责读取生产者的消息，在go-product中，可以是库存worker,通知worker，订单worker等
* Consumer Group：
> 这个是 Kafka 比较有代表性的设计。\
同一个 group 内：\
一个 partition 同一时刻只会被一个 consumer 消费 \
不同 group 之间： \
可以各自独立消费同一份消息 \
用你的项目举例 \
比如 order.paid 这个 Topic： \
一个 inventory-group 消费它，用来处理库存确认 \
另一个 notify-group 消费它，用来给用户发通知 \
同一条消息，可以被多个系统各自消费一次。 \
> 这就是典型的事件驱动思维。 

* Offset: Offset 是消息在 partition 中的位置。

  Kafka 消费本质上就是：

  * 拉取消息
  * 处理消息
  * 提交 offset

  也就是告诉 Kafka：
  
  我已经消费到这里了，下次从后面继续。
* Replica：副本，是为了broker挂掉时，副本还能顶上继续使用，提高kafka的稳定性的

* Kafka 和 Redis Stream 的区别

  * Redis Stream 更适合

    * 轻量异步任务
    * 已有 Redis 技术栈
    * 快速接入
    * 中小型项目

  * Kafka 更适合

    * 高吞吐
    * 海量消息堆积
    * 多系统订阅同一事件
    * 更复杂的事件驱动架构
    * 更成熟的分区扩展与副本机制


任务 2：设计 go-product 的 Kafka Topic 草案

* order.created
* order.cancelled
* order.paid
* stock.deduct.requested
* stock.restore.requested

每个 Topic 写清楚：

生产者是谁

消费者是谁

key 用什么

payload 包含哪些字段

是否要求顺序

任务 3：设计一层 MQ 抽象接口
```go
  type EventBus interface {
      Publish(ctx context.Context, topic string, key string, payload []byte) error
  }
```