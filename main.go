package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"connectrpc.com/connect"
	v1 "remote-claude-code-api/gen/claude/v1"
	"remote-claude-code-api/gen/claude/v1/claudev1connect"
)

var repoPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)

type ClaudeServer struct{}

func buildCloneArgs(repo, tmpDir string) []string {
	args := []string{"clone", "--depth=1"}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		args = append(args, "-c", "http.extraheader=AUTHORIZATION: bearer "+token)
	}
	return append(args, "https://github.com/"+repo+".git", tmpDir)
}

func (s *ClaudeServer) Run(ctx context.Context, req *v1.RunRequest) (*v1.RunResponse, error) {
	var workDir string

	if req.Repository != "" {
		if !repoPattern.MatchString(req.Repository) {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid repository format: must be owner/repo"))
		}

		tmpDir, err := os.MkdirTemp("", "claude-repo-*")
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		cloneArgs := buildCloneArgs(req.Repository, tmpDir)
		cloneCmd := exec.CommandContext(ctx, "git", cloneArgs...)
		if out, err := cloneCmd.CombinedOutput(); err != nil {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("git clone failed: %w\n%s", err, out))
		}

		workDir = tmpDir
	}

	cmd := exec.CommandContext(ctx, "claude", "--remote", req.Prompt)
	if workDir != "" {
		cmd.Dir = workDir
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return &v1.RunResponse{Output: strings.TrimRight(string(out), "\n")}, nil
}

// pathRewriter は "/" へのリクエストを "/claude.v1.ClaudeService/Run" にリライトする
func pathRewriter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			r = r.Clone(r.Context())
			r.URL.Path = claudev1connect.ClaudeServiceRunProcedure
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.Handle(claudev1connect.NewClaudeServiceHandler(&ClaudeServer{}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: pathRewriter(mux),
	}
	// Go 1.26: http.Protocols で h2c サポート（golang.org/x/net/http2/h2c 不要）
	var protos http.Protocols
	protos.SetHTTP1(true)
	protos.SetUnencryptedHTTP2(true)
	srv.Protocols = &protos

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
