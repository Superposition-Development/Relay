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


@sio.on('response')
def response(data):
    print(data)

@click.group(invoke_without_command=True)
@click.option("--command", prompt=">")
@click.pass_context
def cli(ctx, command):
    cmd = cli.get_command(ctx, command)
    ctx.invoke(cmd)

@cli.command(name='connect')
def connect():
    username = click.prompt('Username', type=click.STRING)
    sio.emit('message', {'from': username})

@cli.command(name='q')
def quitapp():
    raise click.exceptions.Abort()
    

def main():    
    while True:
        try:
           cli.main(standalone_mode=False)
        except click.exceptions.Abort:
           break 


if __name__ == '__main__':
    main()