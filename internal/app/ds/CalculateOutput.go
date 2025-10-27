package ds

import (
	"fmt"
)

func CalculateProductionOutput(foundDefects int) string {
	const (
		DefectiveRate = 0.3
		FoundRate     = 0.06
	)
	// Чтобы избежать деления на ноль, если константы вдруг изменятся
	if DefectiveRate == 0 || FoundRate == 0 {
		return "0 шт."
	}

	calculatedOutput := float64(foundDefects) / (DefectiveRate * FoundRate)
	calculatedOutputInt := int(calculatedOutput)

	return fmt.Sprintf("%d шт.", calculatedOutputInt)
}
