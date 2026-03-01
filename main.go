package main

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"connectrpc.com/connect"
	v1 "remote-claude-code-api/gen/claude/v1"
	"remote-claude-code-api/gen/claude/v1/claudev1connect"
)

type ClaudeServer struct{}

func (s *ClaudeServer) Run(ctx context.Context, req *v1.RunRequest) (*v1.RunResponse, error) {
	out, err := exec.CommandContext(ctx, "claude", "--remote", req.Prompt).Output()
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
