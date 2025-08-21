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
	FindByEmail(email string) (*entities.LoginUserModel, error)
	FindByID(userID string) (*entities.UserDataModel, error)
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
		db.Caretaker.StartWorkingTime.Set(data.StartWorkTime),
		db.Caretaker.EndWorkingTime.Set(data.EndWorkTime),
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
		StartWorkTime:   createdData.StartWorkingTime,
		EndWorkTime:     createdData.EndWorkingTime,
	}, nil
}

func (repo *caretakerRepository) FindByEmail(email string) (*entities.LoginUserModel, error) {
	user, err := repo.Collection.Caretaker.FindUnique(
		db.Caretaker.Email.Equals(email),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("users -> FindByEmail: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByEmail: user data is nil")
	}
	return &entities.LoginUserModel{
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
