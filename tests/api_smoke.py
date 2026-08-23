"""API smoke tests. Run inside docker compose profile qa. Cost ¥0."""
import json
import os
import time
import urllib.error
import urllib.request

API = os.environ.get("API_BASE", "http://backend:8080").rstrip("/")
LAB = os.environ.get("LAB_BASE", "http://target-lab:8090").rstrip("/")
WEB = os.environ.get("WEB_BASE", "http://frontend-user").rstrip("/")


def get(url, timeout=8):
    with urllib.request.urlopen(url, timeout=timeout) as r:
        return r.status, r.read()


def post_json(url, payload, timeout=8):
    data = json.dumps(payload).encode()
    req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=timeout) as r:
        return r.status, json.loads(r.read().decode())


def test_health_and_lab():
    st, body = get(f"{API}/api/v1/health")
    assert st == 200
    assert b"ok" in body
    st, body = get(f"{LAB}/health")
    assert st == 200
    st, body = get(f"{LAB}/swagger.json")
    assert st == 200
    spec = json.loads(body.decode())
    assert "paths" in spec


def test_frontend_shell():
    st, body = get(WEB + "/")
    assert st == 200
    text = body.decode("utf-8", "ignore")
    assert "id=\"app\"" in text
    assert "GoAPIScanner" in text


def test_reject_unauthorized_scan():
    try:
        post_json(f"{API}/api/v1/scans", {"base_url": LAB, "authorized": False})
        raise AssertionError("should reject")
    except urllib.error.HTTPError as e:
        assert e.code == 400


def test_scan_lab_recall():
    st, env = post_json(
        f"{API}/api/v1/scans",
        {"base_url": LAB, "authorized": True, "concurrency": 12, "timeout_ms": 8000},
    )
    assert st == 201
    task_id = env["data"]["id"]
    classes = set()
    deadline = time.time() + 90
    while time.time() < deadline:
        _, raw = get(f"{API}/api/v1/scans/{task_id}")
        task = json.loads(raw.decode())["data"]
        if task["status"] in ("succeeded", "failed", "cancelled"):
            assert task["status"] == "succeeded", task
            _, raw = get(f"{API}/api/v1/scans/{task_id}/findings")
            pack = json.loads(raw.decode())["data"]
            classes = {f["class"] for f in pack["findings"]}
            break
        time.sleep(1.2)
    expect = {
        "sql_injection",
        "time_blind_sqli",
        "xss",
        "unauthorized",
        "path_traversal",
        "command_injection",
    }
    missing = expect - classes
    assert not missing, f"recall miss {missing} got={classes}"
    _, raw = get(f"{API}/api/v1/scans/{task_id}/report")
    report = json.loads(raw.decode())["data"]
    assert report["advice"]
    st, pdf = get(f"{API}/api/v1/scans/{task_id}/report.pdf", timeout=20)
    assert st == 200
    assert pdf[:4] == b"%PDF"
