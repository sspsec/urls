package common

import "flag"

func Flag(Info *FofaQuery) {
	flag.StringVar(&UrlFile, "f", "", "urls file default url.txt")
	flag.StringVar(&Info.Query, "ffq", "", "fofa query field")
	flag.IntVar(&Info.Page, "p", 10, "fofa query page")
	flag.Parse()
}
