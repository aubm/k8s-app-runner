from flask import Flask
import os
app = Flask(__name__)

@app.route('/')
def hello_world():
    pod_name = os.environ.get('POD_NAME', 'some pod')
    app_name = os.environ.get('APP_NAME', 'some app')
    app_runtime = os.environ.get('APP_RUNTIME', 'some runtime')
    return 'Hello from pod {}, this is app {} with runtime'.format(pod_name, app_name, app_runtime)


if __name__ == '__main__':
    app.run()
