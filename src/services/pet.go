package services

import (
	"lama-backend/domain/entities"
	"lama-backend/domain/repositories"
)

type PetService struct {
	PetRepository repositories.IPetRepository
}

type IPetService interface {
	InsertPet(data entities.CreatedPetModel) (*entities.PetDataModel, error)
	FindByOwnerID(ownerID string) ([]entities.PetDataModel, error)
	FindAll() ([]entities.PetDataModel, error)
	UpdatePet(petID string, data entities.UpdatePetModel) (*entities.PetDataModel, error)
	DeletePet(petID string) (*entities.PetDataModel, error)
}

func NewPetService(petRepo repositories.IPetRepository) IPetService {
	return &PetService{
		PetRepository: petRepo,
	}
}

func (s *PetService) InsertPet(data entities.CreatedPetModel) (*entities.PetDataModel, error) {
	return s.PetRepository.InsertPet(data)
}

func (s *PetService) FindByOwnerID(ownerID string) ([]entities.PetDataModel, error) {
	return s.PetRepository.FindByOwnerID(ownerID)
}

func (s *PetService) FindAll() ([]entities.PetDataModel, error) {
	return s.PetRepository.FindAll()
}

func (s *PetService) UpdatePet(petID string, data entities.UpdatePetModel) (*entities.PetDataModel, error) {
	return s.PetRepository.UpdatePet(petID, data)
}

func (s *PetService) DeletePet(petID string) (*entities.PetDataModel, error) {
	return s.PetRepository.DeletePet(petID)
}
