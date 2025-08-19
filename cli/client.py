# @click.command()
# @click.option('--count', default=1, help='Number of greetings.')
# @click.option('--name', prompt='Your name',
#               help='The person to greet.')
# def hello(count, name):
#     """Simple program that greets NAME for a total of COUNT times."""
#     for x in range(count):
#         click.echo(click.style(f"Hello {name}!",fg="blue    "))

import click
import socketio
import os
from enum import Enum

localCache = []

class FSM(Enum):
    BOOT=0
    CHANNEL=1
    SIGNUP=2
    LOGIN=3

fsmState = FSM.BOOT

sio = socketio.Client()
sio.connect('http://127.0.0.1:6221')

username = ""
isChatActive = False

@sio.on('response')
def response(data):
  localCache.append(data)
  click.clear()
  draw()
  draw()

@click.group(invoke_without_command=True)
@click.option("--command", prompt=">")
@click.pass_context
def cli(ctx, command):
    cmd = cli.get_command(ctx, command)
    ctx.invoke(cmd)

@cli.command(name='username')
def username():
    global username
    username = click.prompt('Username', type=click.STRING)

@cli.command(name='message')
def message():
    global username
    message = click.prompt('Message', type=click.STRING)
    sio.emit('message', {'from': username,
                         "message":message})

@cli.command(name='clear')
def clear():
   draw()

def draw():
    global fsmState
    global username
    click.clear()
    cols = os.get_terminal_size().columns
    rows = os.get_terminal_size().lines
    mainWindowWidth = int(cols * 0.9)
    # line = "-" * cols
    # line = click.style("-" * cols,fg=(206, 176, 126))
    line = click.style("=" * cols,fg="yellow")
    click.echo(line)
    chatstatus = click.style(f"\033[{mainWindowWidth}G Chat Active",fg="green")
    if(not isChatActive):
        chatstatus = click.style(f"\033[{mainWindowWidth}G Chat Inactive",fg="red")
    title = click.style("Relay CLI Edition",fg="bright_yellow") + chatstatus
    print(title)
    click.echo(line)

    if(fsmState == FSM.BOOT):
        prompt = click.prompt("Please boot into an option: \n 1: Signup \n 2: Login \n 3: Join")
        if(prompt == "1"):
            fsmState = FSM.SIGNUP
            draw()
        if(prompt == "2"):
            fsmState = FSM.LOGIN
        if(prompt == "3"):
            fsmState = FSM.CHANNEL
        return
    if(fsmState == FSM.SIGNUP):
        username = click.prompt("Enter a username")
        fsmState = FSM.CHANNEL
        draw()
        return
    


    
        
    

    for message in localCache:
        click.echo(f"{message['from']} says: \n {message['message']}")

    #DRAW THE RIGHT SIDE UI
    # Header at the right edge
    print(f"\033[{4};{mainWindowWidth}H| Online Users")
    print(f"\033[{5};{mainWindowWidth}H|{'-' * (cols-mainWindowWidth)}")
    # print(f"\033[{mainWindowWidth}G|" + "-" * (cols-mainWindowWidth))

    for _ in range(6,rows):
        print(f"\033[{_};{mainWindowWidth}H|")
        

@cli.command(name='q')
def quitapp():
    raise click.exceptions.Abort()
    

def client():
    draw()   
    global isChatActive 
    global username
    while True:
        try:
            user_input = click.prompt("> ", prompt_suffix="")

            if user_input.lower() == "chat":
                isChatActive = not isChatActive
                draw()
                continue

            if isChatActive:
                sio.emit("message", {"from": username, "message": str(user_input)})
            else:
                cmd = user_input.lower()
                if cmd == "q" or cmd == "quit":
                    break
                else:
                    click.secho(f"Unknown command: {user_input}", fg="red")

        except click.exceptions.Abort:
           break 

client()