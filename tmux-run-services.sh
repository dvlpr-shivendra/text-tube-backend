#!/bin/bash
set -e

tmux new-session -d -s texttube 'cd auth-service && air; exec bash'
tmux split-window -h 'cd video-service && air; exec bash'
tmux split-window -v 'cd gateway && air; exec bash'
tmux select-pane -t 0
tmux attach-session -t texttube
