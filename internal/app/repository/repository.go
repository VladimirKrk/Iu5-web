package repository

import (
	"fmt"
	"strings"
)

const MINIO_URL = "http://localhost:9000/vlk-images/"

type Service struct {
	ID               int
	Name             string
	Description      string
	ShortDescription string
	Century          string
	ImageKey         string
	ImageURL         string
	ExtraImageKey    string
	ExtraImageURL    string
}

// remove extra fields, everything else will be added
// from the previous struct
type ApplicationItem struct {
	ServiceID       int
	FoundDefects    int
	PredictedOutput string
}

type Application struct {
	ID    int
	Items []ApplicationItem
}

type Repository struct {
	services    []Service
	application Application
}

func NewRepository() (*Repository, error) {
	services := []Service{
		{ID: 1, Name: "Фарфорный завод", Description: "Создание фарфоровых кухонных принадлежностей и индивидуальных заказов.", ShortDescription: "Создание фарфоровых кухонных принадлежностей и индивидуальных заказов.", Century: "XIX век", ImageKey: "Cup.png", ExtraImageKey: "extra_cup.png"},
		{ID: 2, Name: "Костромская кузница", Description: "Круглогодичная ковка предметов быта из металла.", ShortDescription: "Круглогодичная ковка предметов быта из металла.", Century: "XIX век", ImageKey: "horse_shoe.png", ExtraImageKey: "extra_horse_shoe.png"},
		{ID: 3, Name: "Мастерская плетения прута", Description: "Плетение корзин из ивового прута.", ShortDescription: "Плетение корзин из ивового прута.", Century: "XX век", ImageKey: "basket.png", ExtraImageKey: "extra_basket.png"},
	}

	application := Application{
		ID: 101,
		Items: []ApplicationItem{
			// ЗАПОЛНЯЕМ НОВЫЕ ПОЛЯ
			{
				ServiceID:       1,
				FoundDefects:    80,
				PredictedOutput: "5000 шт.",
			},
			{
				ServiceID:       2,
				FoundDefects:    80,
				PredictedOutput: "1000 шт.",
			},
		},
	}

	return &Repository{services: services, application: application}, nil
}

// ... остальной код файла остается без изменений ...
func (r *Repository) GetServices() ([]Service, error)      { return r.services, nil }
func (r *Repository) GetApplication() (Application, error) { return r.application, nil }
func (r *Repository) GetServiceByID(id int) (Service, error) {
	for _, service := range r.services {
		if service.ID == id {
			return service, nil
		}
	}
	return Service{}, fmt.Errorf("услуга с id %d не найдена", id)
}
func (r *Repository) GetServicesByName(name string) ([]Service, error) {
	var result []Service
	for _, service := range r.services {
		if strings.Contains(strings.ToLower(service.Name), strings.ToLower(name)) {
			result = append(result, service)
		}
	}
	return result, nil
}
