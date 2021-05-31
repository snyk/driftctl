package github

import (
	"reflect"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
)

func TestGithubBranchProtection_Attributes(t *testing.T) {
	tests := []struct {
		name string
		res  GithubBranchProtection
		want map[string]string
	}{
		{
			name: "when pattern is nil",
			res: GithubBranchProtection{
				Id: "ABCDEF=",
			},
			want: map[string]string{
				"Id": "ABCDEF=",
			},
		},
		{
			name: "when repo_id is nil",
			res: GithubBranchProtection{
				Id:      "ABCDEF=",
				Pattern: awssdk.String("my-branch"),
			},
			want: map[string]string{
				"Branch": "my-branch",
				"Id":     "ABCDEF=",
			},
		},
		{
			name: "when repo_id is invalid base64 string",
			res: GithubBranchProtection{
				Id:           "ABCDEF=",
				Pattern:      awssdk.String("my-branch"),
				RepositoryId: awssdk.String("invalid"),
			},
			want: map[string]string{
				"Branch": "my-branch",
				"Id":     "ABCDEF=",
			},
		},
		{
			name: "when all fields are valid",
			res: GithubBranchProtection{
				Id:           "ABCDEF=",
				Pattern:      awssdk.String("my-branch"),
				RepositoryId: awssdk.String("MDEwOlJlcG9zaXRvcnkxMjM0NTY="),
			},
			want: map[string]string{
				"Branch": "my-branch",
				"RepoId": "010:Repository123456",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.res.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}
