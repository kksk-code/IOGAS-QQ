### synctoqq  

#### 实现的功能：  

当github上发起一个issue时可以通过qq机器人发送到qq群中，并自动设置精华消息以及发布群公告  

#### 实现流程：  

github发起issue ，程序接收issue的json数据， webshock连接qq机器人， 通过api发送消息，设置精华消息，以及发送群公告  

#### 感谢：
1. github webhook: https://docs.github.com/zh/webhooks
2. onebot（统一的聊天机器人应用接口标准）: https://onebot.dev/
3. napcat(qq 的 onebot 实现): https://napneko.com/
4. go-cqhttp(基于 Mirai 以及 MiraiGo 的 OneBot Golang 原生实现) : https://docs.go-cqhttp.org/api/