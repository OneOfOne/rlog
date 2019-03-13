package rlog_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/OneOfOne/rlog"
)

func TestLog(t *testing.T) {
	tdir, err := ioutil.TempDir("", "rlog")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	t.Logf("dir: %s", tdir)

	rl := rlog.New(tdir, "2006-01-02.json", nil)
	defer rl.Close()

	rl.Log(0, rlog.M{"hi": 1})

}
