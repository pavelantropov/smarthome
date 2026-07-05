from flask import Flask, request, jsonify
import random
app = Flask(__name__)

@app.route('/temperature')
def temperature():
    location = request.args.get('location', '')
    sensor_id = request.args.get('sensorId', '')
    if location == '':
        if sensor_id == '1': location = 'Living Room'
        elif sensor_id == '2': location = 'Bedroom'
        elif sensor_id == '3': location = 'Kitchen'
        else: location = 'Unknown'
    if sensor_id == '':
        if location == 'Living Room': sensor_id = '1'
        elif location == 'Bedroom': sensor_id = '2'
        elif location == 'Kitchen': sensor_id = '3'
        else: sensor_id = '0'
    return jsonify({'location': location, 'sensorId': sensor_id, 'temperature': round(random.uniform(15, 30), 1)})
if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8081)