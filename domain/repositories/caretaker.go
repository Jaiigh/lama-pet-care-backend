package repositories

import (
	"context"
	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"

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
	FindAvailableCaretaker(dates entities.RDateRange) (*[]db.CaretakerModel, error)
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

func (repo *caretakerRepository) FindAvailableCaretaker(dates entities.RDateRange) (*[]db.CaretakerModel, error) {
	caretakers, err := repo.Collection.Caretaker.FindMany(
		db.Caretaker.Leaveday.None(
			db.Leaveday.Leaveday.Gte(dates.StartDate),
			db.Leaveday.Leaveday.Lte(dates.EndDate),
		),
		db.Caretaker.Cservice.None(
			db.Cservice.Service.Where(
				db.Service.RdateStart.Lte(dates.EndDate),
				db.Service.RdateEnd.Gte(dates.StartDate),
			),
		),
	).With(
		db.Caretaker.Cservice.Fetch().With(
			db.Cservice.Service.Fetch(),
		),
		db.Caretaker.Users.Fetch(),
	).OrderBy(
		db.Caretaker.Rating.Order(db.SortOrderAsc),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}
	if caretakers == nil {
		return nil, nil
	}

	return &caretakers, nil
}
