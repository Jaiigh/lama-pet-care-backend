package repositories

import (
	"context"
	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"

	"fmt"
)

type ownerRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type IOwnerRepository interface {
	InsertOwner(data entities.CreatedUserModel) (*entities.UserDataModel, error)
	FindByEmail(email string) (*entities.LoginUserModel, error)
	FindByID(userID string) (*entities.UserDataModel, error)
	DeleteByID(userID string) (*entities.UserDataModel, error)
	UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error)
}

func NewOwnerRepository(db *ds.PrismaDB) IOwnerRepository {
	return &ownerRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *ownerRepository) InsertOwner(data entities.CreatedUserModel) (*entities.UserDataModel, error) {
	createdData, err := repo.Collection.Owner.CreateOne(
		db.Owner.Email.Set(data.Email),
		db.Owner.Password.Set(data.Password),
		db.Owner.Name.Set(data.Name),
		db.Owner.Birthdate.Set(data.BirthDate),
		db.Owner.TelephoneNumber.Set(data.TelephoneNumber),
		db.Owner.Address.Set(data.Address),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> InsertUser: %v", err)
	}

	return &entities.UserDataModel{
		UserID:          createdData.Oid,
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

func (repo *ownerRepository) FindByEmail(email string) (*entities.LoginUserModel, error) {
	user, err := repo.Collection.Owner.FindUnique(
		db.Owner.Email.Equals(email),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("users -> FindByEmail: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByEmail: user data is nil")
	}
	return &entities.LoginUserModel{
		UserID:   user.Oid,
		Email:    user.Email,
		Password: user.Password,
	}, nil
}

func (repo *ownerRepository) FindByID(userID string) (*entities.UserDataModel, error) {
	user, err := repo.Collection.Owner.FindUnique(
		db.Owner.Oid.Equals(userID),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByID: user data is nil")
	}

	return &entities.UserDataModel{
		UserID:          user.Oid,
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

func (repo *ownerRepository) DeleteByID(userID string) (*entities.UserDataModel, error) {
	deletedUser, err := repo.Collection.Owner.FindUnique(
		db.Owner.Oid.Equals(userID),
	).Delete().Exec(repo.Context) // ลบและคืนค่าที่ถูกลบเลย

	if err != nil {
		return nil, fmt.Errorf("users -> DeleteByID: %v", err)
	}
	if deletedUser == nil {
		return nil, fmt.Errorf("users -> DeleteByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:          deletedUser.Oid,
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

func (repo *ownerRepository) UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// Start with an empty list of updates
	updates := []db.OwnerSetParam{}

	if data.Email != nil {
		updates = append(updates, db.Owner.Email.Set(*data.Email))
	}
	if data.Password != nil {
		updates = append(updates, db.Owner.Password.Set(*data.Password))
	}
	if data.Name != nil {
		updates = append(updates, db.Owner.Name.Set(*data.Name))
	}
	if data.BirthDate != nil {
		updates = append(updates, db.Owner.Birthdate.Set(*data.BirthDate))
	}
	if data.TelephoneNumber != nil {
		updates = append(updates, db.Owner.TelephoneNumber.Set(*data.TelephoneNumber))
	}
	if data.Address != nil {
		updates = append(updates, db.Owner.Address.Set(*data.Address))
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("users -> UpdateByID: no fields to update")
	}

	// Execute update
	updatedUser, err := repo.Collection.Owner.FindUnique(
		db.Owner.Oid.Equals(userID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> UpdateByID: %v", err)
	}
	if updatedUser == nil {
		return nil, fmt.Errorf("users -> UpdateByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:          updatedUser.Oid,
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
