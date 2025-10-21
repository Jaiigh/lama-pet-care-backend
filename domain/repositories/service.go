package repositories

import (
	"context"
	"fmt"
	"time"

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
	FindByOwnerID(ownerID string, status string, month, year, offset int, limit int) ([]*entities.ServiceModel, error)
	FindByDoctorID(doctorID string, status string, month, year, offset int, limit int) ([]*entities.ServiceModel, error)
	FindByCaretakerID(caretakerID string, status string, month, year, offset int, limit int) ([]*entities.ServiceModel, error)
	FindAll(status string, month, year, offset int, limit int) ([]*entities.ServiceModel, error)
	UpdateStatus(serviceID, status string) error
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

	if len(updates) == 0 {
		return nil, fmt.Errorf("service -> UpdateByID: no fields to update")
	}

	updatedService, err := repo.Collection.Service.FindUnique(
		db.Service.Sid.Equals(serviceID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("service -> UpdateByID: %v", err)
	}
	if updatedService == nil {
		return nil, fmt.Errorf("service -> UpdateByID: service not found")
	}

	result := mapServiceModel(updatedService)
	if data.StaffID != nil {
		tmpResult, _ := repo.FindByID(serviceID)
		result.StaffID = tmpResult.StaffID
	}
	result.Disease = data.Disease
	result.Comment = data.Comment
	result.Score = data.Score

	return result, nil
}

func (repo *serviceRepository) FindByOwnerID(ownerID string, status string, month, year, offset int, limit int) ([]*entities.ServiceModel, error) {
	params := []db.ServiceWhereParam{
		db.Service.Oid.Equals(ownerID),
	}
	if status != "" && status != "all" {
		if serviceStatus, ok := toServiceStatus(status); ok {
			params = append(params, db.Service.Status.Equals(serviceStatus))
		}
	}
	if month > 0 && year > 0 {
		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, 0)
		params = append(params, db.Service.Rdate.Gte(startDate))
		params = append(params, db.Service.Rdate.Lt(endDate))
	}
	services, err := repo.Collection.Service.FindMany(params...).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).OrderBy(
		db.Service.Rdate.Order(db.SortOrderAsc),
	).Skip(offset).Take(limit).Exec(repo.Context)

	if err != nil {
		return nil, err
	}
	var result []*entities.ServiceModel
	for i := range services {
		result = append(result, mapServiceModel(&services[i]))
	}
	return result, nil
}

func (repo *serviceRepository) FindByDoctorID(doctorID string, status string, month, year, offset int, limit int) ([]*entities.ServiceModel, error) {
	params := []db.ServiceWhereParam{
		db.Service.Mservice.Where(
			db.Mservice.Did.Equals(doctorID),
		),
	}
	if status != "" && status != "all" {
		if serviceStatus, ok := toServiceStatus(status); ok {
			params = append(params, db.Service.Status.Equals(serviceStatus))
		}
	}
	if month > 0 && year > 0 {
		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, 0)
		params = append(params, db.Service.Rdate.Gte(startDate))
		params = append(params, db.Service.Rdate.Lt(endDate))
	}
	services, err := repo.Collection.Service.FindMany(params...).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).OrderBy(
		db.Service.Rdate.Order(db.SortOrderAsc),
	).Skip(offset).Take(limit).Exec(repo.Context)

	if err != nil {
		return nil, err
	}
	var result []*entities.ServiceModel
	for i := range services {
		result = append(result, mapServiceModel(&services[i]))
	}
	return result, nil
}

func (repo *serviceRepository) FindByCaretakerID(caretakerID string, status string, month, year, offset int, limit int) ([]*entities.ServiceModel, error) {
	params := []db.ServiceWhereParam{
		db.Service.Cservice.Where(
			db.Cservice.Cid.Equals(caretakerID),
		),
	}
	if status != "" && status != "all" {
		if serviceStatus, ok := toServiceStatus(status); ok {
			params = append(params, db.Service.Status.Equals(serviceStatus))
		}
	}
	if month > 0 && year > 0 {
		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, 0)
		params = append(params, db.Service.Rdate.Gte(startDate))
		params = append(params, db.Service.Rdate.Lt(endDate))
	}
	services, err := repo.Collection.Service.FindMany(params...).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).OrderBy(
		db.Service.Rdate.Order(db.SortOrderAsc),
	).Skip(offset).Take(limit).Exec(repo.Context)

	if err != nil {
		return nil, err
	}
	var result []*entities.ServiceModel
	for i := range services {
		result = append(result, mapServiceModel(&services[i]))
	}
	return result, nil
}

func (repo *serviceRepository) FindAll(status string, month, year, offset int, limit int) ([]*entities.ServiceModel, error) {
	params := []db.ServiceWhereParam{}

	if status != "" && status != "all" {
		if serviceStatus, ok := toServiceStatus(status); ok {
			params = append(params, db.Service.Status.Equals(serviceStatus))
		}
	}
	if month > 0 && year > 0 {
		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, 0)
		params = append(params, db.Service.Rdate.Gte(startDate))
		params = append(params, db.Service.Rdate.Lt(endDate))
	}

	services, err := repo.Collection.Service.FindMany(params...).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).OrderBy(
		db.Service.Rdate.Order(db.SortOrderAsc),
	).Skip(offset).Take(limit).Exec(repo.Context)
	if err != nil {
		return nil, err
	}
	var result []*entities.ServiceModel
	for i := range services {
		result = append(result, mapServiceModel(&services[i]))
	}
	return result, nil
}

func (repo *serviceRepository) UpdateStatus(serviceID, status string) error {
	serviceStatus, ok := toServiceStatus(status)
	if !ok {
		return fmt.Errorf("service -> UpdateStatus: invalid status value")
	}

	updateStatus, err := repo.Collection.Service.FindUnique(
		db.Service.Sid.Equals(serviceID),
	).Update(
		db.Service.Status.Set(serviceStatus),
	).Exec(repo.Context)

	if err != nil {
		return fmt.Errorf("service -> UpdateStatus: %v", err)
	}
	if updateStatus == nil {
		return fmt.Errorf("service -> UpdateStatus: service not found")
	}

	return nil
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
