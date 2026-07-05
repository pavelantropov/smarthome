from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from typing import Any
from uuid import uuid4

from flask import Flask, jsonify, request


app = Flask(__name__)


@dataclass
class Device:
    id: str
    name: str
    type: str
    location: str = ""
    status: str = "offline"
    metadata: dict[str, Any] = field(default_factory=dict)
    created_at: str = field(default_factory=lambda: now_iso())
    updated_at: str = field(default_factory=lambda: now_iso())


@dataclass
class Command:
    id: str
    device_id: str
    command: str
    payload: dict[str, Any] = field(default_factory=dict)
    status: str = "queued"
    created_at: str = field(default_factory=lambda: now_iso())


devices: dict[str, Device] = {}
commands: dict[str, list[Command]] = {}


@app.get("/health")
def health():
    return jsonify({"status": "ok", "service": "device"})


@app.post("/api/devices")
def create_device():
    body = request.get_json(silent=True) or {}
    error = validate_device_payload(body, partial=False)
    if error:
        return jsonify({"error": error}), 400

    device = Device(
        id=uuid(),
        name=body["name"].strip(),
        type=body["type"].strip(),
        location=(body.get("location") or "").strip(),
        status=(body.get("status") or "offline").strip(),
        metadata=body.get("metadata") or {},
    )
    devices[device.id] = device
    commands[device.id] = []
    return jsonify(asdict(device)), 201


@app.get("/api/devices")
def list_devices():
    device_type = request.args.get("type")
    status = request.args.get("status")
    result = list(devices.values())

    if device_type:
        result = [device for device in result if device.type == device_type]
    if status:
        result = [device for device in result if device.status == status]

    result.sort(key=lambda device: device.created_at)
    return jsonify([asdict(device) for device in result])


@app.get("/api/devices/<device_id>")
def get_device(device_id: str):
    device = devices.get(device_id)
    if not device:
        return jsonify({"error": "device not found"}), 404
    return jsonify(asdict(device))


@app.put("/api/devices/<device_id>")
def update_device(device_id: str):
    device = devices.get(device_id)
    if not device:
        return jsonify({"error": "device not found"}), 404

    body = request.get_json(silent=True) or {}
    error = validate_device_payload(body, partial=True)
    if error:
        return jsonify({"error": error}), 400

    if "name" in body:
        device.name = body["name"].strip()
    if "type" in body:
        device.type = body["type"].strip()
    if "location" in body:
        device.location = (body.get("location") or "").strip()
    if "status" in body:
        device.status = body["status"].strip()
    if "metadata" in body:
        device.metadata = body.get("metadata") or {}
    device.updated_at = now_iso()

    return jsonify(asdict(device))


@app.delete("/api/devices/<device_id>")
def delete_device(device_id: str):
    if device_id not in devices:
        return jsonify({"error": "device not found"}), 404
    del devices[device_id]
    commands.pop(device_id, None)
    return "", 204


@app.post("/api/devices/<device_id>/commands")
def create_command(device_id: str):
    if device_id not in devices:
        return jsonify({"error": "device not found"}), 404

    body = request.get_json(silent=True) or {}
    command_name = body.get("command")
    if not isinstance(command_name, str) or not command_name.strip():
        return jsonify({"error": "command is required"}), 400
    payload = body.get("payload") or {}
    if not isinstance(payload, dict):
        return jsonify({"error": "payload must be an object"}), 400

    command = Command(
        id=uuid(),
        device_id=device_id,
        command=command_name.strip(),
        payload=payload,
    )
    commands.setdefault(device_id, []).append(command)
    return jsonify(asdict(command)), 202


@app.get("/api/devices/<device_id>/commands")
def list_commands(device_id: str):
    if device_id not in devices:
        return jsonify({"error": "device not found"}), 404
    return jsonify([asdict(command) for command in commands.get(device_id, [])])


def validate_device_payload(body: dict[str, Any], partial: bool) -> str | None:
    required = ("name", "type")
    for field_name in required:
        if not partial and not isinstance(body.get(field_name), str):
            return f"{field_name} is required"
        if field_name in body and not body[field_name].strip():
            return f"{field_name} cannot be empty"

    if "status" in body and not isinstance(body["status"], str):
        return "status must be a string"
    if "metadata" in body and not isinstance(body["metadata"], dict):
        return "metadata must be an object"
    return None


def now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()


def uuid() -> str:
    return str(uuid4())


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8082)
