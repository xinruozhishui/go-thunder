package common

import (
	"os/user"
	"strconv"
	"os"
)

func GetDbPath() string  {
	u, _ := user.Current()
	st := strconv.QuoteRune(os.PathSeparator)
	st = st[1 : len(st)-1]
	return u.HomeDir + st + DbName
}