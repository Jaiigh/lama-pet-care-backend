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
	FindByEmail(email string) (*entities.LoginUserModel, error)
	FindByID(userID string) (*entities.UserDataModel, error)
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

func (repo *adminRepository) FindByEmail(email string) (*entities.LoginUserModel, error) {
	user, err := repo.Collection.Admin.FindUnique(
		db.Admin.Email.Equals(email),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("users -> FindByEmail: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByEmail: user data is nil")
	}
	return &entities.LoginUserModel{
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
