package master

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestSSHKey(t *testing.T) {
	m := Master{SSHKeysPath: "./"}

	fn := "./test_key"
	contents := "secretstuff"
	err := ioutil.WriteFile(fn, []byte(contents), 0644)
	if err != nil {
		t.Fatalf("error writing dummy key: %v", err)
	}
	defer func() {
		err := os.Remove(fn)
		if err != nil {
			t.Logf("could not clean up key after testing: %v", err)
		}
	}()

	k, err := m.SSHKey("test_key")
	if err != nil {
		t.Fatalf("error reading key: %v", err)
	}

	if k != contents {
		t.Fatalf("wrong contents read from key: got %v want %v", k, contents)
	}
}
