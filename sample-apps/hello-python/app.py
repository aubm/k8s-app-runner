from flask import Flask
import os
app = Flask(__name__)

@app.route('/')
def hello_world():
    hello_to = os.environ.get('HELLO_TO', 'world')
    app_name = os.environ.get('APP_NAME', 'some app')
    app_runtime = os.environ.get('APP_RUNTIME', 'some runtime')
    pod_name = os.environ.get('POD_NAME', 'some pod')
    return 'Hello {}! This is app {}, using runtime {} in pod {}'.format(hello_to, app_name, app_runtime, app_runtime)


if __name__ == '__main__':
    app.run()
