from nonebot import on_command, CommandSession
from nonebot import on_natural_language, NLPSession, IntentCommand
from jieba import posseg
import sqlite3
import re
from datetime import datetime
import os

from nonebot.default_config import SUPERUSERS

from modules import IO_QID


# 确保数据库文件存在
DB_PATH = 'seminars.db'
if not os.path.exists(DB_PATH):
    open(DB_PATH, 'a').close()

# 连接到SQLite数据库
conn = sqlite3.connect(DB_PATH)
cursor = conn.cursor()

# 创建新的表结构
cursor.executescript('''
    CREATE TABLE IF NOT EXISTS speaker (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT UNIQUE
    );

    CREATE TABLE IF NOT EXISTS topic (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT UNIQUE
    );

    CREATE TABLE IF NOT EXISTS speaker_topic (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        speaker_id INTEGER,
        topic_id INTEGER,
        FOREIGN KEY (speaker_id) REFERENCES speaker (id),
        FOREIGN KEY (topic_id) REFERENCES topic (id),
        UNIQUE(speaker_id, topic_id)
    );

    CREATE TABLE IF NOT EXISTS event (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date TEXT,
        time TEXT,
        name TEXT,
        speaker_topic_id INTEGER,
        FOREIGN KEY (speaker_topic_id) REFERENCES speaker_topic (id)
    );
''')
conn.commit()

@on_command('schedule_seminar', aliases=('预订', '安排'))
async def schedule_seminar(session: CommandSession):
    if session.event.group_id != IO_QID:
        return

    # 获取解析后的参数
    date = session.state.get('date')
    time = session.state.get('time')
    speakers = session.state.get('speakers')

    # 再次检查信息完整性
    while not (date and time and speakers):
        missing_info = []

        if not date:
            missing_info.append("日期")
        if not time:
            missing_info.append("时间")
        if not speakers:
            missing_info.append("分享者")

        prompt = '请提供以下信息：' + '、'.join(missing_info)

        seminar_info = await session.aget(prompt=prompt)
        parsed_date, parsed_time, parsed_speakers = parse_seminar_info_from_words(posseg.lcut(seminar_info))
        date = parsed_date if parsed_date else date
        time = parsed_time if parsed_time else time
        speakers = parsed_speakers if parsed_speakers else speakers

    # 存储到数据库
    for speaker_name, topic_name in speakers.items():
        # 插入或获取 speaker
        cursor.execute('INSERT OR IGNORE INTO speaker (name) VALUES (?)', (speaker_name,))
        cursor.execute('SELECT id FROM speaker WHERE name = ?', (speaker_name,))
        speaker_id = cursor.fetchone()[0]

        # 插入或获取 topic
        cursor.execute('INSERT OR IGNORE INTO topic (name) VALUES (?)', (topic_name,))
        cursor.execute('SELECT id FROM topic WHERE name = ?', (topic_name,))
        topic_id = cursor.fetchone()[0]

        # 插入或获取 speaker_topic
        cursor.execute('INSERT OR IGNORE INTO speaker_topic (speaker_id, topic_id) VALUES (?, ?)',
                       (speaker_id, topic_id))
        cursor.execute('SELECT id FROM speaker_topic WHERE speaker_id = ? AND topic_id = ?',
                       (speaker_id, topic_id))
        speaker_topic_id = cursor.fetchone()[0]

        # 插入 event
        cursor.execute('INSERT INTO event (date, time, name, speaker_topic_id) VALUES (?, ?, ?, ?)',
                       (date, time, "技术茶话会", speaker_topic_id))

    conn.commit()

    # 构建回复消息
    speakers_str = ', '.join([f"{name}({topic})" for name, topic in speakers.items()])
    response = f"Seminar scheduled.\nDate: {date}\nTime: {time}\nTopic: 技术茶话会\nSpeakers: {speakers_str}"
    await session.send(response)

@on_natural_language(keywords={'预订', '安排', '茶话会'})
async def _(session: NLPSession):
    if session.event.group_id != IO_QID and session.event.user_id not in SUPERUSERS:
        return
    # 去掉消息首尾的空白符
    stripped_msg = session.msg_text.strip().replace('\n', '，')
    # 对消息进行分词和词性标注
    words = posseg.lcut(stripped_msg)

    date, time, speakers = parse_seminar_info_from_words(words)

    # 构建命令参数
    args = {
        'date': date,
        'time': time,
        'speakers': speakers
    }

    # 返回意图命令，使用正确的参数顺序
    return IntentCommand(60.0, 'schedule_seminar', args=args, current_arg=stripped_msg)

def parse_seminar_info_from_words(words):
    time_keywords = {
        '早上': 'AM',
        '上午': 'AM',
        '中午': 'PM',
        '下午': 'PM',
        '晚上': 'PM',
        '晚': 'PM'
    }

    date = time = None
    speakers = {}
    date_words = []
    other_words = []

    for word, flag in words:
        if flag in ['m', 't']:
            date_words.append(word)
        else:
            other_words.append(word)

    # 解析日期
    date_str = ''.join(date_words)
    date_match = re.search(r'(\d+)月(\d+)[号日]', date_str)
    if date_match:
        month, day = date_match.groups()
        date = f"{datetime.now().year}-{month.zfill(2)}-{day.zfill(2)}"

    # 解析时间
    time_match = re.search(r'(\d+)\s*[点时]', date_str)
    if time_match:
        hour = int(time_match.group(1))
        for keyword, period in time_keywords.items():
            if keyword in date_str:
                if period == 'PM' and hour < 12:
                    hour += 12
                break
        time = f"{hour:02d}:00"

    # 解析分享者和分享内容
    full_text = ''.join(other_words)
    speaker_matches = re.findall(r'(\w+)(?:：|:|分享)\s*([^，。\s]+(?:\s+[^，。\s]+)*)', full_text)

    for name, content in speaker_matches:
        speakers[name] = content.strip()

    return date, time, speakers

# 在程序结束时关闭数据库连接
# conn.close()
