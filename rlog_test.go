package rlog_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/OneOfOne/rlog"
)

func TestLog(t *testing.T) {
	tdir, err := ioutil.TempDir("", "rlog")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	t.Logf("dir: %s", tdir)

	rl := rlog.New(tdir, "2006-01-02-15-04-05.json.gz", nil)
	defer rl.Close()

	rl.Log(rlog.M{"hi": 1})
	time.Sleep(time.Second)
	rl.Log(rlog.M{"hi": 1})
}
