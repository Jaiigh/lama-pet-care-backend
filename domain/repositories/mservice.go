package repositories

import (
	"context"
	"fmt"

	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
)

type mserviceRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type IMServiceRepository interface {
	Insert(data entities.SubService) (*entities.SubService, error)
	FindByID(serviceID string) (*entities.SubService, error)
	DeleteByID(serviceID string) (*entities.SubService, error)
	UpdateByID(data entities.SubService) (*entities.SubService, error)
	FindByDoctorID(doctorID string) ([]*entities.SubService, error)
	FindAll() ([]*entities.SubService, error)
}

func NewMServiceRepository(db *ds.PrismaDB) IMServiceRepository {
	return &mserviceRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *mserviceRepository) Insert(data entities.SubService) (*entities.SubService, error) {
	createdMService, err := repo.Collection.Mservice.CreateOne(
		db.Mservice.Disease.Set(*data.Disease),
		db.Mservice.Service.Link(db.Service.Sid.Equals(data.ServiceID)),
		db.Mservice.Doctor.Link(db.Doctor.UserID.Equals(data.StaffID)),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("mservice -> Insert: %w", err)
	}

	return mapMserviceToSubService(createdMService), nil
}

func (repo *mserviceRepository) FindByID(serviceID string) (*entities.SubService, error) {
	mservice, err := repo.Collection.Mservice.FindUnique(
		db.Mservice.Sid.Equals(serviceID),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("mservice -> FindByID: %w", err)
	}

	return mapMserviceToSubService(mservice), nil
}

func (repo *mserviceRepository) DeleteByID(serviceID string) (*entities.SubService, error) {
	deletedService, err := repo.Collection.Mservice.FindUnique(
		db.Mservice.Sid.Equals(serviceID),
	).Delete().Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("mservice -> DeleteByID: %w", err)
	}
	if deletedService == nil {
		return nil, fmt.Errorf("mservice -> DeleteByID: mservice not found")
	}

	return mapMserviceToSubService(deletedService), nil
}

func (repo *mserviceRepository) UpdateByID(data entities.SubService) (*entities.SubService, error) {
	updates := []db.MserviceSetParam{}
	if data.StaffID != "" {
		updates = append(updates, db.Mservice.Did.Set(data.StaffID))
	}
	if data.Disease != nil {
		updates = append(updates, db.Mservice.Disease.Set(*data.Disease))
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("mservice -> UpdateByID: no fields to update")
	}

	// Execute update
	updatedMService, err := repo.Collection.Mservice.FindUnique(
		db.Mservice.Sid.Equals(data.ServiceID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("mservice -> UpdateByID: %v", err)
	}
	if updatedMService == nil {
		return nil, fmt.Errorf("mservice -> UpdateByID: mservice not found")
	}

	return mapMserviceToSubService(updatedMService), nil
}

func (repo *mserviceRepository) FindByDoctorID(doctorID string) ([]*entities.SubService, error) {
	mservices, err := repo.Collection.Mservice.FindMany(
		db.Mservice.Did.Equals(doctorID),
	).Exec(repo.Context)

	if err != nil {
		return nil, err
	}
	var result []*entities.SubService
	for i := range mservices {
		result = append(result, mapMserviceToSubService(&mservices[i]))
	}
	return result, nil
}

func (repo *mserviceRepository) FindAll() ([]*entities.SubService, error) {
	mservices, err := repo.Collection.Mservice.FindMany().Exec(repo.Context)
	if err != nil {
		return nil, err
	}
	var result []*entities.SubService
	for i := range mservices {
		result = append(result, mapMserviceToSubService(&mservices[i]))
	}
	return result, nil
}

func mapMserviceToSubService(model *db.MserviceModel) *entities.SubService {
	doctorID, _ := model.Did()

	return &entities.SubService{
		ServiceID: model.Sid,
		StaffID:   doctorID,
		Disease:   &model.Disease,
	}
}
