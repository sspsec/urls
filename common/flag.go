package common

import "flag"

func Flag() {
	flag.StringVar(&UrlFile, "f", "", "urls file default url.txt")
	flag.BoolVar(&OutputToFile, "o", false, "Output results to file")
	flag.Parse()
}
