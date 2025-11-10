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
		db.Service.Status.Set(db.ServiceStatus(data.Status)),
		db.Service.RdateStart.Set(data.ReserveDateStart),
		db.Service.RdateEnd.Set(data.ReserveDateEnd),
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

	if data.Status != nil {
		updates = append(updates, db.Service.Status.Set(db.ServiceStatus(*data.Status)))
	}
	if data.ReserveDateStart != nil {
		updates = append(updates, db.Service.RdateStart.Set(*data.ReserveDateStart))
	}
	if data.ReserveDateEnd != nil {
		updates = append(updates, db.Service.RdateEnd.Set(*data.ReserveDateEnd))
	}
	if data.OwnerID != nil {
		updates = append(updates, db.Service.Owner.Link(db.Owner.UserID.Equals(*data.OwnerID)))
	}
	if data.PetID != nil {
		updates = append(updates, db.Service.Pet.Link(db.Pet.Petid.Equals(*data.PetID)))
	}

	if len(updates) == 0 {
		// No direct Service table fields to update. Fetch current service and
		// return it so the caller (service layer) can proceed to update the
		// related subservice (cservice/mservice) fields like comment/score.
		current, err := repo.Collection.Service.FindUnique(
			db.Service.Sid.Equals(serviceID),
		).With(
			db.Service.Cservice.Fetch(),
			db.Service.Mservice.Fetch(),
		).Exec(repo.Context)
		if err != nil {
			return nil, fmt.Errorf("service -> UpdateByID: %v", err)
		}
		if current == nil {
			return nil, fmt.Errorf("service -> UpdateByID: service not found")
		}

		result := mapServiceModel(current)
		// Merge any subservice-related incoming values so caller can use them
		// when updating the subservice.
		if data.StaffID != nil {
			result.StaffID = *data.StaffID
		}
		result.Disease = data.Disease
		result.Comment = data.Comment
		result.Score = data.Score

		return result, nil
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
	params = addServiceStatusParams(params, status)
	if month > 0 && year > 0 {
		limit = 31
		params = addRDateRangeParams(params, month, year)
	}
	services, err := repo.Collection.Service.FindMany(params...).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).OrderBy(
		db.Service.RdateStart.Order(db.SortOrderAsc),
	).Skip(offset).Take(limit).Exec(repo.Context)

	if err != nil {
		return nil, err
	}

	return filterUniqueDays(services, month, year)
}

func (repo *serviceRepository) FindByDoctorID(doctorID string, status string, month, year, offset int, limit int) ([]*entities.ServiceModel, error) {
	params := []db.ServiceWhereParam{
		db.Service.Mservice.Where(
			db.Mservice.Did.Equals(doctorID),
		),
	}
	params = addServiceStatusParams(params, status)
	if month > 0 && year > 0 {
		limit = 31
		params = addRDateRangeParams(params, month, year)
	}
	services, err := repo.Collection.Service.FindMany(params...).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).OrderBy(
		db.Service.RdateStart.Order(db.SortOrderAsc),
	).Skip(offset).Take(limit).Exec(repo.Context)

	if err != nil {
		return nil, err
	}

	return filterUniqueDays(services, month, year)
}

func (repo *serviceRepository) FindByCaretakerID(caretakerID string, status string, month, year, offset int, limit int) ([]*entities.ServiceModel, error) {
	params := []db.ServiceWhereParam{
		db.Service.Cservice.Where(
			db.Cservice.Cid.Equals(caretakerID),
		),
	}
	params = addServiceStatusParams(params, status)
	if month > 0 && year > 0 {
		limit = 31
		params = addRDateRangeParams(params, month, year)
	}
	services, err := repo.Collection.Service.FindMany(params...).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).OrderBy(
		db.Service.RdateStart.Order(db.SortOrderAsc),
	).Skip(offset).Take(limit).Exec(repo.Context)

	if err != nil {
		return nil, err
	}

	return filterUniqueDays(services, month, year)
}

func (repo *serviceRepository) FindAll(status string, month, year, offset int, limit int) ([]*entities.ServiceModel, error) {
	params := []db.ServiceWhereParam{}

	params = addServiceStatusParams(params, status)
	if month > 0 && year > 0 {
		limit = 31
		params = addRDateRangeParams(params, month, year)
	}

	services, err := repo.Collection.Service.FindMany(params...).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
	).OrderBy(
		db.Service.RdateStart.Order(db.SortOrderAsc),
	).Skip(offset).Take(limit).Exec(repo.Context)
	if err != nil {
		return nil, err
	}

	return filterUniqueDays(services, month, year)
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
		Sid:              model.Sid,
		OwnerID:          model.Oid,
		PetID:            model.Petid,
		PaymentID:        model.Payid,
		Status:           model.Status,
		ReserveDateStart: model.RdateStart,
		ReserveDateEnd:   model.RdateEnd,
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

func filterUniqueDays(services []db.ServiceModel, month, year int) ([]*entities.ServiceModel, error) {
	var result []*entities.ServiceModel
	var serviceAdded bool
	if month > 0 && year > 0 {
		uniqueDays := make(map[int]bool)
		for i := range services {
			serviceAdded = false
			if services[i].RdateStart.Year() == year && services[i].RdateStart.Month() == time.Month(month) {
				startDay := services[i].RdateStart.Day()
				if !uniqueDays[startDay] {
					result = append(result, mapServiceModel(&services[i]))
					serviceAdded = true
					uniqueDays[startDay] = true
				}
			}
			if services[i].RdateEnd.Year() == year && services[i].RdateEnd.Month() == time.Month(month) {
				endDay := services[i].RdateEnd.Day()
				if !uniqueDays[endDay] {
					if !serviceAdded {
						result = append(result, mapServiceModel(&services[i]))
					}
					uniqueDays[endDay] = true
				}
			}
		}
	} else {
		for i := range services {
			result = append(result, mapServiceModel(&services[i]))
		}
	}
	return result, nil
}

func addServiceStatusParams(params []db.ServiceWhereParam, status string) []db.ServiceWhereParam {
	if status != "" && status != "all" {
		if serviceStatus, ok := toServiceStatus(status); ok {
			params = append(params, db.Service.Status.Equals(serviceStatus))
		}
	}
	return params
}

func addRDateRangeParams(params []db.ServiceWhereParam, month, year int) []db.ServiceWhereParam {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	params = append(params,
		db.Service.Or(
			db.Service.And(
				db.Service.RdateStart.Gte(startDate),
				db.Service.RdateStart.Lt(endDate),
			),
			db.Service.And(
				db.Service.RdateEnd.Gte(startDate),
				db.Service.RdateEnd.Lt(endDate),
			),
		),
	)
	return params
}
