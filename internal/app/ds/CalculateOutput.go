package ds

import (
	"fmt"
)

func CalculateProductionOutput(foundDefects int) string {
	const (
		DefectiveRate = 0.3
		FoundRate     = 0.06
	)

	// Расчет выпуск
	calculatedOutput := float64(foundDefects) / (DefectiveRate * FoundRate)

	// Получаем целое число
	calculatedOutputInt := int(calculatedOutput)

	return fmt.Sprintf("%d шт.", calculatedOutputInt)
}
