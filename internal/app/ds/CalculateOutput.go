package ds

import "fmt"

// ФОРМУЛА ДЛЯ РАСЧЕТА ОБЪЕМА ПРОИЗВОДСТВА ПО КОЛИЧЕСТВУ ДЕФЕКТОВ

func CalculateProductionOutput(foundDefects int) string {
	const (
		DefectiveRate = 0.3
		FoundRate     = 0.06
	)

	// Расчет выпуск
	calculatedOutput := float64(foundDefects) / (DefectiveRate * FoundRate)

	// Выпуск не может быть отрицательным
	if calculatedOutput < 0 {
		calculatedOutput = 0
	}

	return fmt.Sprintf("%f шт.", calculatedOutput)
}
