#!/usr/bin/env python3
import os
import sys
import shutil
import subprocess
import platform
import re

# =========================================================
# Config
# =========================================================

APP_NAME = "encrypted-file-box"
MIN_GO_VERSION = (1, 20)

ROOT = os.path.dirname(os.path.abspath(__file__))
BIN_DIR = os.path.join(ROOT, "bin")

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
    log(cmd if isinstance(cmd, str) else " ".join(cmd))
    return subprocess.run(
        cmd,
        shell=isinstance(cmd, str),
        check=check,
        cwd=cwd,
        text=True,
        capture_output=capture,
    )

def has_cmd(name):
    return shutil.which(name) is not None

# =========================================================
# Go check
# =========================================================

def parse_go_version(text):
    m = re.search(r"go(\d+)\.(\d+)", text)
    if not m:
        return None
    return int(m.group(1)), int(m.group(2))

def check_go():
    if not has_cmd("go"):
        err("Go not found in PATH")
        sys.exit(1)

    r = run(["go", "version"], capture=True)
    ver = parse_go_version(r.stdout)

    if not ver:
        err("Cannot parse Go version")
        sys.exit(1)

    if ver < MIN_GO_VERSION:
        err(f"Go version too low: {ver}, need >= {MIN_GO_VERSION}")
        sys.exit(1)

    ok(f"Go version OK: {ver[0]}.{ver[1]}")

# =========================================================
# Build helpers
# =========================================================

def ensure_bin():
    os.makedirs(BIN_DIR, exist_ok=True)
    ok("bin/ ready")

def get_platform_ext(goos):
    return ".exe" if goos == "windows" else ""

def build(goos=None, goarch=None, version="dev"):
    ensure_bin()

    env = os.environ.copy()

    if goos:
        env["GOOS"] = goos
    if goarch:
        env["GOARCH"] = goarch

    ext = get_platform_ext(env.get("GOOS", platform.system().lower()))
    output = os.path.join(BIN_DIR, APP_NAME + ext)

    ldflags = f"-s -w -X main.version={version}"

    cmd = [
        "go",
        "build",
        "-ldflags", ldflags,
        "-o", output,
        "./cmd/server",
    ]

    log(f"Building for {env.get('GOOS','native')}/{env.get('GOARCH','native')}")
    run(cmd, cwd=ROOT)

    ok(f"Build success → {output}")

# =========================================================
# Clean
# =========================================================

def clean():
    if os.path.exists(BIN_DIR):
        shutil.rmtree(BIN_DIR)
        ok("bin/ cleaned")
    else:
        warn("bin/ not exists")

# =========================================================
# CLI
# =========================================================

def help_msg():
    print(f"""
=== {APP_NAME} build tool ===

Usage:
  python build.py              # native build
  python build.py prod         # optimized build
  python build.py cross        # cross compile common targets
  python build.py clean        # remove bin
""")

# =========================================================
# Main
# =========================================================

def main():
    check_go()

    mode = sys.argv[1] if len(sys.argv) > 1 else "dev"

    if mode in ("-h", "--help", "help"):
        help_msg()
        return

    if mode == "clean":
        clean()
        return

    if mode == "cross":
        targets = [
            ("linux", "amd64"),
            ("windows", "amd64"),
            ("darwin", "arm64"),
        ]
        for goos, goarch in targets:
            build(goos, goarch, version="release")
        return

    if mode == "prod":
        build(version="release")
        return

    # default dev build
    build(version="dev")

if __name__ == "__main__":
    main()
