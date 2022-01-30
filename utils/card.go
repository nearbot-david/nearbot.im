package utils

import "fmt"

func MaskCard(rawNumber string) string {
	return fmt.Sprintf("%s %s %s %s", rawNumber[:4], "****", "****", rawNumber[len(rawNumber)-4:])
}
