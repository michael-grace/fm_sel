import datetime
from flask import Flask, request, abort
from flask_cors import CORS
import zmq

"""
Sources:
	0: PGM Output
	1: Jukebox
	2: AutoNews
"""

app = Flask(__name__)
FILE_PATH = 'source.txt'
LOG_FILE = "sel.log"
ALLOWED_IP = '127.0.0.1'  # Only requests from localhost are allowed
CORS(app)
cors = CORS(app, resource={
    r"/*":{
        "origins":"*"
    }
})
ctx = zmq.Context()

@app.route('/source', methods=['POST'])
def set_source():
    if request.remote_addr != ALLOWED_IP:
        abort(403)  # Forbidden if request is not from localhost

    source = request.form.get('source')
    if source not in ['0', '1', '2']:
        return 'Invalid source. Please select 0, 1, or 2.', 400

    with open(FILE_PATH, 'w') as file:
        file.write(source)

    socket = ctx.socket(zmq.REQ)
    socket.connect("tcp://localhost:5555")
    for i in range(3):
        socket.send(bytes(f"volume@s{i} volume 0", "UTF-8"))
        socket.recv()
    socket.send(bytes(f"volume@s{source} volume 1", "UTF-8"))
    socket.recv()

    with open(LOG_FILE, "a") as log_file:
        log_file.write(f"{str(datetime.datetime.now())}: Selected {source}\n")

    return 'Source set successfully.', 200

@app.route('/source', methods=['GET'])
def get_source():
    try:
        with open(FILE_PATH, 'r') as file:
            source = file.read().strip()
        return source, 200
    except FileNotFoundError:
        return 'No source found.', 404

if __name__ == '__main__':
    app.run(debug=True)

