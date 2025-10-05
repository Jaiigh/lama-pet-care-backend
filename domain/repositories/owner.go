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
	InsertOwner(user_id string) (*entities.UserDataModel, error)
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

func (repo *ownerRepository) InsertOwner(user_id string) (*entities.UserDataModel, error) {
	createdData, err := repo.Collection.Owner.CreateOne(
		db.Owner.Users.Link(db.Users.ID.Equals(user_id)),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> InsertUser: %v", err)
	}

	return &entities.UserDataModel{
		UserID:        createdData.UserID,
		TotalSpending: createdData.TotalSpending,
	}, nil
}

func (repo *ownerRepository) FindByID(userID string) (*entities.UserDataModel, error) {
	user, err := repo.Collection.Owner.FindUnique(
		db.Owner.UserID.Equals(userID),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByID: user data is nil")
	}

	return &entities.UserDataModel{
		UserID:        user.UserID,
		TotalSpending: user.TotalSpending,
	}, nil
}

func (repo *ownerRepository) DeleteByID(userID string) (*entities.UserDataModel, error) {
	deletedUser, err := repo.Collection.Owner.FindUnique(
		db.Owner.UserID.Equals(userID),
	).Delete().Exec(repo.Context) // ลบและคืนค่าที่ถูกลบเลย

	if err != nil {
		return nil, fmt.Errorf("users -> DeleteByID: %v", err)
	}
	if deletedUser == nil {
		return nil, fmt.Errorf("users -> DeleteByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:        deletedUser.UserID,
		TotalSpending: deletedUser.TotalSpending,
	}, nil
}

func (repo *ownerRepository) UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	updates := []db.OwnerSetParam{}

	if data.TotalSpending != nil {
		updates = append(updates, db.Owner.TotalSpending.Set(*data.TotalSpending))
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("users -> UpdateByID: no fields to update")
	}

	updatedUser, err := repo.Collection.Owner.FindUnique(
		db.Owner.UserID.Equals(userID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> UpdateByID: %v", err)
	}
	if updatedUser == nil {
		return nil, fmt.Errorf("users -> UpdateByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:        updatedUser.UserID,
		TotalSpending: updatedUser.TotalSpending,
	}, nil
}
