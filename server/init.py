from flask import Flask
from flask_cors import CORS
from flask_socketio import SocketIO
from flask_sqlalchemy import SQLAlchemy
import os

app = Flask("Relay Server")
cors = CORS(app, supports_credentials=True)
dbURI = 'sqlite:///sqlite.db' #fix this to only create the script in the server folder
app.config['SQLALCHEMY_DATABASE_URI'] = dbURI
app.config['SECRET_KEY'] = "test" #os.getenv("SECRET_KEY")
db = SQLAlchemy(app)    
socketio = SocketIO(app, cors_allowed_origins="*")