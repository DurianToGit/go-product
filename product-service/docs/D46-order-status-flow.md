### 本节只定义三态：created / paid / cancelled
### 合法流转有哪些
合法：
created -> paid
created -> cancelled
非法：
paid -> cancelled
cancelled -> paid
### 为什么 paid 后不能 cancel、cancelled 后不能 pay
因为 paid 后，订单已经支付成功，不能再取消,此时取消订单，又涉及到补库存等动作，cancelled 后，订单已经取消，不能再支付。