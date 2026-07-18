package store

import (
	"time"

	"decoreba/internal/core"
)

func seedCommands() []core.Command {
	now := time.Now()
	mk := func(ctx, title, command string, tags ...string) core.Command {
		return core.Command{
			ID:        core.GenID(),
			Context:   ctx,
			Title:     title,
			Command:   command,
			Tags:      tags,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	return []core.Command{
		mk("tmux", "Split pane horizontally", "tmux split-window -h", "pane", "split"),
		mk("tmux", "Split pane vertically", "tmux split-window -v", "pane", "split"),
		mk("tmux", "Create named session", "tmux new -s session_name", "session"),
		mk("tmux", "Attach to last session", "tmux attach", "session", "attach"),
		mk("tmux", "List sessions", "tmux ls", "session", "list"),
		mk("tmux", "Kill session by name", "tmux kill-session -t session_name", "session", "kill"),
		mk("git", "Undo last commit keeping changes", "git reset --soft HEAD~1", "reset", "undo"),
		mk("git", "Compact graph log", "git log --oneline --graph --decorate --all", "log"),
		mk("git", "Change last commit message", `git commit --amend -m "new message"`, "amend"),
		mk("git", "Apply and drop most recent stash", "git stash pop", "stash"),
		mk("git", "Delete merged local branch", "git branch -d branch_name", "branch", "delete"),
		mk("git", "Show staged diff", "git diff --staged", "diff"),
		mk("docker", "Remove stopped containers", "docker container prune", "cleanup"),
		mk("docker", "Follow container logs", "docker logs -f container_name", "logs"),
		mk("docker", "Enter a running container", "docker exec -it container_name sh", "shell", "exec"),
	}
}
