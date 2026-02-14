from dotenv import load_dotenv
import os
load_dotenv()
SERVER_NAME = "Relay Server"
DATABASE_NAME = os.getenv("DATABASE_NAME")