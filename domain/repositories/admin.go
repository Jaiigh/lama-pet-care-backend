package repositories

import (
	"context"
	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"

	"fmt"
)

type adminRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type IAdminRepository interface {
	InsertAdmin(data entities.CreatedUserModel) (*entities.UserDataModel, error)
	FindByEmail(email string) (*entities.LoginUserResponseModel, error)
	FindByID(userID string) (*entities.UserDataModel, error)
	DeleteByID(userID string) (*entities.UserDataModel, error)
	UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error)
}

func NewAdminRepository(db *ds.PrismaDB) IAdminRepository {
	return &adminRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *adminRepository) InsertAdmin(data entities.CreatedUserModel) (*entities.UserDataModel, error) {
	createdData, err := repo.Collection.Admin.CreateOne(
		db.Admin.Email.Set(data.Email),
		db.Admin.Password.Set(data.Password),
		db.Admin.Name.Set(data.Name),
		db.Admin.Birthdate.Set(data.BirthDate),
		db.Admin.TelephoneNumber.Set(data.TelephoneNumber),
		db.Admin.Address.Set(data.Address),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> InsertUser: %v", err)
	}

	return &entities.UserDataModel{
		UserID:          createdData.Aid,
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

func (repo *adminRepository) FindByEmail(email string) (*entities.LoginUserResponseModel, error) {
	user, err := repo.Collection.Admin.FindUnique(
		db.Admin.Email.Equals(email),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("users -> FindByEmail: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByEmail: user data is nil")
	}
	return &entities.LoginUserResponseModel{
		UserID:   user.Aid,
		Email:    user.Email,
		Password: user.Password,
	}, nil
}

func (repo *adminRepository) FindByID(userID string) (*entities.UserDataModel, error) {
	user, err := repo.Collection.Admin.FindUnique(
		db.Admin.Aid.Equals(userID),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByID: user data is nil")
	}

	return &entities.UserDataModel{
		UserID:          user.Aid,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		Email:           user.Email,
		Password:        user.Password,
		Name:            user.Name,
		BirthDate:       user.Birthdate,
		TelephoneNumber: user.TelephoneNumber,
		Address:         user.Address,
	}, nil
}

func (repo *adminRepository) DeleteByID(userID string) (*entities.UserDataModel, error) {
	deletedUser, err := repo.Collection.Admin.FindUnique(
		db.Admin.Aid.Equals(userID),
	).Delete().Exec(repo.Context) // ลบและคืนค่าที่ถูกลบเลย

	if err != nil {
		return nil, fmt.Errorf("users -> DeleteByID: %v", err)
	}
	if deletedUser == nil {
		return nil, fmt.Errorf("users -> DeleteByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:          deletedUser.Aid,
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

func (repo *adminRepository) UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// Start with an empty list of updates
	updates := []db.AdminSetParam{}

	if data.Email != nil {
		updates = append(updates, db.Admin.Email.Set(*data.Email))
	}
	if data.Password != nil {
		updates = append(updates, db.Admin.Password.Set(*data.Password))
	}
	if data.Name != nil {
		updates = append(updates, db.Admin.Name.Set(*data.Name))
	}
	if data.BirthDate != nil {
		updates = append(updates, db.Admin.Birthdate.Set(*data.BirthDate))
	}
	if data.TelephoneNumber != nil {
		updates = append(updates, db.Admin.TelephoneNumber.Set(*data.TelephoneNumber))
	}
	if data.Address != nil {
		updates = append(updates, db.Admin.Address.Set(*data.Address))
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("users -> UpdateByID: no fields to update")
	}

	// Execute update
	updatedUser, err := repo.Collection.Admin.FindUnique(
		db.Admin.Aid.Equals(userID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> UpdateByID: %v", err)
	}
	if updatedUser == nil {
		return nil, fmt.Errorf("users -> UpdateByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:          updatedUser.Aid,
		CreatedAt:       updatedUser.CreatedAt,
		UpdatedAt:       updatedUser.UpdatedAt,
		Email:           updatedUser.Email,
		Password:        updatedUser.Password,
		Name:            updatedUser.Name,
		BirthDate:       updatedUser.Birthdate,
		TelephoneNumber: updatedUser.TelephoneNumber,
		Address:         updatedUser.Address,
	}, nil
}
