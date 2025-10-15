package repositories

import (
	"context"
	"fmt"

	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
)

type serviceRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type IServiceRepository interface {
	Insert(data entities.CreateServiceRequest) (*entities.ServiceModel, error)
	FindByID(serviceID string) (*entities.ServiceModel, error)
	DeleteByID(serviceID string) (*entities.ServiceModel, error)
	UpdateByID(serviceID string, data entities.UpdateServiceRequest) (*entities.ServiceModel, error)
	FindByOwnerID(ownerID string, status string) ([]*entities.ServiceModel, error)
    FindAll(status string) ([]*entities.ServiceModel, error)
}

func NewServiceRepository(db *ds.PrismaDB) IServiceRepository {
	return &serviceRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *serviceRepository) Insert(data entities.CreateServiceRequest) (*entities.ServiceModel, error) {
	createdService, err := repo.Collection.Service.CreateOne(
		db.Service.Price.Set(data.Price),
		db.Service.Status.Set(db.ServiceStatus(data.Status)),
		db.Service.Rdate.Set(data.ReserveDate),
		db.Service.Owner.Link(db.Owner.UserID.Equals(data.OwnerID)),
		db.Service.Payment.Link(db.Payment.Payid.Equals(data.PaymentID)),
		db.Service.Pet.Link(db.Pet.Petid.Equals(data.PetID)),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("service -> Insert: %w", err)
	}

	return mapServiceModel(createdService), nil
}

func (repo *serviceRepository) FindByID(serviceID string) (*entities.ServiceModel, error) {
	service, err := repo.Collection.Service.FindUnique(
		db.Service.Sid.Equals(serviceID),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("service -> FindByID: %w", err)
	}
	if service == nil {
		return nil, fmt.Errorf("service -> FindByID: service not found")
	}

	return mapServiceModel(service), nil
}

func (repo *serviceRepository) DeleteByID(serviceID string) (*entities.ServiceModel, error) {
	deletedService, err := repo.Collection.Service.FindUnique(
		db.Service.Sid.Equals(serviceID),
	).Delete().Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("service -> DeleteByID: %w", err)
	}
	if deletedService == nil {
		return nil, fmt.Errorf("service -> DeleteByID: service not found")
	}

	return mapServiceModel(deletedService), nil
}

func (repo *serviceRepository) UpdateByID(serviceID string, data entities.UpdateServiceRequest) (*entities.ServiceModel, error) {
	updates := []db.ServiceSetParam{}

	if data.Price != nil {
		updates = append(updates, db.Service.Price.Set(*data.Price))
	}
	if data.Status != nil {
		updates = append(updates, db.Service.Status.Set(db.ServiceStatus(*data.Status)))
	}
	if data.ReserveDate != nil {
		updates = append(updates, db.Service.Rdate.Set(*data.ReserveDate))
	}
	if data.OwnerID != nil {
		updates = append(updates, db.Service.Owner.Link(db.Owner.UserID.Equals(*data.OwnerID)))
	}
	if data.PetID != nil {
		updates = append(updates, db.Service.Pet.Link(db.Pet.Petid.Equals(*data.PetID)))
	}
	if data.PaymentID != nil {
		updates = append(updates, db.Service.Payment.Link(db.Payment.Payid.Equals(*data.PaymentID)))
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("service -> UpdateByID: no fields to update")
	}

	updatedService, err := repo.Collection.Service.FindUnique(
		db.Service.Sid.Equals(serviceID),
	).Update(updates...).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("service -> UpdateByID: %w", err)
	}
	if updatedService == nil {
		return nil, fmt.Errorf("service -> UpdateByID: service not found")
	}

	return mapServiceModel(updatedService), nil
}

func mapServiceModel(model *db.ServiceModel) *entities.ServiceModel {
	return &entities.ServiceModel{
		Sid:         model.Sid,
		OwnerID:     model.Oid,
		PetID:       model.Petid,
		PaymentID:   model.Payid,
		Price:       model.Price,
		Status:      model.Status,
		ReserveDate: model.Rdate,
	}
}

func toServiceStatus(s string) (db.ServiceStatus, bool) {
	switch s {
	case "wait":
		return db.ServiceStatusWait, true
	case "ongoing":
		return db.ServiceStatusOngoing, true
	case "finish":
		return db.ServiceStatusFinish, true
	default:
		return "", false // Return an empty value and false if the string is not a valid status
	}
}

func (repo *serviceRepository) FindByOwnerID(ownerID string, status string) ([]*entities.ServiceModel, error) {
	params := []db.ServiceWhereParam{
        db.Service.Oid.Equals(ownerID),
    }
	if status != "" && status != "all" {
        
		if serviceStatus, ok := toServiceStatus(status); ok {
            
            params = append(params, db.Service.Status.Equals(serviceStatus))
        }
    }
	services, err := repo.Collection.Service.FindMany(params...).Exec(repo.Context)
    
    if err != nil {
        return nil, err
    }
    var result []*entities.ServiceModel
    for _, s := range services {
        result = append(result, mapServiceModel(&s))
    }
    return result, nil
}

func (repo *serviceRepository) FindAll(status string) ([]*entities.ServiceModel, error) {
	 params := []db.ServiceWhereParam{}

    
    if status != "" && status != "all" {
        if serviceStatus, ok := toServiceStatus(status); ok {
            params = append(params, db.Service.Status.Equals(serviceStatus))
        }
    }
    
	services, err := repo.Collection.Service.FindMany(params...).Exec(repo.Context)
    if err != nil {
        return nil, err
    }
    var result []*entities.ServiceModel
    for _, s := range services {
        result = append(result, mapServiceModel(&s))
    }
    return result, nil
}