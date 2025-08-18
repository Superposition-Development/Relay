from flask import Flask, request, jsonify, make_response, Response
from threading import Thread
from init import app, socketio
from flask_socketio import join_room, emit
from init import cors
import time
import datetime
import json


@app.route('/')
def home():
    return "Relay's Server"

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

run()