from nonebot.command import CommandSession
from nonebot import SenderRoles
from nonebot.plugin import on_command

def normal_group_member(sender: SenderRoles):
    return sender.is_groupchat and not sender.is_admin and not sender.is_owner

@on_command("kickme")
async def kickme(session: CommandSession):
    sender = await SenderRoles.create(session.bot, session.event)
    if not normal_group_member(sender):
        return
    if session.ctx.group_id is None or session.event.user_id is None:
        return
    await session.bot.set_group_kick(group_id=session.ctx.group_id, user_id=session.event.user_id)
    await session.send("Kicked per request.")
