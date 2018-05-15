package operations

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestScript(t *testing.T) {
	fakeScript := "this is a fake script: {{.what}} {{.ever}}"
	ioutil.WriteFile("test.txt", []byte(fakeScript), 0644)
	defer func() {
		os.Remove("test.txt")
	}()

	o := Operation{
		Description: "test_op",
		ScriptName:  "test.txt",
		Attributes: map[string]string{
			"what": "what",
			"ever": "ever",
		},
	}

	s, err := o.Script(".")
	if err != nil {
		t.Fatal(err)
	}

	want := "this is a fake script: what ever"
	if s != want {
		t.Fatalf("wrong content: got %s want %s", s, want)
	}
}
