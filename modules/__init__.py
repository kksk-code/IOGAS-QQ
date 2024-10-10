import os
from dotenv import load_dotenv


load_dotenv()

IO_QID = int(os.getenv("IO_QID"))
if not IO_QID:
    raise ValueError("Environment variables IO_QID is missing.")

SUPERUSERS = []
super_user = os.getenv("SUPERUSER")
if super_user is not None:
    SUPERUSERS.append(super_user)
