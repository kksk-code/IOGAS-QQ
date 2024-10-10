from nonebot.default_config import *
import os
from dotenv import load_dotenv


load_dotenv()

SUPERUSERS = []
super_user = os.getenv("SUPERUSER")
if super_user is not None:
    SUPERUSERS.append(super_user)
