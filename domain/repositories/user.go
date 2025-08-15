package repositories

import (
	"context"
	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"

	"fmt"
)

type usersRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type IUsersRepository interface {
	InsertUser(data entities.CreatedUserModel) (*entities.UserDataModel, error)
	FindAll() (*[]entities.UserDataModel, error)
	FindByID(userID string) (*entities.UserDataModel, error)
	FindByEmail(email string) (*entities.UserDataModel, error)
	UpdateUser(data entities.UserDataModel) (*entities.UserDataModel, error)
	DeleteUser(userID string) error
}

func NewUsersRepository(db *ds.PrismaDB) IUsersRepository {
	return &usersRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *usersRepository) InsertUser(data entities.CreatedUserModel) (*entities.UserDataModel, error) {
	createdData, err := repo.Collection.User.CreateOne(
		db.User.Email.Set(data.Email),
		db.User.Password.Set(data.Password),
		db.User.Role.Set(data.Role),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> InsertUser: %v", err)
	}

	createdAt, ok := createdData.CreatedAt()
	if !ok {
		return nil, fmt.Errorf("users -> FindAll: createdAt not ok")
	}
	updatedAt, ok := createdData.UpdatedAt()
	if !ok {
		return nil, fmt.Errorf("users -> FindAll: updatedAt not ok")
	}
	return &entities.UserDataModel{
		UserID:    createdData.UserID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Email:     createdData.Email,
		Password:  createdData.Password,
		Role:      createdData.Role,
	}, nil
}

func (repo *usersRepository) FindAll() (*[]entities.UserDataModel, error) {
	users, err := repo.Collection.User.FindMany().Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("users -> FindAll: %v", err)
	}

	var result []entities.UserDataModel
	for _, u := range users {
		createdAt, ok := u.CreatedAt()
		if !ok {
			return nil, fmt.Errorf("users -> FindAll: createdAt not ok")
		}
		updatedAt, ok := u.UpdatedAt()
		if !ok {
			return nil, fmt.Errorf("users -> FindAll: updatedAt not ok")
		}
		result = append(result, entities.UserDataModel{
			UserID:    u.UserID,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Email:     u.Email,
			Password:  u.Password,
			Role:      u.Role,
		})
	}

	return &result, nil
}

func (repo *usersRepository) FindByID(userID string) (*entities.UserDataModel, error) {
	user, err := repo.Collection.User.FindUnique(
		db.User.UserID.Equals(userID),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByID: user data is nil")
	}
	createdAt, ok := user.CreatedAt()
	if !ok {
		return nil, fmt.Errorf("users -> FindByID: createdAt not ok")
	}
	updatedAt, ok := user.UpdatedAt()
	if !ok {
		return nil, fmt.Errorf("users -> FindByID: updatedAt not ok")
	}
	return &entities.UserDataModel{
		UserID:    user.UserID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Email:     user.Email,
		Password:  user.Password,
		Role:      user.Role,
	}, nil
}

func (repo *usersRepository) FindByEmail(email string) (*entities.UserDataModel, error) {
	user, err := repo.Collection.User.FindUnique(
		db.User.Email.Equals(email),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByID: user data is nil")
	}
	return &entities.UserDataModel{
		UserID:   user.UserID,
		Email:    user.Email,
		Password: user.Password,
		Role:     user.Role,
	}, nil
}

func (repo *usersRepository) UpdateUser(data entities.UserDataModel) (*entities.UserDataModel, error) {
	updatedUser, err := repo.Collection.User.
		FindUnique(db.User.UserID.Equals(data.UserID)).
		Update(
			db.User.Email.Set(data.Email),
			db.User.Password.Set(data.Password),
			db.User.Role.Set(data.Role),
		).
		Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("users -> UpdateUser: %v", err)
	}

	createdAt, ok := updatedUser.CreatedAt()
	if !ok {
		return nil, fmt.Errorf("users -> UpdateUser: createdAt not ok")
	}
	updatedAt, ok := updatedUser.UpdatedAt()
	if !ok {
		return nil, fmt.Errorf("users -> UpdateUser: updatedAt not ok")
	}

	return &entities.UserDataModel{
		UserID:    updatedUser.UserID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Email:     updatedUser.Email,
		Password:  updatedUser.Password,
		Role:      updatedUser.Role,
	}, nil
}

func (repo *usersRepository) DeleteUser(userID string) error {
	_, err := repo.Collection.User.
		FindUnique(db.User.UserID.Equals(userID)).
		Delete().
		Exec(repo.Context)
	if err != nil {
		return fmt.Errorf("users -> DeleteUser: %v", err)
	}
	return nil
}
