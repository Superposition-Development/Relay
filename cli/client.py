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

sio = socketio.Client()
sio.connect('http://127.0.0.1:6221')

username = ""


@sio.on('response')
def response(data):
    click.echo(f"{data['from']} says: \n {data['message']}")

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
    click.clear()
    cols = os.get_terminal_size().columns
    rows = os.get_terminal_size().lines
    mainWindowWidth = int(cols * 0.9)
    # line = "-" * cols
    # line = click.style("-" * cols,fg=(206, 176, 126))
    line = click.style("-" * cols,fg="yellow")
    click.echo(line)
    click.echo(click.style("Relay CLI Edition",fg="bright_yellow"))
    click.echo(line)

    # Header at the right edge
    print(f"\033[{mainWindowWidth}G| Online Users")
    print(f"\033[{mainWindowWidth}G|" + "-" * (cols-mainWindowWidth))

    # Draw vertical bar
    for _ in range(rows - 6):
        print(f"\033[{mainWindowWidth}G|")
        

@cli.command(name='q')
def quitapp():
    raise click.exceptions.Abort()
    

def client():
    draw()    
    while True:
        try:
           cli.main(standalone_mode=False)
        except click.exceptions.Abort:
           break 

client()