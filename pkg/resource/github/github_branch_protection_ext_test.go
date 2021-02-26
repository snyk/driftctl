package github

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
)

func TestGithubBranchProtection_String(t *testing.T) {
	tests := []struct {
		name string
		res  GithubBranchProtection
		want string
	}{
		{
			name: "when pattern is nil",
			res: GithubBranchProtection{
				Id: "ABCDEF=",
			},
			want: "ABCDEF=",
		},
		{
			name: "when repo_id is nil",
			res: GithubBranchProtection{
				Id:      "ABCDEF=",
				Pattern: awssdk.String("my-branch"),
			},
			want: "Branch: my-branch (Id: ABCDEF=)",
		},
		{
			name: "when repo_id is invalid base64 string",
			res: GithubBranchProtection{
				Id:           "ABCDEF=",
				Pattern:      awssdk.String("my-branch"),
				RepositoryId: awssdk.String("invalid"),
			},
			want: "Branch: my-branch (Id: ABCDEF=)",
		},
		{
			name: "when all fields are valid",
			res: GithubBranchProtection{
				Id:           "ABCDEF=",
				Pattern:      awssdk.String("my-branch"),
				RepositoryId: awssdk.String("MDEwOlJlcG9zaXRvcnkxMjM0NTY="),
			},
			want: "Branch: my-branch (RepoId: 010:Repository123456)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.res.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
