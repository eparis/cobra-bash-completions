package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd"
	"github.com/spf13/cobra"
)

func genTest(cmd *cobra.Command, filename string) {
	out := new(bytes.Buffer)

	cobra.GenCompletion(cmd, out)

	outFile, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()
	_, err = outFile.Write(out.Bytes())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	kubectl := cmd.NewFactory(nil).NewKubectlCommand(os.Stdin, ioutil.Discard, ioutil.Discard)
	genTest(kubectl, "/tmp/kubectl")

	//osc := cli.NewCommandCLI("osc", "osc")
	//genTest(osc, "/tmp/osc")
}
