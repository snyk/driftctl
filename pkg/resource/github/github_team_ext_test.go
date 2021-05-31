package github

import (
	"reflect"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
)

func TestGithubTeam_Attributes(t *testing.T) {
	tests := []struct {
		name string
		team GithubTeam
		want map[string]string
	}{
		{
			name: "test with name",
			team: GithubTeam{
				Id:   "1234",
				Name: awssdk.String("my-org-name"),
			},
			want: map[string]string{
				"Name": "my-org-name",
				"Id":   "1234",
			},
		},
		{
			name: "test without name",
			team: GithubTeam{
				Id: "1234",
			},
			want: map[string]string{
				"Id": "1234",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.team.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}
