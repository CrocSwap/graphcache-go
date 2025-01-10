package tables

import "strings"

func translateUser(r string, txHash string) string {
	if strings.ToLower(r) == "0x2be293361aea6136a42036ef68ff248fc379b4f8" {
		return "0x888d768764a2e304215247f0ba3457ccb0f0ab4f"
	}
	if txHash == "0x1073ce18cb08b0147a8f150ed37401cc9fd01eb0b99478923815877b8ac58d18" {
		return "0xbb42ae1ee6d201cc96c58622af4e22e15bc93d12"
	}
	return r
}
