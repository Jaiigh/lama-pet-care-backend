package repositories

import (
	"context"
	"fmt"

	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
)

type cserviceRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type ICServiceRepository interface {
	Insert(data entities.SubService) (*entities.SubService, error)
	FindByID(serviceID string) (*entities.SubService, error)
	DeleteByID(serviceID string) (*entities.SubService, error)
	UpdateByID(data entities.SubService) (*entities.SubService, error)
	FindByCaretakerID(caretakerID string) ([]*entities.SubService, error)
	FindAll() ([]*entities.SubService, error)
}

func NewCServiceRepository(db *ds.PrismaDB) ICServiceRepository {
	return &cserviceRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *cserviceRepository) Insert(data entities.SubService) (*entities.SubService, error) {
	createdCService, err := repo.Collection.Cservice.CreateOne(
		db.Cservice.Score.Set(0),
		db.Cservice.Caretaker.Link(db.Caretaker.UserID.Equals(data.StaffID)),
		db.Cservice.Service.Link(db.Service.Sid.Equals(data.ServiceID)),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("cservice -> Insert: %w", err)
	}
	return mapCserviceToSubService(createdCService), nil
}

func (repo *cserviceRepository) FindByID(serviceID string) (*entities.SubService, error) {
	cservice, err := repo.Collection.Cservice.FindUnique(
		db.Cservice.Sid.Equals(serviceID),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("cservice -> FindByID: %w", err)
	}

	return mapCserviceToSubService(cservice), nil
}

func (repo *cserviceRepository) DeleteByID(serviceID string) (*entities.SubService, error) {
	deletedService, err := repo.Collection.Cservice.FindUnique(
		db.Cservice.Sid.Equals(serviceID),
	).Delete().Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("cservice -> DeleteByID: %w", err)
	}
	if deletedService == nil {
		return nil, fmt.Errorf("cservice -> DeleteByID: cservice not found")
	}

	return mapCserviceToSubService(deletedService), nil
}

func (repo *cserviceRepository) UpdateByID(data entities.SubService) (*entities.SubService, error) {
	updates := []db.CserviceSetParam{}
	if data.StaffID != "" {
		updates = append(updates, db.Cservice.Cid.Set(data.StaffID))
	}
	if data.Comment != nil {
		updates = append(updates, db.Cservice.Comment.Set(*data.Comment))
	}
	if data.Score != nil {
		updates = append(updates, db.Cservice.Score.Set(*data.Score))
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("cservice -> UpdateByID: no fields to update")
	}

	// Execute update
	updatedCService, err := repo.Collection.Cservice.FindUnique(
		db.Cservice.Sid.Equals(data.ServiceID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("cservice -> UpdateByID: %v", err)
	}
	if updatedCService == nil {
		return nil, fmt.Errorf("cservice -> UpdateByID: cservice not found")
	}

	return mapCserviceToSubService(updatedCService), nil
}

func (repo *cserviceRepository) FindByCaretakerID(caretakerID string) ([]*entities.SubService, error) {
	cservices, err := repo.Collection.Cservice.FindMany(
		db.Cservice.Cid.Equals(caretakerID),
	).Exec(repo.Context)

	if err != nil {
		return nil, err
	}
	var result []*entities.SubService
	for i := range cservices {
		result = append(result, mapCserviceToSubService(&cservices[i]))
	}
	return result, nil
}

func (repo *cserviceRepository) FindAll() ([]*entities.SubService, error) {
	cservices, err := repo.Collection.Cservice.FindMany().Exec(repo.Context)
	if err != nil {
		return nil, err
	}
	var result []*entities.SubService
	for i := range cservices {
		result = append(result, mapCserviceToSubService(&cservices[i]))
	}
	return result, nil
}

func mapCserviceToSubService(model *db.CserviceModel) *entities.SubService {
	comment, ok := model.Comment()
	if !ok {
		comment = ""
	}
	return &entities.SubService{
		ServiceID: model.Sid,
		StaffID:   model.Cid,
		Comment:   &comment,
		Score:     &model.Score,
	}
}
