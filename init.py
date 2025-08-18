from flask import Flask
from flask_cors import CORS
from flask_socketio import SocketIO
import os

app = Flask("Relay Server")
cors = CORS(app, supports_credentials=True)
socketio = SocketIO(app, cors_allowed_origins="*")