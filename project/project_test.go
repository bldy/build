package project

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetGitDir(t *testing.T) {
	tests := map[string]string{
		filepath.Join(os.Getenv("GOPATH"), "src/bldy.build/build/blaze/processor"): filepath.Join(os.Getenv("GOPATH"), "src/bldy.build/build"),
	}

	for dir, expected := range tests {
		t.Run(fmt.Sprintf("GetGitDir-%s", dir), func(t *testing.T) {
			if GetGitDir(dir) != expected {
				t.Errorf("expected %s got %s", expected, GetGitDir(dir))
			}
		})
	}
}
