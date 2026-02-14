from flask import Flask
import sqlite3
import init
import pathlib

app = Flask(init.SERVER_NAME)

print(pathlib.Path(__file__).parent)

connection = sqlite3.connect(f"./{init.DATABASE_NAME}.db")

@app.route("/")
def home():
    return "eventually lets put something cool here"


app.run()