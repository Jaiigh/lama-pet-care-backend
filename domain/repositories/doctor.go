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
	FindByEmail(email string) (*entities.LoginUserResponseModel, error)
	FindByID(userID string) (*entities.UserDataModel, error)
	DeleteByID(userID string) (*entities.UserDataModel, error)
	UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error)
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

func (repo *doctorRepository) FindByEmail(email string) (*entities.LoginUserResponseModel, error) {
	user, err := repo.Collection.Doctor.FindUnique(
		db.Doctor.Email.Equals(email),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("users -> FindByEmail: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByEmail: user data is nil")
	}
	return &entities.LoginUserResponseModel{
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

func (repo *doctorRepository) DeleteByID(userID string) (*entities.UserDataModel, error) {
	deletedUser, err := repo.Collection.Doctor.FindUnique(
		db.Doctor.Did.Equals(userID),
	).Delete().Exec(repo.Context) // ลบและคืนค่าที่ถูกลบเลย

	if err != nil {
		return nil, fmt.Errorf("users -> DeleteByID: %v", err)
	}
	if deletedUser == nil {
		return nil, fmt.Errorf("users -> DeleteByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:          deletedUser.Did,
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

func (repo *doctorRepository) UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// Start with an empty list of updates
	updates := []db.DoctorSetParam{}

	if data.Email != nil {
		updates = append(updates, db.Doctor.Email.Set(*data.Email))
	}
	if data.Password != nil {
		updates = append(updates, db.Doctor.Password.Set(*data.Password))
	}
	if data.Name != nil {
		updates = append(updates, db.Doctor.Name.Set(*data.Name))
	}
	if data.BirthDate != nil {
		updates = append(updates, db.Doctor.Birthdate.Set(*data.BirthDate))
	}
	if data.TelephoneNumber != nil {
		updates = append(updates, db.Doctor.TelephoneNumber.Set(*data.TelephoneNumber))
	}
	if data.Address != nil {
		updates = append(updates, db.Doctor.Address.Set(*data.Address))
	}
	if data.LicenseNumber != nil {
		updates = append(updates, db.Doctor.LicenseNumber.Set(*data.LicenseNumber))
	}
	if data.StartDate != nil {
		updates = append(updates, db.Doctor.StartDate.Set(db.DateTime(*data.StartDate)))
	}
	if data.StartWorkTime != nil {
		updates = append(updates, db.Doctor.StartWorkingTime.Set(*data.StartWorkTime))
	}
	if data.EndWorkTime != nil {
		updates = append(updates, db.Doctor.EndWorkingTime.Set(*data.EndWorkTime))
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("users -> UpdateByID: no fields to update")
	}

	// Execute update
	updatedUser, err := repo.Collection.Doctor.FindUnique(
		db.Doctor.Did.Equals(userID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> UpdateByID: %v", err)
	}
	if updatedUser == nil {
		return nil, fmt.Errorf("users -> UpdateByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:          updatedUser.Did,
		CreatedAt:       updatedUser.CreatedAt,
		UpdatedAt:       updatedUser.UpdatedAt,
		Email:           updatedUser.Email,
		Password:        updatedUser.Password,
		Name:            updatedUser.Name,
		BirthDate:       updatedUser.Birthdate,
		TelephoneNumber: updatedUser.TelephoneNumber,
		Address:         updatedUser.Address,
		LicenseNumber:   updatedUser.LicenseNumber,
		StartDate:       updatedUser.StartDate,
		StartWorkTime:   updatedUser.StartWorkingTime,
		EndWorkTime:     updatedUser.EndWorkingTime,
	}, nil
}
