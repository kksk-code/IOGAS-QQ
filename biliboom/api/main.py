import time
import core

if __name__ == "__main__":
    for comment in core.iter_comments():
        p = comment.get("parent_info")
        print(
            f'BV: {comment['bvid']}, 用户：{comment['member']['uname']} 评论：{comment['content']['message']} {f'，回复 p["content"]["message"]' if p else ""}'
        )
