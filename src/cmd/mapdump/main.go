// Diagnostic: decrypt the first bytes of a Nox map with the exact same
// cryptfile path the engine uses, to compare output across architectures.
package main

import (
	"encoding/hex"
	"fmt"
	"os"

	crypt "github.com/opennox/noxcrypt"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: mapdump <file.map>")
		os.Exit(1)
	}
	if err := cryptfile.OpenGlobal(os.Args[1], cryptfile.ReadOnly, crypt.MapKey); err != nil {
		panic(err)
	}
	cf := cryptfile.Global()
	buf := make([]byte, 512)
	if _, err := cf.ReadWrite(buf); err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(buf))
}
