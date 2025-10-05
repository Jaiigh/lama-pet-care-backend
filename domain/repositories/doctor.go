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
	InsertDoctor(user_id, license_number string) (*entities.UserDataModel, error)
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

func (repo *doctorRepository) InsertDoctor(user_id, license_number string) (*entities.UserDataModel, error) {
	createdData, err := repo.Collection.Doctor.CreateOne(
		db.Doctor.LicenseNumber.Set(license_number),
		db.Doctor.Users.Link(db.Users.ID.Equals(user_id)),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> InsertUser: %v", err)
	}

	return &entities.UserDataModel{
		UserID:        createdData.UserID,
		LicenseNumber: &createdData.LicenseNumber,
	}, nil
}

func (repo *doctorRepository) FindByID(userID string) (*entities.UserDataModel, error) {
	user, err := repo.Collection.Doctor.FindUnique(
		db.Doctor.UserID.Equals(userID),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByID: user data is nil")
	}

	return &entities.UserDataModel{
		UserID:        user.UserID,
		LicenseNumber: &user.LicenseNumber,
		StartDate:     &user.StartDate,
		StartWorkTime: &user.StartWorkingTime,
		EndWorkTime:   &user.EndWorkingTime,
	}, nil
}

func (repo *doctorRepository) DeleteByID(userID string) (*entities.UserDataModel, error) {
	deletedUser, err := repo.Collection.Doctor.FindUnique(
		db.Doctor.UserID.Equals(userID),
	).Delete().Exec(repo.Context) // ลบและคืนค่าที่ถูกลบเลย

	if err != nil {
		return nil, fmt.Errorf("users -> DeleteByID: %v", err)
	}
	if deletedUser == nil {
		return nil, fmt.Errorf("users -> DeleteByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:        deletedUser.UserID,
		LicenseNumber: &deletedUser.LicenseNumber,
		StartDate:     &deletedUser.StartDate,
		StartWorkTime: &deletedUser.StartWorkingTime,
		EndWorkTime:   &deletedUser.EndWorkingTime,
	}, nil
}

func (repo *doctorRepository) UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// Start with an empty list of updates
	updates := []db.DoctorSetParam{}

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
		db.Doctor.UserID.Equals(userID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> UpdateByID: %v", err)
	}
	if updatedUser == nil {
		return nil, fmt.Errorf("users -> UpdateByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:        updatedUser.UserID,
		LicenseNumber: &updatedUser.LicenseNumber,
		StartDate:     &updatedUser.StartDate,
		StartWorkTime: &updatedUser.StartWorkingTime,
		EndWorkTime:   &updatedUser.EndWorkingTime,
	}, nil
}
