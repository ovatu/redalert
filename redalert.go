package main // import "github.com/ovatu/redalert"

import (
	"github.com/ovatu/redalert/cmd"
	"github.com/ovatu/redalert/utils"
)

var (
	version string
	commit  string
)

func main() {
	utils.RegisterVersionAndBuild(version, commit)
	cmd.Execute()
}
