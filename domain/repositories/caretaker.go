package repositories

import (
	"context"
	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
	"time"

	"fmt"
)

type caretakerRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type ICaretakerRepository interface {
	InsertCaretaker(user_id, specialization string) (*entities.UserDataModel, error)
	FindByID(userID string) (*entities.UserDataModel, error)
	DeleteByID(userID string) (*entities.UserDataModel, error)
	UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error)
	FindAvailableCaretaker(startDate, endDate time.Time) ([]*entities.AvailableStaffResponse, error)
	FindBusyTimeSlot(staffID string, startDate00, startDate23, endDate00, endDate23 time.Time) (*[]db.ServiceModel, error)
}

func NewCaretakerRepository(db *ds.PrismaDB) ICaretakerRepository {
	return &caretakerRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *caretakerRepository) InsertCaretaker(user_id, specialization string) (*entities.UserDataModel, error) {
	createdData, err := repo.Collection.Caretaker.CreateOne(
		db.Caretaker.Users.Link(db.Users.ID.Equals(user_id)),
		db.Caretaker.Specialties.Set(specialization),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> InsertUser: %v", err)
	}

	specialties, _ := createdData.Specialties()

	return &entities.UserDataModel{
		UserID:         createdData.UserID,
		Specialization: specialties,
	}, nil
}

func (repo *caretakerRepository) FindByID(userID string) (*entities.UserDataModel, error) {
	user, err := repo.Collection.Caretaker.FindUnique(
		db.Caretaker.UserID.Equals(userID),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByID: user data is nil")
	}

	specialties, _ := user.Specialties()
	rating, _ := user.Rating()
	return &entities.UserDataModel{
		UserID:         user.UserID,
		Specialization: specialties,
		StartWorkTime:  user.StartWorkingTime,
		EndWorkTime:    user.EndWorkingTime,
		Rating:         rating,
	}, nil
}

func (repo *caretakerRepository) DeleteByID(userID string) (*entities.UserDataModel, error) {
	deletedUser, err := repo.Collection.Caretaker.FindUnique(
		db.Caretaker.UserID.Equals(userID),
	).Delete().Exec(repo.Context) // ลบและคืนค่าที่ถูกลบเลย

	if err != nil {
		return nil, fmt.Errorf("users -> DeleteByID: %v", err)
	}
	if deletedUser == nil {
		return nil, fmt.Errorf("users -> DeleteByID: user not found")
	}

	specialties, _ := deletedUser.Specialties()
	rating, _ := deletedUser.Rating()
	return &entities.UserDataModel{
		UserID:         deletedUser.UserID,
		Specialization: specialties,
		StartWorkTime:  deletedUser.StartWorkingTime,
		EndWorkTime:    deletedUser.EndWorkingTime,
		Rating:         rating,
	}, nil
}

func (repo *caretakerRepository) UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// Start with an empty list of updates
	updates := []db.CaretakerSetParam{}

	if data.Specialization != nil {
		updates = append(updates, db.Caretaker.Specialties.Set(*data.Specialization))
	}
	if data.StartWorkTime != nil {
		updates = append(updates, db.Caretaker.StartWorkingTime.Set(*data.StartWorkTime))
	}
	if data.EndWorkTime != nil {
		updates = append(updates, db.Caretaker.EndWorkingTime.Set(*data.EndWorkTime))
	}
	if data.Rating != nil {
		updates = append(updates, db.Caretaker.Rating.Set(*data.Rating))
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("users -> UpdateByID: no fields to update")
	}

	// Execute update
	updatedUser, err := repo.Collection.Caretaker.FindUnique(
		db.Caretaker.UserID.Equals(userID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> UpdateByID: %v", err)
	}
	if updatedUser == nil {
		return nil, fmt.Errorf("users -> UpdateByID: user not found")
	}

	specialization, _ := updatedUser.Specialties()
	rating, _ := updatedUser.Rating()
	return &entities.UserDataModel{
		UserID:         updatedUser.UserID,
		Specialization: specialization,
		StartWorkTime:  updatedUser.StartWorkingTime,
		EndWorkTime:    updatedUser.EndWorkingTime,
		Rating:         rating,
	}, nil
}

func (repo *caretakerRepository) FindAvailableCaretaker(startDate, endDate time.Time) ([]*entities.AvailableStaffResponse, error) {
	// find Rend < start or Rstart > end
	caretakers, err := repo.Collection.Caretaker.FindMany(
		db.Caretaker.Leaveday.None(
			db.Leaveday.Leaveday.Gte(startDate),
			db.Leaveday.Leaveday.Lte(endDate),
		),
		db.Caretaker.Cservice.None(
			db.Cservice.Service.Where(
				db.Service.Or(
					db.Service.Status.Equals("finish"),
					db.Service.And(
						db.Service.RdateStart.Lte(endDate),
						db.Service.RdateEnd.Gte(startDate),
					),
				),
			),
		),
	).With(
		db.Caretaker.Users.Fetch(),
	).OrderBy(
		db.Caretaker.Rating.Order(db.SortOrderAsc),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}
	if len(caretakers) == 0 { //len(nil) = 0
		return nil, nil
	}

	results := make([]*entities.AvailableStaffResponse, 0, len(caretakers))
	for _, c := range caretakers {
		user := c.Users()
		rating, _ := c.Rating()
		profile, _ := user.ProfileImage()
		entity := entities.AvailableStaffResponse{
			ID:      c.UserID,
			Name:    user.Name,
			Profile: profile,
			Rating:  rating,
		}

		results = append(results, &entity)
	}

	return results, nil
}

func (repo *caretakerRepository) FindBusyTimeSlot(staffID string, startDate00, startDate23, endDate00, endDate23 time.Time) (*[]db.ServiceModel, error) {
	services, err := repo.Collection.Service.FindMany(
		db.Service.Cservice.Where(
			db.Cservice.Cid.Equals(staffID),
		),
		db.Service.Or(
			db.Service.And(
				db.Service.RdateStart.Gt(endDate00),
				db.Service.RdateStart.Lt(endDate23),
			),
			db.Service.And(
				db.Service.RdateEnd.Gt(startDate00),
				db.Service.RdateEnd.Lt(startDate23),
			),
		),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}

	return &services, nil
}
