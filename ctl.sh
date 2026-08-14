#!/bin/sh

set -u

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
APP_BIN="$ROOT_DIR/nova"
PID_FILE="$ROOT_DIR/nova.pid"
STOP_TIMEOUT=30

read_pid() {
    if [ ! -f "$PID_FILE" ]; then
        return 1
    fi

    pid=$(tr -d '[:space:]' < "$PID_FILE")
    case "$pid" in
        ''|*[!0-9]*) return 1 ;;
    esac
    if [ "$pid" -le 0 ] 2>/dev/null; then
        return 1
    fi

    printf '%s\n' "$pid"
}

is_running() {
    kill -0 "$1" 2>/dev/null
}

is_nova_process() {
    checked_pid=$1
    command_line=$(ps -p "$checked_pid" -o args= 2>/dev/null || true)
    binary_name=$(basename "$APP_BIN")

    case "$command_line" in
        *"$APP_BIN"*|*"$binary_name"*) return 0 ;;
        *) return 1 ;;
    esac
}

build_if_needed() {
    rebuild=false
    if [ ! -x "$APP_BIN" ]; then
        rebuild=true
    elif find "$ROOT_DIR" -type f \( -name '*.go' -o -name 'go.mod' -o -name 'go.sum' \) -newer "$APP_BIN" -print -quit | grep -q .; then
        rebuild=true
    fi

    if [ "$rebuild" = true ]; then
        echo "building $APP_BIN"
        (cd "$ROOT_DIR" && go build -o "$APP_BIN" .)
    fi
}

has_config_argument() {
    for argument in "$@"; do
        case "$argument" in
            -config=*) return 0 ;;
        esac
    done
    return 1
}

start_service() {
    if pid=$(read_pid) && is_running "$pid"; then
        echo "nova is already running (pid $pid)"
        return 1
    fi

    if ! has_config_argument "$@"; then
        echo "missing Nova argument: -config=/path/to/env.yaml" >&2
        usage >&2
        return 1
    fi

    rm -f -- "$PID_FILE"
    build_if_needed || return 1

    cd "$ROOT_DIR" || return 1
    if ! "$APP_BIN" -check-config "$@"; then
        echo "nova configuration check failed" >&2
        return 1
    fi

    nohup "$APP_BIN" -start "$@" >/dev/null 2>&1 &
    pid=$!
    printf '%s\n' "$pid" > "$PID_FILE"

    sleep 1
    if ! is_running "$pid"; then
        rm -f -- "$PID_FILE"
        echo "nova failed to start; check env.yaml and the configured logger.filename" >&2
        return 1
    fi

    echo "nova started (pid $pid)"
}

stop_service() {
    if ! pid=$(read_pid) || ! is_running "$pid"; then
        rm -f -- "$PID_FILE"
        echo "nova is not running"
        return 0
    fi

    if ! is_nova_process "$pid"; then
        echo "refusing to stop pid $pid: it does not look like $APP_BIN" >&2
        return 1
    fi

    kill -TERM "$pid"

    waited=0
    while is_running "$pid" && [ "$waited" -lt "$STOP_TIMEOUT" ]; do
        sleep 1
        waited=$((waited + 1))
    done

    if is_running "$pid"; then
        echo "graceful stop timed out after ${STOP_TIMEOUT}s; force stopping pid $pid" >&2
        kill -KILL "$pid"
    fi

    rm -f -- "$PID_FILE"
    echo "nova stopped"
}

status_service() {
    if pid=$(read_pid) && is_running "$pid" && is_nova_process "$pid"; then
        echo "nova is running (pid $pid)"
        return 0
    fi

    echo "nova is not running"
    return 1
}

usage() {
    echo "usage: sh $0 start -config=/etc/nova/env.yaml"
    echo "       sh $0 stop"
    echo "       sh $0 restart -config=/etc/nova/env.yaml"
    echo "       sh $0 status"
}

command=${1:-}
if [ "$#" -gt 0 ]; then
    shift
fi

case "$command" in
    start)
        start_service "$@"
        ;;
    stop)
        stop_service
        ;;
    restart)
        stop_service && start_service "$@"
        ;;
    status)
        status_service
        ;;
    *)
        usage
        exit 2
        ;;
esac
