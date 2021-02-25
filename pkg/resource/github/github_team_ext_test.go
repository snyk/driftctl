package github

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
)

func TestGithubTeam_String(t *testing.T) {
	tests := []struct {
		name string
		team GithubTeam
		want string
	}{
		{
			name: "test with name",
			team: GithubTeam{
				Id:   "1234",
				Name: awssdk.String("my-org-name"),
			},
			want: "my-org-name (Id: 1234)",
		},
		{
			name: "test without name",
			team: GithubTeam{
				Id: "1234",
			},
			want: "1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.team.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
