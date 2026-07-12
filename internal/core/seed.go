package core

import "time"

func seedCommands() []Command {
	now := time.Now()
	mk := func(ctx, title, command string, tags ...string) Command {
		return Command{
			ID:        GenID(),
			Context:   ctx,
			Title:     title,
			Command:   command,
			Tags:      tags,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	return []Command{
		mk("tmux", "Split pane horizontally", "tmux split-window -h", "pane", "split"),
		mk("tmux", "Split pane vertically", "tmux split-window -v", "pane", "split"),
		mk("tmux", "Create named session", "tmux new -s nome_sessao", "session"),
		mk("tmux", "Attach to last session", "tmux attach", "session", "attach"),
		mk("tmux", "List sessions", "tmux ls", "session", "list"),
		mk("tmux", "Kill session by name", "tmux kill-session -t nome_sessao", "session", "kill"),
		mk("git", "Undo last commit keeping changes", "git reset --soft HEAD~1", "reset", "undo"),
		mk("git", "Compact graph log", "git log --oneline --graph --decorate --all", "log"),
		mk("git", "Change last commit message", `git commit --amend -m "new message"`, "amend"),
		mk("git", "Apply and drop most recent stash", "git stash pop", "stash"),
		mk("git", "Delete merged local branch", "git branch -d nome_branch", "branch", "delete"),
		mk("git", "Show staged diff", "git diff --staged", "diff"),
		mk("docker", "Remove stopped containers", "docker container prune", "cleanup"),
		mk("docker", "Follow container logs", "docker logs -f nome_container", "logs"),
		mk("docker", "Enter a running container", "docker exec -it nome_container sh", "shell", "exec"),
	}
}
