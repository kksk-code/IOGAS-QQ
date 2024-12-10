### synctoqq

#### 功能：

当 github 上发起一个 issue 时可以通过 qq 机器人发送到 qq 群中，并自动设置精华消息以及发布群公告

#### 流程：

github 发起 issue，程序接收 issue 的 json 数据，WebSocket 连接 qq 机器人，通过 api 发送消息，设置精华消息，以及发送群公告

#### 使用（待完成）

- 需要 linux 系统
- 需要配合 napcat 使用

参考资料：

- https://github.com/NapNeko/NapCat-Docker
- https://napneko.com/guide/boot/Shell

```bash
# 安装时选择使用 docker 安装，选择 ws，输入机器人 qq
curl -o napcat.sh https://nclatest.znin.net/NapNeko/NapCat-Installer/main/script/install.sh && sudo bash napcat.sh

# 扫描二维码登录，二维码显示在 docker 日志中，可以通过下面的命令查看
docker logs napcat
```

如何使 synctoqq 后台运行，可以使用 tmux

```bash
apt-get install tmux -y
tmux new -s synctoqq
# 在这个窗口中启动synctoqq
# 按住ctrl b + d 推出
```

（可选）内网穿透部分使用免费的 ngrok 内网穿透（公网 ip 或 frp 选手请绕路）
```bash
ngrok http http://localhost:8080
```

#### 感谢：

1. github webhook: https://docs.github.com/zh/webhooks
2. onebot（统一的聊天机器人应用接口标准）: https://onebot.dev/
3. napcat(qq 的 onebot 实现): https://napneko.com/
4. go-cqhttp(基于 Mirai 以及 MiraiGo 的 OneBot Golang 原生实现) : https://docs.go-cqhttp.org/api/
