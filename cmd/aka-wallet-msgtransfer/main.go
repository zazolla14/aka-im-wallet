package main

import (
	"github.com/1nterdigital/aka-im-tools/system/program"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/cmd"
)

func main() {
	if err := cmd.NewMsgTransferCmd().Exec(); err != nil {
		program.ExitWithError(err)
	}
}
