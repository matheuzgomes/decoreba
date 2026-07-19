package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/sync"
)

func cmdSync(args []string) {
	cfg, err := sync.LoadConfig()
	check(err)

	sub := ""
	force := false
	useEncrypt := false
	for _, a := range args {
		if a == "--force" {
			force = true
		} else if a == "--encrypt" {
			useEncrypt = true
		} else if !strings.HasPrefix(a, "-") {
			sub = a
		}
	}

	switch sub {
	case "init":
		cmdSyncInit(cfg)
	case "status":
		cmdSyncStatus(cfg)
	case "push":
		cmdSyncPush(cfg, force, useEncrypt)
	case "pull":
		cmdSyncPull(cfg, force, useEncrypt)
	default:
		fmt.Println("Usage: decoreba sync <init|status|push|pull> [--force] [--encrypt]")
		fmt.Println()
		fmt.Println("  init     Create a new gist and upload commands.json")
		fmt.Println("  status   Show sync status (ahead/behind/diverged/clean)")
		fmt.Println("  push     Upload local commands to remote gist")
		fmt.Println("  pull     Download remote commands from gist")
		fmt.Println()
		fmt.Println("  Set " + sync.TokenEnvVar + " to your GitHub personal access token (classic, scope: gist).")
	}
}

func getBackend(cfg *sync.Config) sync.Backend {
	token := os.Getenv(sync.TokenEnvVar)
	if token == "" {
		fmt.Fprintln(os.Stderr, "Set " + sync.TokenEnvVar + " to your GitHub personal access token.")
		os.Exit(1)
	}
	return sync.NewGistBackend(token)
}

func cmdSyncInit(cfg *sync.Config) {
	be := getBackend(cfg)

	storePath, err := sync.StorePath()
	check(err)

	data, err := os.ReadFile(storePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No commands found. Add some commands first with 'decoreba add'.")
			return
		}
		check(err)
	}

	id, err := be.Upload(data, "")
	check(err)

	cfg.GistID = id
	cfg.LastLocalHash = sync.HashBytes(data)
	cfg.LastRemoteHash = cfg.LastLocalHash
	check(sync.SaveConfig(cfg))

	fmt.Printf("✓ Gist created: https://gist.github.com/%s\n", id)
}

func cmdSyncStatus(cfg *sync.Config) {
	if cfg.GistID == "" {
		fmt.Println("Not configured. Run 'decoreba sync init' first.")
		return
	}

	storePath, err := sync.StorePath()
	check(err)

	localHash, err := sync.FileHash(storePath)
	if err != nil {
		if os.IsNotExist(err) {
			localHash = sync.HashBytes([]byte("{}"))
		} else {
			check(err)
		}
	}

	var remoteHash string
	remoteData, err := getBackend(cfg).Download(cfg.GistID)
	if err != nil {
		fmt.Printf("Remote:  error (%v)\n", err)
		remoteHash = ""
	} else {
		remoteHash = sync.HashBytes(remoteData)
	}

	status := sync.ComputeStatus(cfg, localHash, remoteHash)

	fmt.Println("Local SHA256:   " + localHash)
	fmt.Println("Remote SHA256:  " + remoteHash)
	fmt.Println("Last sync hash: " + cfg.LastLocalHash)
	fmt.Println()
	fmt.Print("Status: ")

	switch status {
	case sync.StatusClean:
		fmt.Println("✓ clean (local and remote are in sync)")
	case sync.StatusAhead:
		fmt.Println("↑ ahead (local has changes, remote is behind)")
		fmt.Println("  Run 'decoreba sync push' to upload.")
	case sync.StatusBehind:
		fmt.Println("↓ behind (remote has changes, local is behind)")
		fmt.Println("  Run 'decoreba sync pull' to download.")
	case sync.StatusDiverged:
		fmt.Println("⚠ diverged (both local and remote have changes)")
		fmt.Println("  Use 'decoreba sync push --force' to overwrite remote,")
		fmt.Println("  or 'decoreba sync pull --force' to overwrite local.")
	case sync.StatusNoRemote:
		fmt.Println("? no remote data (gist may be empty)")
	}
}

func cmdSyncPush(cfg *sync.Config, force, useEncrypt bool) {
	if cfg.GistID == "" {
		fmt.Println("No gist configured. Run 'decoreba sync init' first.")
		return
	}

	be := getBackend(cfg)
	storePath, err := sync.StorePath()
	check(err)

	data, err := os.ReadFile(storePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No commands found. Nothing to push.")
			return
		}
		check(err)
	}

	localHash := sync.HashBytes(data)

	remoteData, err := be.Download(cfg.GistID)
	remoteHash := ""
	if err == nil {
		remoteHash = sync.HashBytes(remoteData)
	}

	status := sync.ComputeStatus(cfg, localHash, remoteHash)
	if status == sync.StatusBehind && !force {
		fmt.Println("Remote is ahead of local. Pull first or use --force to overwrite.")
		return
	}
	if status == sync.StatusDiverged && !force {
		fmt.Println("Local and remote have diverged. Use --force to overwrite remote.")
		return
	}

	if status == sync.StatusClean {
		fmt.Println("Nothing to push.")
		return
	}

	if useEncrypt {
		fmt.Fprint(os.Stderr, "Passphrase: ")
		passphrase, err := readPassphrase()
		check(err)
		data, err = sync.Encrypt(data, passphrase)
		check(err)
		cfg.Encrypt = true
	}

	_, err = be.Upload(data, cfg.GistID)
	check(err)

	cfg.LastLocalHash = localHash
	cfg.LastRemoteHash = localHash
	check(sync.SaveConfig(cfg))

	fmt.Println("✓ Pushed " + humanCount(data) + " to gist.")
}

func cmdSyncPull(cfg *sync.Config, force, useEncrypt bool) {
	if cfg.GistID == "" {
		fmt.Println("No gist configured. Run 'decoreba sync init' first.")
		return
	}

	be := getBackend(cfg)
	storePath, err := sync.StorePath()
	check(err)

	localHash := ""
	if data, err := os.ReadFile(storePath); err == nil {
		localHash = sync.HashBytes(data)
	}

	remoteData, err := be.Download(cfg.GistID)
	check(err)

	remoteHash := sync.HashBytes(remoteData)

	status := sync.ComputeStatus(cfg, localHash, remoteHash)
	if status == sync.StatusAhead && !force {
		fmt.Println("Local is ahead of remote. Push first or use --force to overwrite local.")
		return
	}
	if status == sync.StatusDiverged && !force {
		fmt.Println("Local and remote have diverged. Use --force to overwrite local.")
		return
	}
	if status == sync.StatusClean {
		fmt.Println("Nothing to pull.")
		return
	}

	decrypted := remoteData
	if useEncrypt || cfg.Encrypt {
		fmt.Fprint(os.Stderr, "Passphrase: ")
		passphrase, err := readPassphrase()
		check(err)
		decrypted, err = sync.Decrypt(remoteData, passphrase)
		check(err)
	}

	if !useEncrypt && !cfg.Encrypt {
		var s core.Store
		if err := json.Unmarshal(remoteData, &s); err != nil {
			fmt.Fprintln(os.Stderr, "Remote data is not valid JSON. Try --encrypt (wrong passphrase?)")
			os.Exit(1)
		}
	}

	if err := os.WriteFile(storePath, decrypted, 0o600); err != nil {
		check(err)
	}

	cfg.LastLocalHash = remoteHash
	cfg.LastRemoteHash = remoteHash
	check(sync.SaveConfig(cfg))

	fmt.Println("✓ Pulled from gist.")
}

func humanCount(data []byte) string {
	var s core.Store
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Sprintf("%d bytes", len(data))
	}
	if len(s.Commands) == 1 {
		return "1 command"
	}
	return fmt.Sprintf("%d commands", len(s.Commands))
}

func readPassphrase() (string, error) {
	var pass string
	for {
		var b [1]byte
		_, err := os.Stdin.Read(b[:])
		if err != nil {
			return "", err
		}
		if b[0] == '\n' {
			fmt.Fprintln(os.Stderr)
			break
		}
		if b[0] == '\r' {
			continue
		}
		if b[0] == 127 || b[0] == '\b' {
			if len(pass) > 0 {
				pass = pass[:len(pass)-1]
			}
			continue
		}
		pass += string(b[:])
	}
	if pass == "" {
		return "", fmt.Errorf("passphrase cannot be empty")
	}
	return pass, nil
}
