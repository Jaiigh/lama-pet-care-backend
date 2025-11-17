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
	FindByOwnerID(ownerID string, status string, month, year int, offset, limit int) ([]*entities.ServiceModel, int, error)
	FindByDoctorID(doctorID string, status string, month, year int, offset, limit int) ([]*entities.ServiceModel, int, error)
	FindByCaretakerID(caretakerID string, status string, month, year int, offset, limit int) ([]*entities.ServiceModel, int, error)
	FindAll(status string, month, year int, offset, limit int) ([]*entities.ServiceModel, int, error)
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

	return result, nil
}

func (repo *serviceRepository) FindByID(serviceID string) (*entities.ServiceModel, error) {
	service, err := repo.Collection.Service.FindUnique(
		db.Service.Sid.Equals(serviceID),
	).With(
		db.Service.Cservice.Fetch(),
		db.Service.Mservice.Fetch(),
		db.Service.Pet.Fetch(),
		db.Service.Payment.Fetch(),
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

func (repo *serviceRepository) FindByOwnerID(ownerID string, status string, month, year int, offset, limit int) ([]*entities.ServiceModel, int, error) {
	params := []db.ServiceWhereParam{
		db.Service.Oid.Equals(ownerID),
	}
	params = addServiceStatusParams(params, status)
	if month > 0 && year > 0 {
		limit = 31
		params = addRDateRangeParams(params, month, year)
	}

	var sqlResult []entities.CountResult
	sql, args, err := getSqlService("owner", ownerID, status, month, year)
	if err != nil {
		return nil, 0, err
	}
	err = repo.Collection.Prisma.QueryRaw(sql, args...).Exec(repo.Context, &sqlResult)
	if err != nil {
		return nil, 0, err
	}
	total := sqlResult[0].Count

	services, err := repo.Collection.Service.
		FindMany(params...).
		With(
			db.Service.Mservice.Fetch().With(
				db.Mservice.Doctor.Fetch().With(
					db.Doctor.Users.Fetch(),
				),
			),
			db.Service.Cservice.Fetch().With(
				db.Cservice.Caretaker.Fetch().With(
					db.Caretaker.Users.Fetch(),
				),
			),
			db.Service.Pet.Fetch(),
			db.Service.Payment.Fetch(),
		).
		OrderBy(
			db.Service.RdateStart.Order(db.SortOrderAsc),
		).
		Skip(offset).
		Take(limit).
		Exec(repo.Context)

	if err != nil {
		return nil, 0, err
	}

	results, err := filterUniqueDays(services, month, year)
	return results, total, err
}

func (repo *serviceRepository) FindByDoctorID(doctorID string, status string, month, year int, offset, limit int) ([]*entities.ServiceModel, int, error) {
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

	var sqlResult []entities.CountResult
	sql, args, err := getSqlService("doctor", doctorID, status, month, year)
	if err != nil {
		return nil, 0, err
	}
	err = repo.Collection.Prisma.QueryRaw(sql, args...).Exec(repo.Context, &sqlResult)
	if err != nil {
		return nil, 0, err
	}
	total := sqlResult[0].Count

	services, err := repo.Collection.Service.
		FindMany(params...).
		With(
			db.Service.Mservice.Fetch().With(
				db.Mservice.Doctor.Fetch().With(
					db.Doctor.Users.Fetch(),
				),
			),
			db.Service.Pet.Fetch(),
			db.Service.Payment.Fetch(),
		).
		OrderBy(
			db.Service.RdateStart.Order(db.SortOrderAsc),
		).
		Skip(offset).
		Take(limit).
		Exec(repo.Context)

	if err != nil {
		return nil, 0, err
	}

	result, err := filterUniqueDays(services, month, year)
	return result, total, err
}

func (repo *serviceRepository) FindByCaretakerID(caretakerID string, status string, month, year int, offset, limit int) ([]*entities.ServiceModel, int, error) {
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

	var sqlResult []entities.CountResult
	sql, args, err := getSqlService("caretaker", caretakerID, status, month, year)
	if err != nil {
		return nil, 0, err
	}
	err = repo.Collection.Prisma.QueryRaw(sql, args...).Exec(repo.Context, &sqlResult)
	if err != nil {
		return nil, 0, err
	}
	total := sqlResult[0].Count

	services, err := repo.Collection.Service.
		FindMany(params...).
		With(
			db.Service.Cservice.Fetch().With(
				db.Cservice.Caretaker.Fetch().With(
					db.Caretaker.Users.Fetch(),
				),
			),
			db.Service.Pet.Fetch(),
			db.Service.Payment.Fetch(),
		).
		OrderBy(
			db.Service.RdateStart.Order(db.SortOrderAsc),
		).
		Skip(offset).
		Take(limit).
		Exec(repo.Context)

	if err != nil {
		return nil, 0, err
	}

	result, err := filterUniqueDays(services, month, year)
	return result, total, err
}

func (repo *serviceRepository) FindAll(status string, month, year int, offset, limit int) ([]*entities.ServiceModel, int, error) {
	params := []db.ServiceWhereParam{}

	params = addServiceStatusParams(params, status)
	if month > 0 && year > 0 {
		limit = 31
		params = addRDateRangeParams(params, month, year)
	}

	var sqlResult []entities.CountResult
	sql, args, err := getSqlService("all", "", status, month, year)
	if err != nil {
		return nil, 0, err
	}
	err = repo.Collection.Prisma.QueryRaw(sql, args...).Exec(repo.Context, &sqlResult)
	if err != nil {
		return nil, 0, err
	}
	total := sqlResult[0].Count

	services, err := repo.Collection.Service.
		FindMany(params...).
		With(
			db.Service.Mservice.Fetch().With(
				db.Mservice.Doctor.Fetch().With(
					db.Doctor.Users.Fetch(),
				),
			),
			db.Service.Cservice.Fetch().With(
				db.Cservice.Caretaker.Fetch().With(
					db.Caretaker.Users.Fetch(),
				),
			),
			db.Service.Pet.Fetch(),
			db.Service.Payment.Fetch(),
		).
		OrderBy(
			db.Service.RdateStart.Order(db.SortOrderAsc),
		).
		Skip(offset).
		Take(limit).
		Exec(repo.Context)

	if err != nil {
		return nil, 0, err
	}

	result, err := filterUniqueDays(services, month, year)
	return result, total, err
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
		ShowId:           model.ShowID,
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
		result.Score = &cservice.Score

		if comment, ok := cservice.Comment(); ok {
			commentStr := string(comment)
			result.Comment = &commentStr
		}
	} else if mservice, ok := model.Mservice(); ok {
		result.ServiceType = "mservice"
		result.StaffID = mservice.Did
		disease, _ := mservice.Disease()
		result.Disease = &disease
	}

	return result
}

func addServiceAdditionModel(service *entities.ServiceModel, model *db.ServiceModel) *entities.ServiceModel {
	if cservice, ok := model.Cservice(); ok {
		caretaker := cservice.Caretaker()
		user := caretaker.Users()

		profile, _ := user.ProfileImage()
		specialization, _ := caretaker.Specialties()
		rating, _ := caretaker.Rating()
		service.Staff = entities.StaffCommonData{
			Role:            user.Role,
			Name:            user.Name,
			TelephoneNumber: user.TelephoneNumber,
			Profile:         profile,
			Specialization:  specialization,
			Rating:          rating,
		}
	} else if mservice, ok := model.Mservice(); ok {
		doctor := mservice.Doctor()
		user := doctor.Users()

		profile, _ := user.ProfileImage()
		service.Staff = entities.StaffCommonData{
			Role:            user.Role,
			Name:            user.Name,
			TelephoneNumber: user.TelephoneNumber,
			Profile:         profile,
			LicenseNumber:   doctor.LicenseNumber,
		}
	}

	pet := model.Pet()
	breed, _ := pet.Breed()
	name, _ := pet.Name()
	service.Pet = entities.PetCommonModel{
		Breed:     breed,
		Name:      name,
		BirthDate: pet.Birthdate,
		Weight:    pet.Weight,
		Kind:      pet.Kind,
		Sex:       pet.Sex,
	}

	payment := model.Payment()
	typeStr, _ := payment.Type()
	payDate, _ := payment.PayDate()
	service.Payment = entities.PaymentCommonModel{
		Status:  payment.Status,
		Price:   payment.Price,
		Type:    &typeStr,
		PayDate: &payDate,
	}

	return service
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
	var service *entities.ServiceModel
	var serviceAdded bool
	if month > 0 && year > 0 {
		uniqueDays := make(map[int]bool)
		for i := range services {
			serviceAdded = false
			if services[i].RdateStart.Year() == year && services[i].RdateStart.Month() == time.Month(month) {
				startDay := services[i].RdateStart.Day()
				if !uniqueDays[startDay] {
					service = mapServiceModel(&services[i])
					service = addServiceAdditionModel(service, &services[i])
					result = append(result, service)
					serviceAdded = true
					uniqueDays[startDay] = true
				}
			}
			if services[i].RdateEnd.Year() == year && services[i].RdateEnd.Month() == time.Month(month) {
				endDay := services[i].RdateEnd.Day()
				if !uniqueDays[endDay] {
					if !serviceAdded {
						service = mapServiceModel(&services[i])
						service = addServiceAdditionModel(service, &services[i])
						result = append(result, service)
					}
					uniqueDays[endDay] = true
				}
			}
		}
	} else {
		for i := range services {
			service = mapServiceModel(&services[i])
			service = addServiceAdditionModel(service, &services[i])
			result = append(result, service)
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

func getSqlService(sqltype, userID string, status string, month, year int) (string, []interface{}, error) {
	whereSQL := ""
	args := []interface{}{}
	idx := 1

	switch sqltype {
	case "owner":
		whereSQL += fmt.Sprintf(`"OID" = $%d::uuid`, idx)
		args = append(args, userID)
		idx++
	case "doctor":
		whereSQL += fmt.Sprintf(`
        EXISTS (
            SELECT 1 FROM "Cservice"
            WHERE "Cservice"."SID" = "Service"."SID"
              AND "Cservice"."CID" = $%d::uuid
        )
		`, idx)
		args = append(args, userID)
		idx++
	case "caretaker":
		whereSQL += fmt.Sprintf(`
        EXISTS (
            SELECT 1 FROM "Mservice"
            WHERE "Mservice"."SID" = "Service"."SID"
              AND "Mservice"."DID" = $%d::uuid
        )
		`, idx)
		args = append(args, userID)
		idx++
	default:
	}

	if status != "" && status != "all" {
		whereSQL += fmt.Sprintf("status = $%d", idx)
		args = append(args, status)
		idx++
	}

	if month > 0 && year > 0 {
		if whereSQL != "" {
			whereSQL += " AND "
		}
		whereSQL += fmt.Sprintf(`
			(
				(rdate_start >= $%d AND rdate_start < $%d) OR
				(rdate_end   >= $%d AND rdate_end   < $%d)
			)
		`, idx, idx+1, idx+2, idx+3)

		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, 0)

		args = append(args, startDate, endDate, startDate, endDate)
		idx += 4
	}

	if whereSQL != "" {
		whereSQL = "WHERE " + whereSQL
	}

	sql := fmt.Sprintf(`
		SELECT CAST(COUNT(*) AS INTEGER) AS count
		FROM "Service"
		%s`, whereSQL)
	return sql, args, nil
}
