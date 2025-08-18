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

@cli.command(name='q')
def quitapp():
    raise click.exceptions.Abort()
    

def client():    
    while True:
        try:
           cli.main(standalone_mode=False)
        except click.exceptions.Abort:
           break 

client()