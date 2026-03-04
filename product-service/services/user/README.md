## Service 目标（职责）
* 注册/登录/JWT

* 用户资料、用户状态

* （后续）用户地址、权限
## Owner 数据表（D37.1 里那套）
* users / user_profiles（你的项目里具体表名按现有来）
## 对外接口（列 3~5 条）
* POST /v1/auth/login

* POST /v1/auth/register

* GET /v1/users/me

* GET /v1/users/:id（可选）
## 依赖关系
无