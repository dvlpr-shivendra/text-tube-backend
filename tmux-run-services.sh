tmux new-session -d -s texttube 'cd auth-service && go run ./cmd/main.go'
tmux split-window -h 'cd video-service && go run ./cmd/main.go'
tmux split-window -v 'cd gateway && go run ./cmd/main.go'
tmux select-pane -t 0
tmux attach-session -t texttube
