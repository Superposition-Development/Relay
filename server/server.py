from flask import Flask, request, jsonify, make_response, Response
from threading import Thread
from init import app, socketio
from flask_sqlalchemy import SQLAlchemy 
from flask_socketio import join_room
from sqlalchemy import and_
# from authToken import token_required
import jwt
from werkzeug.security import generate_password_hash, check_password_hash
from init import db, cors
import time
from model.user import Users, initUserTable
# from model.channel import Channel, initChannelTable
# from model.server import Servers, initServerTable
# from model.message import Message, initMessageTable
# from model.serverUser import ServerUser, initServerUser
import datetime
import json

@app.route('/')
def home():
    return "Relay's Server"

@app.route("/signup/", methods=["POST"]) 
def signup():
    data = request.get_json()
    hashed_password = generate_password_hash(data['password'], method='pbkdf2:sha256')
    user = Users(userID = data["userID"], password=hashed_password, username=data["username"]) 
    try:
        db.session.add(user)  
        db.session.commit()
    except:
        return jsonify({"error":"UserID taken!"})
    token = jwt.encode({'userID' : user.userID, 'exp' : datetime.datetime.utcnow() + datetime.timedelta(minutes=60)}, app.config['SECRET_KEY'], algorithm="HS256")
    response = {"jwt":token}
    json_data = json.dumps(response)
    resp = Response(json_data, content_type="application/json")
    resp.set_cookie("jwt", token,
                                max_age=3600,
                                secure=True,
                                httponly=True,
                                path='/',
                                samesite='None',  # cite
                                domain="172.27.233.236"
                                )
    return resp

@app.before_request
def before_request(): 
    allowed_origin = request.headers.get('Origin')
    if allowed_origin in ['http://localhost:4100', 'http://172.27.233.236:8080','https://spooketti.github.io']:
        cors._origins = allowed_origin
        
@socketio.on('join') # a room is used to choose who to send data to
def join(room):
    join_room(room)
        
@socketio.on('message')
def message(data):
    print(data)  # {'from': 'client'}
    socketio.emit('response', {'from': data["from"],
                      "message":data["message"]})


def run():
  socketio.run(app, host="0.0.0.0",port=6221,allow_unsafe_werkzeug=False)#bad idea to leave unsafe

initUserTable()
run()