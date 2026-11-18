package k8s

import (
	"os"
	"testing"

	"github.com/mhristof/germ/internal/testenv"
)

func TestMain(m *testing.M) {
	os.Exit(testenv.Run(m))
}
