package repositories

import (
	"context"
	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"

	"fmt"
)

type doctorRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type IDoctorRepository interface {
	InsertDoctor(data entities.CreatedUserModel) (*entities.UserDataModel, error)
	FindByEmail(email string) (*entities.LoginUserModel, error)
	FindByID(userID string) (*entities.UserDataModel, error)
}

func NewDoctorRepository(db *ds.PrismaDB) IDoctorRepository {
	return &doctorRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *doctorRepository) InsertDoctor(data entities.CreatedUserModel) (*entities.UserDataModel, error) {
	createdData, err := repo.Collection.Doctor.CreateOne(
		db.Doctor.Email.Set(data.Email),
		db.Doctor.Password.Set(data.Password),
		db.Doctor.Name.Set(data.Name),
		db.Doctor.Birthdate.Set(data.BirthDate),
		db.Doctor.TelephoneNumber.Set(data.TelephoneNumber),
		db.Doctor.Address.Set(data.Address),
		db.Doctor.LicenseNumber.Set(data.LicenseNumber),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> InsertUser: %v", err)
	}

	return &entities.UserDataModel{
		UserID:          createdData.Did,
		CreatedAt:       createdData.CreatedAt,
		UpdatedAt:       createdData.UpdatedAt,
		Email:           createdData.Email,
		Password:        createdData.Password,
		Name:            createdData.Name,
		BirthDate:       createdData.Birthdate,
		TelephoneNumber: createdData.TelephoneNumber,
		Address:         createdData.Address,
		LicenseNumber:   createdData.LicenseNumber,
	}, nil
}

func (repo *doctorRepository) FindByEmail(email string) (*entities.LoginUserModel, error) {
	user, err := repo.Collection.Doctor.FindUnique(
		db.Doctor.Email.Equals(email),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("users -> FindByEmail: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByEmail: user data is nil")
	}
	return &entities.LoginUserModel{
		UserID:   user.Did,
		Email:    user.Email,
		Password: user.Password,
	}, nil
}

func (repo *doctorRepository) FindByID(userID string) (*entities.UserDataModel, error) {
	user, err := repo.Collection.Doctor.FindUnique(
		db.Doctor.Did.Equals(userID),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByID: user data is nil")
	}

	return &entities.UserDataModel{
		UserID:          user.Did,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		Email:           user.Email,
		Password:        user.Password,
		Name:            user.Name,
		BirthDate:       user.Birthdate,
		TelephoneNumber: user.TelephoneNumber,
		Address:         user.Address,
		LicenseNumber:   user.LicenseNumber,
		StartDate:       user.StartDate,
		StartWorkTime:   user.StartWorkingTime,
		EndWorkTime:     user.EndWorkingTime,
	}, nil
}
