#!/usr/bin/env python3
import os
import sys
import shutil
import subprocess
import platform
import time
import re

# =========================================================

# Config

# =========================================================

APP_NAME = "Secure File Box"
MIN_GO_VERSION = (1, 20)
DB_NAME = "secure_file_box"
DB_USER = "root"
DB_PASS = "0827"
DB_HOST = "127.0.0.1"
DB_PORT = 3306

ROOT = os.path.dirname(os.path.abspath(__file__))

# =========================================================

# Color log

# =========================================================

class C:
    GREEN = "\033[92m"
    RED = "\033[91m"
    YELLOW = "\033[93m"
    CYAN = "\033[96m"
    END = "\033[0m"

def log(msg): print(f"{C.CYAN}➡ {msg}{C.END}")
def ok(msg): print(f"{C.GREEN}✔ {msg}{C.END}")
def warn(msg): print(f"{C.YELLOW}⚠ {msg}{C.END}")
def err(msg): print(f"{C.RED}✖ {msg}{C.END}")

# =========================================================

# Utils

# =========================================================

def run(cmd, check=True, capture=False, cwd=None):
    log(cmd)
    return subprocess.run(
        cmd,
        shell=True,
        check=check,
        cwd=cwd,
        text=True,
        capture_output=capture
    )

def has_cmd(name):
    return shutil.which(name) is not None

def system():
    return platform.system().lower()

# =========================================================

# Go check

# =========================================================

def parse_go_version(text):
    m = re.search(r"go(\d+).(\d+)", text)
    if not m:
        return None
    return int(m.group(1)), int(m.group(2))

def check_go():
    if not has_cmd("go"):
        err("Go not found in PATH")
        sys.exit(1)

    r = run("go version", capture=True)
    ver = parse_go_version(r.stdout)

    if not ver:
        err("Cannot parse Go version")
        sys.exit(1)

    if ver < MIN_GO_VERSION:
        err(f"Go version too low: {ver}, need >= {MIN_GO_VERSION}")
        sys.exit(1)

    ok(f"Go version OK: {ver[0]}.{ver[1]}")


# =========================================================

# Docker path (BEST)

# =========================================================

def has_docker():
    return has_cmd("docker") and has_cmd("docker compose")

def docker_up():
    if not os.path.exists(os.path.join(ROOT, "docker-compose.yml")):
        return False

    ok("Using Docker deployment (recommended)")
    run("docker compose up -d", cwd=ROOT)
    return True

# =========================================================

# MySQL

# =========================================================

def wait_mysql(timeout=60):
    ok("Waiting for MySQL...")
    start = time.time()

    while time.time() - start < timeout:
        try:
            run(
                f'mysqladmin ping -h {DB_HOST} -P {DB_PORT} -u{DB_USER} -p{DB_PASS} --silent',
                check=True,
            )
            ok("MySQL is ready")
            return True
        except Exception:
            time.sleep(2)

    err("MySQL not ready after timeout")
    return False

def ensure_mysql_client():
    if not has_cmd("mysql"):
        err("mysql client not found")

        os_name = system()
        print("\nInstall MySQL client:\n")
        if os_name == "linux":
            print("  sudo apt install mysql-client")
        elif os_name == "darwin":
            print("  brew install mysql-client")
        elif os_name == "windows":
            print("  winget install MySQL.MySQLServer")

        sys.exit(1)

def init_database():
    ok("Initializing database")


    sql = f"CREATE DATABASE IF NOT EXISTS {DB_NAME};"
    run(
        f'mysql -h {DB_HOST} -P {DB_PORT} -u{DB_USER} -p{DB_PASS} -e "{sql}"',
        check=True,
    )

# =========================================================

# Storage

# =========================================================

def ensure_storage():
    path = os.path.join(ROOT, "storage")
    os.makedirs(path, exist_ok=True)
    ok("storage ready")

# =========================================================

# Go run/build

# =========================================================

def run_dev():
    ok("Starting in DEV mode")
    run("go run ./cmd/server", cwd=ROOT)
    
def run_prod():
    ok("Building production binary")
    run("go build -o bin/app ./cmd/server", cwd=ROOT)
    
    ok("Starting in PROD mode")
    run("./bin/app", cwd=ROOT)

# =========================================================

# Main

# =========================================================

def main():
    print(f"\n=== {APP_NAME} Launcher ===\n")

    mode = "dev"
    if len(sys.argv) > 1:
        mode = sys.argv[1]
    
    # ---------- checks ----------
    check_go()
    ensure_storage()
    
    # ---------- docker fast path ----------
    if has_docker():
        if docker_up():
            ok("Application started via Docker")
            return
    
    # ---------- native mysql ----------
    ensure_mysql_client()
    
    if not wait_mysql():
        err("Please start MySQL service first")
        sys.exit(1)
    
    init_database()
    
    # ---------- run ----------
    if mode == "prod":
        run_prod()
    else:
        run_dev()
    
if __name__ == "__main__":
    main()

