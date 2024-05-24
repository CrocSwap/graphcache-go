package tables

import "strings"

func translateUser(r string) string {
	if strings.ToLower(r) == "0x2be293361aea6136a42036ef68ff248fc379b4f8" {
		return "0x888d768764a2e304215247f0ba3457ccb0f0ab4f"
	}
	return r
}
