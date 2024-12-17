import time
import requests

from config import SESSDATA


def iter_comments():
    """
    ### 获取创作中心评论列表接口
    请直接以 `for comment in iter_comments()` 的方式使用
    """

    def _get_comment_raw(page, size):
        url = "https://api.bilibili.com/x/v2/reply/up/fulllist"
        params = {
            "order": 1,
            "filter": -1,
            "type": 1,
            "bvid": "",
            "pn": page,
            "ps": size,
            "charge_plus_filter": "false",
        }

        headers = {
            "accept": "application/json, text/plain, */*",
            "accept-language": "zh-CN,zh;q=0.9",
            "cache-control": "no-cache",
            "origin": "https://member.bilibili.com",
            "pragma": "no-cache",
            "priority": "u=1, i",
            "referer": "https://member.bilibili.com/",
            "sec-ch-ua": '"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"',
            "sec-ch-ua-mobile": "?0",
            "sec-ch-ua-platform": '"Windows"',
            "sec-fetch-dest": "empty",
            "sec-fetch-mode": "cors",
            "sec-fetch-site": "same-site",
            "user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
        }

        # 直接把原始数据返回
        return requests.get(
            url, headers=headers, params=params, cookies={"SESSDATA": SESSDATA}
        ).json()

    count = 0
    page = 0
    size = 10
    count = 0
    while True:
        page += 1
        res = _get_comment_raw(page, size)
        count += res["data"]["page"]["size"]
        for comment in res["data"]["list"]:
            # 神奇的 yeild 可以让外面使用 for 循环调用这个函数
            # 直接吐出原始的 coment，具体的数据提取逻辑交给用户处理
            yield comment

        # 遍历完当然要退出
        if page > res["data"]["page"]["total"] / size:
            break
        # 超过一定数目也退出，保命要紧
        if count > 100:
            break
        time.sleep(5)
