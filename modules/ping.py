from nonebot.command import CommandSession
from nonebot.plugin import on_command


@on_command('ping')
async def ping(session: CommandSession):
    print(session.ctx.discuss_id)
    await session.send("ACK")
