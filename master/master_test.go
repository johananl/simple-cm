package master

import (
	"io/ioutil"
	"net/rpc"
	"os"
	"testing"
)

func TestSSHKey(t *testing.T) {
	m := Master{SSHKeysDir: "./"}

	// TODO Generate a real SSH key
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

func TestSelectWorker(t *testing.T) {
	w := []*rpc.Client{
		&rpc.Client{},
		&rpc.Client{},
		&rpc.Client{},
		&rpc.Client{},
		&rpc.Client{},
	}
	m := Master{
		Workers:        w,
		LastUsedWorker: 0,
	}

	var want int
	for i := 0; i < len(w); i++ {
		if want == len(w)-1 {
			want = 0
		} else {
			want = i + 1
		}

		c := m.SelectWorker()
		if c != m.Workers[want] {
			t.Fatalf("Wrong worker selected: got %v want %v", &c, &m.Workers[want])
		}
	}
}
