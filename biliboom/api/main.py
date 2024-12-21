import time
import core
import sqlite3
import schedule
import os

# 创建数据库连接
conn = sqlite3.connect('bilibilcomments.db')
cursor = conn.cursor()
cursor.execute('''
    CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        bv TEXT,
        user TEXT,
        content TEXT,
        like_count INTEGER,
        comment_time TEXT,
        rpid TEXT UNIQUE
    )
''')
print(os.path.abspath('bilibilcomments.db'))  # 防止找不到文件打印一下路径

def insert_comment(comment_data):
    # 检查评论是否已经存在，如果不存在则插入
    #在 SQL 语句中使用 ? 作为占位符，表示该位置将会通过程序传入的实际数据来替代。
    #这是为了防止 SQL 注入攻击，并确保数据的正确性。
    cursor.execute('''
        INSERT OR IGNORE INTO comments (rpid, bv, user, content, like_count, comment_time)
        VALUES (?, ?, ?, ?, ?, ?)
    ''', (
        comment_data['rpid'], comment_data['bv'], comment_data['user'], comment_data['content'], 
        comment_data['like_count'], comment_data['comment_time']
    ))

    # 提交数据（更改数据库，如INSERT，DELECT等都要提交数据）
    conn.commit()

def update_comments():
    """
    这个函数将从 core.iter_comments() 获取评论数据并更新数据库
    """
    
    for comment in core.iter_comments():
        p = comment.get("parent_info")
        rpid = comment.get("rpid", "未找到 rpid")
        comment_time = comment.get("ctime", "未找到评论时间")
        like_count = comment.get("like", "未找到点赞数")

        # 格式化评论时间
        formatted_time = time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(comment_time)) if comment_time != 0 else "未找到评论时间"

        # 提取评论的必要信息
        comment_data = {
            'rpid': rpid,
            'bv': comment.get("bvid", "未找到 BV"),
            'user': comment.get("member", {}).get("uname", "未知用户"),
            'content': comment.get("content", {}).get("message", "没有评论内容"),
            'like_count': like_count,
            'comment_time': formatted_time
        }

        # 输出评论信息
        print(f'BV: {comment["bvid"]}, 用户：{comment["member"]["uname"]} 评论：{comment["content"]["message"]} {f"，回复 {p["content"]["message"]}" if p else ""}')
        print(f'评论ID: {rpid}, 评论时间: {formatted_time}, 点赞数: {like_count}')

        # 将评论保存到数据库
        insert_comment(comment_data)

def job():
    """
    每周运行一次的任务函数
    """
    print("更新评论数据...")
    update_comments()
    print("评论更新完成⁽⁽ ◟(∗ ˊωˋ ∗)◞ ⁾⁾！")

if __name__ == "__main__":
    # 使用 schedule 库来设置每周运行一次
    schedule.every(1).weeks.do(job)

    # 让程序一直运行，每次检查任务是否到期
    while True:
        schedule.run_pending() 
        time.sleep(1)  # 每秒检查一次任务
