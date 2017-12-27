package rook

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Southclaws/sampctl/versioning"
)

func TestPackage_Install(t *testing.T) {
	type args struct {
		targets     []versioning.DependencyString
		development bool
	}
	tests := []struct {
		name    string
		pkg     []byte
		args    args
		wantErr bool
	}{
		{"simple", []byte(`{
			"user": "Southclaws",
			"repo": "install-test",
			"entry": "gamemodes/test.pwn",
			"output": "gamemodes/test.amx",
			"dependencies": ["Southclaws/samp-stdlib:0.3.7-R2-2-1"]
		}`), args{[]versioning.DependencyString{"Southclaws/samp-ini"}, false}, false},
		{"dev", []byte(`{
			"user": "Southclaws",
			"repo": "install-test",
			"entry": "gamemodes/test.pwn",
			"output": "gamemodes/test.amx",
			"dependencies": ["Southclaws/samp-stdlib:0.3.7-R2-2-1"]
		}`), args{[]versioning.DependencyString{"Southclaws/samp-ini"}, true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join("./tests/install", tt.name)
			os.RemoveAll(dir)
			os.MkdirAll(dir, 0755)

			ioutil.WriteFile(filepath.Join(dir, "pawn.json"), tt.pkg, 0755) // nolint

			pkg, err := PackageFromDir(true, dir, "")
			if err != nil {
				t.Error(err)
			}

			err = Install(pkg, tt.args.targets, tt.args.development)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			pkg, err = PackageFromDir(true, dir, "")
			if err != nil {
				t.Error(err)
			}

			if tt.args.development {
				for _, target := range tt.args.targets {
					assert.Contains(t, pkg.Development, target)
				}
			} else {
				for _, target := range tt.args.targets {
					assert.Contains(t, pkg.Dependencies, target)
				}
			}
		})
	}
}
