package repositories

import (
	"context"
	"errors"
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

	result := mapServiceModel(createdService)
	result.ServiceType = data.ServiceType
	result.StaffID = data.StaffID
	result.Disease = data.Disease
	result.Comment = data.Comment
	result.Score = nil

	return result, nil
}

func (repo *serviceRepository) FindByID(serviceID string) (*entities.ServiceModel, error) {
	service, err := repo.Collection.Service.FindUnique(
		db.Service.Sid.Equals(serviceID),
	).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("service -> FindByID: %w", err)
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
	serviceRecord, err := repo.Collection.Service.FindUnique(
		db.Service.Sid.Equals(serviceID),
	).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("service -> UpdateByID: %w", err)
	}

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

	if len(updates) > 0 {
		if _, err := repo.Collection.Service.FindUnique(
			db.Service.Sid.Equals(serviceID),
		).Update(updates...).Exec(repo.Context); err != nil {
			return nil, fmt.Errorf("service -> UpdateByID: %w", err)
		}
	}

	desiredType := ""
	if data.ServiceType != nil && *data.ServiceType != "" {
		desiredType = *data.ServiceType
	}

	existingType := ""
	if _, ok := serviceRecord.Cservice(); ok {
		existingType = "cservice"
	}
	if _, ok := serviceRecord.Mservice(); ok {
		existingType = "mservice"
	}

	if desiredType == "" {
		desiredType = existingType
	}

	switch desiredType {
	case "cservice":
		if existingType == "mservice" {
			if _, err := repo.Collection.Mservice.FindUnique(
				db.Mservice.Sid.Equals(serviceID),
			).Delete().Exec(repo.Context); err != nil {
				return nil, fmt.Errorf("service -> UpdateByID: %w", err)
			}
		}

		if _, err := repo.Collection.Cservice.FindUnique(
			db.Cservice.Sid.Equals(serviceID),
		).Exec(repo.Context); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				if data.StaffID == nil {
					return nil, fmt.Errorf("service -> UpdateByID: staff_id required when switching to cservice")
				}

				optional := []db.CserviceSetParam{}
				if data.Comment != nil {
					optional = append(optional, db.Cservice.Comment.SetOptional((*db.String)(data.Comment)))
				}

				if _, err := repo.Collection.Cservice.CreateOne(
					db.Cservice.Score.Set(0),
					db.Cservice.Caretaker.Link(db.Caretaker.UserID.Equals(*data.StaffID)),
					db.Cservice.Service.Link(db.Service.Sid.Equals(serviceID)),
					optional...,
				).Exec(repo.Context); err != nil {
					return nil, fmt.Errorf("service -> UpdateByID: %w", err)
				}
			} else {
				return nil, fmt.Errorf("service -> UpdateByID: %w", err)
			}
		} else {
			cUpdates := []db.CserviceSetParam{}
			if data.StaffID != nil {
				cUpdates = append(cUpdates, db.Cservice.Caretaker.Link(db.Caretaker.UserID.Equals(*data.StaffID)))
			}
			if data.Comment != nil {
				cUpdates = append(cUpdates, db.Cservice.Comment.SetOptional((*db.String)(data.Comment)))
			}
			if len(cUpdates) > 0 {
				if _, err := repo.Collection.Cservice.FindUnique(
					db.Cservice.Sid.Equals(serviceID),
				).Update(cUpdates...).Exec(repo.Context); err != nil {
					return nil, fmt.Errorf("service -> UpdateByID: %w", err)
				}
			}
		}

	case "mservice":
		if existingType == "cservice" {
			if _, err := repo.Collection.Cservice.FindUnique(
				db.Cservice.Sid.Equals(serviceID),
			).Delete().Exec(repo.Context); err != nil {
				return nil, fmt.Errorf("service -> UpdateByID: %w", err)
			}
		}

		if _, err := repo.Collection.Mservice.FindUnique(
			db.Mservice.Sid.Equals(serviceID),
		).Exec(repo.Context); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				if data.StaffID == nil || data.Disease == nil {
					return nil, fmt.Errorf("service -> UpdateByID: staff_id and disease required when switching to mservice")
				}

				if _, err := repo.Collection.Mservice.CreateOne(
					db.Mservice.Disease.Set(*data.Disease),
					db.Mservice.Service.Link(db.Service.Sid.Equals(serviceID)),
					db.Mservice.Doctor.Link(db.Doctor.UserID.Equals(*data.StaffID)),
				).Exec(repo.Context); err != nil {
					return nil, fmt.Errorf("service -> UpdateByID: %w", err)
				}
			} else {
				return nil, fmt.Errorf("service -> UpdateByID: %w", err)
			}
		} else {
			mUpdates := []db.MserviceSetParam{}
			if data.StaffID != nil {
				mUpdates = append(mUpdates, db.Mservice.Doctor.Link(db.Doctor.UserID.Equals(*data.StaffID)))
			}
			if data.Disease != nil {
				mUpdates = append(mUpdates, db.Mservice.Disease.Set(*data.Disease))
			}
			if len(mUpdates) > 0 {
				if _, err := repo.Collection.Mservice.FindUnique(
					db.Mservice.Sid.Equals(serviceID),
				).Update(mUpdates...).Exec(repo.Context); err != nil {
					return nil, fmt.Errorf("service -> UpdateByID: %w", err)
				}
			}
		}

	case "":
		// no related service type to update
	default:
		return nil, fmt.Errorf("service -> UpdateByID: unsupported service_type %q", desiredType)
	}

	updatedService, err := repo.Collection.Service.FindUnique(
		db.Service.Sid.Equals(serviceID),
	).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("service -> UpdateByID: %w", err)
	}

	return mapServiceModel(updatedService), nil
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
	services, err := repo.Collection.Service.FindMany(params...).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).Exec(repo.Context)

	if err != nil {
		return nil, err
	}
	var result []*entities.ServiceModel
	for i := range services {
		result = append(result, mapServiceModel(&services[i]))
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

	services, err := repo.Collection.Service.FindMany(params...).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).Exec(repo.Context)
	if err != nil {
		return nil, err
	}
	var result []*entities.ServiceModel
	for i := range services {
		result = append(result, mapServiceModel(&services[i]))
	}
	return result, nil
}

func mapServiceModel(model *db.ServiceModel) *entities.ServiceModel {
	result := &entities.ServiceModel{
		Sid:         model.Sid,
		OwnerID:     model.Oid,
		PetID:       model.Petid,
		PaymentID:   model.Payid,
		Price:       model.Price,
		Status:      model.Status,
		ReserveDate: model.Rdate,
	}

	if cservice, ok := model.Cservice(); ok {
		result.ServiceType = "cservice"
		result.StaffID = cservice.Cid

		if comment, ok := cservice.Comment(); ok {
			commentStr := string(comment)
			result.Comment = &commentStr
		}
	} else if mservice, ok := model.Mservice(); ok {
		result.ServiceType = "mservice"
		if did, ok := mservice.Did(); ok {
			result.StaffID = string(did)
		}
		result.Disease = &mservice.Disease
	}

	return result
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
