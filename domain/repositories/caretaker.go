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
	InsertCaretaker(data entities.CreatedUserModel) (*entities.UserDataModel, error)
	FindByEmail(email string) (*entities.LoginUserResponseModel, error)
	FindByID(userID string) (*entities.UserDataModel, error)
	DeleteByID(userID string) (*entities.UserDataModel, error)
	UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error)
}

func NewCaretakerRepository(db *ds.PrismaDB) ICaretakerRepository {
	return &caretakerRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *caretakerRepository) InsertCaretaker(data entities.CreatedUserModel) (*entities.UserDataModel, error) {
	createdData, err := repo.Collection.Caretaker.CreateOne(
		db.Caretaker.Email.Set(data.Email),
		db.Caretaker.Password.Set(data.Password),
		db.Caretaker.Name.Set(data.Name),
		db.Caretaker.Birthdate.Set(data.BirthDate),
		db.Caretaker.TelephoneNumber.Set(data.TelephoneNumber),
		db.Caretaker.Address.Set(data.Address),
		db.Caretaker.Specialties.Set(data.Specialization),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> InsertUser: %v", err)
	}

	return &entities.UserDataModel{
		UserID:          createdData.Cid,
		CreatedAt:       createdData.CreatedAt,
		UpdatedAt:       createdData.UpdatedAt,
		Email:           createdData.Email,
		Password:        createdData.Password,
		Name:            createdData.Name,
		BirthDate:       createdData.Birthdate,
		TelephoneNumber: createdData.TelephoneNumber,
		Address:         createdData.Address,
	}, nil
}

func (repo *caretakerRepository) FindByEmail(email string) (*entities.LoginUserResponseModel, error) {
	user, err := repo.Collection.Caretaker.FindUnique(
		db.Caretaker.Email.Equals(email),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("users -> FindByEmail: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByEmail: user data is nil")
	}
	return &entities.LoginUserResponseModel{
		UserID:   user.Cid,
		Email:    user.Email,
		Password: user.Password,
	}, nil
}

func (repo *caretakerRepository) FindByID(userID string) (*entities.UserDataModel, error) {
	user, err := repo.Collection.Caretaker.FindUnique(
		db.Caretaker.Cid.Equals(userID),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByID: user data is nil")
	}

	specialties, ok := user.Specialties()
	if !ok {
		return nil, fmt.Errorf("users -> FindByID: specialties not ok")
	}

	rating, ok := user.Rating()
	if !ok {
		return nil, fmt.Errorf("users -> FindByID: Rating not ok")
	}

	return &entities.UserDataModel{
		UserID:          user.Cid,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		Email:           user.Email,
		Password:        user.Password,
		Name:            user.Name,
		BirthDate:       user.Birthdate,
		TelephoneNumber: user.TelephoneNumber,
		Address:         user.Address,
		Specialization:  specialties,
		StartWorkTime:   user.StartWorkingTime,
		EndWorkTime:     user.EndWorkingTime,
		Rating:          rating,
	}, nil
}

func (repo *caretakerRepository) DeleteByID(userID string) (*entities.UserDataModel, error) {
	deletedUser, err := repo.Collection.Caretaker.FindUnique(
		db.Caretaker.Cid.Equals(userID),
	).Delete().Exec(repo.Context) // ลบและคืนค่าที่ถูกลบเลย

	if err != nil {
		return nil, fmt.Errorf("users -> DeleteByID: %v", err)
	}
	if deletedUser == nil {
		return nil, fmt.Errorf("users -> DeleteByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:          deletedUser.Cid,
		CreatedAt:       deletedUser.CreatedAt,
		UpdatedAt:       deletedUser.UpdatedAt,
		Email:           deletedUser.Email,
		Password:        deletedUser.Password,
		Name:            deletedUser.Name,
		BirthDate:       deletedUser.Birthdate,
		TelephoneNumber: deletedUser.TelephoneNumber,
		Address:         deletedUser.Address,
	}, nil
}

func (repo *caretakerRepository) UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// Start with an empty list of updates
	updates := []db.CaretakerSetParam{}

	if data.Email != nil {
		updates = append(updates, db.Caretaker.Email.Set(*data.Email))
	}
	if data.Password != nil {
		updates = append(updates, db.Caretaker.Password.Set(*data.Password))
	}
	if data.Name != nil {
		updates = append(updates, db.Caretaker.Name.Set(*data.Name))
	}
	if data.BirthDate != nil {
		updates = append(updates, db.Caretaker.Birthdate.Set(*data.BirthDate))
	}
	if data.TelephoneNumber != nil {
		updates = append(updates, db.Caretaker.TelephoneNumber.Set(*data.TelephoneNumber))
	}
	if data.Address != nil {
		updates = append(updates, db.Caretaker.Address.Set(*data.Address))
	}
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
		db.Caretaker.Cid.Equals(userID),
	).Update(updates...).Exec(repo.Context)

	specialization, ok := updatedUser.Specialties()
	if !ok {
		specialization = ""
	}
	rating, ok := updatedUser.Rating()
	if !ok {
		rating = db.Decimal{}
	}

	if err != nil {
		return nil, fmt.Errorf("users -> UpdateByID: %v", err)
	}
	if updatedUser == nil {
		return nil, fmt.Errorf("users -> UpdateByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:          updatedUser.Cid,
		CreatedAt:       updatedUser.CreatedAt,
		UpdatedAt:       updatedUser.UpdatedAt,
		Email:           updatedUser.Email,
		Password:        updatedUser.Password,
		Name:            updatedUser.Name,
		BirthDate:       updatedUser.Birthdate,
		TelephoneNumber: updatedUser.TelephoneNumber,
		Address:         updatedUser.Address,
		Specialization:  specialization,
		StartWorkTime:   updatedUser.StartWorkingTime,
		EndWorkTime:     updatedUser.EndWorkingTime,
		Rating:          rating,
	}, nil
}
