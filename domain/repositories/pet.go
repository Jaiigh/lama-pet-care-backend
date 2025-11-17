package repositories

import (
	"context"
	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"

	"fmt"
)

type petRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type IPetRepository interface {
	InsertPet(data entities.CreatedPetModel) (*entities.PetDataModel, error)
	FindByOwnerID(ownerID string) ([]entities.PetDataModel, error)
	FindPetByID(petID string) (*entities.PetDataModel, error)
	FindAll() ([]entities.PetDataModel, error)
	UpdatePet(petID string, data entities.UpdatePetModel) (*entities.PetDataModel, error)
	DeletePet(petID string) (*entities.PetDataModel, error)
}

func NewPetRepository(db *ds.PrismaDB) IPetRepository {
	return &petRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *petRepository) InsertPet(data entities.CreatedPetModel) (*entities.PetDataModel, error) {
	createdData, err := repo.Collection.Pet.CreateOne(
		db.Pet.Birthdate.Set(data.BirthDate),
		db.Pet.Weight.Set(data.Weight),
		db.Pet.Kind.Set(data.Kind),
		db.Pet.Sex.Set(data.Sex),
		db.Pet.Owner.Link(db.Owner.UserID.Equals(data.OwnerID)),

		// optional fields
		db.Pet.Breed.SetIfPresent(data.Breed),
		db.Pet.Name.SetIfPresent(data.Name),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("pets -> InsertPet: %v", err)
	}

	breed, _ := createdData.Breed()
	name, _ := createdData.Name()

	return &entities.PetDataModel{
		PetID:     createdData.Petid,
		OwnerID:   createdData.Oid,
		Breed:     breed,
		Name:      name,
		BirthDate: createdData.Birthdate,
		Weight:    createdData.Weight,
		Kind:      createdData.Kind,
		Sex:       createdData.Sex,
	}, nil
}

func (repo *petRepository) FindByOwnerID(ownerID string) ([]entities.PetDataModel, error) {
	pets, err := repo.Collection.Pet.FindMany(
		db.Pet.Oid.Equals(ownerID),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("pets -> FindByOwnerID: %v", err)
	}

	var results []entities.PetDataModel
	for i := range pets {
		breed, _ := pets[i].Breed()
		name, _ := pets[i].Name()

		results = append(results, entities.PetDataModel{
			PetID:     pets[i].Petid,
			OwnerID:   pets[i].Oid,
			Breed:     breed,
			Name:      name,
			BirthDate: pets[i].Birthdate,
			Weight:    pets[i].Weight,
			Kind:      pets[i].Kind,
			Sex:       pets[i].Sex,
		})
	}

	return results, nil
}

func (repo *petRepository) FindPetByID(petID string) (*entities.PetDataModel, error) {
	pet, err := repo.Collection.Pet.FindUnique(
		db.Pet.Petid.Equals(petID),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("pets -> FindPetByID: %w", err)
	}
	if pet == nil {
		return nil, fmt.Errorf("pets -> FindPetByID: pet not found")
	}

	breed, _ := pet.Breed()
	name, _ := pet.Name()

	return &entities.PetDataModel{
		PetID:     pet.Petid,
		OwnerID:   pet.Oid,
		Breed:     breed,
		Name:      name,
		BirthDate: pet.Birthdate,
		Weight:    pet.Weight,
		Kind:      pet.Kind,
		Sex:       pet.Sex,
	}, nil
}

func (repo *petRepository) FindAll() ([]entities.PetDataModel, error) {
	pets, err := repo.Collection.Pet.FindMany().Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("pets -> FindAll: %v", err)
	}

	var results []entities.PetDataModel
	for i := range pets {
		breed, _ := pets[i].Breed()
		name, _ := pets[i].Name()

		results = append(results, entities.PetDataModel{
			PetID:     pets[i].Petid,
			OwnerID:   pets[i].Oid,
			Breed:     breed,
			Name:      name,
			BirthDate: pets[i].Birthdate,
			Weight:    pets[i].Weight,
			Kind:      pets[i].Kind,
			Sex:       pets[i].Sex,
		})
	}

	return results, nil
}

func (repo *petRepository) UpdatePet(petID string, data entities.UpdatePetModel) (*entities.PetDataModel, error) {
	updates := []db.PetSetParam{}

	if data.Breed != nil {
		updates = append(updates, db.Pet.Breed.SetIfPresent(data.Breed))
	}
	if data.Name != nil {
		updates = append(updates, db.Pet.Name.SetIfPresent(data.Name))
	}
	if data.BirthDate != nil {
		updates = append(updates, db.Pet.Birthdate.Set(*data.BirthDate))
	}
	if data.Weight != nil {
		updates = append(updates, db.Pet.Weight.Set(*data.Weight))
	}
	if data.Kind != nil {
		updates = append(updates, db.Pet.Kind.Set(*data.Kind))
	}
	if data.Sex != nil {
		updates = append(updates, db.Pet.Sex.Set(*data.Sex))
	}
	if data.OwnerID != nil {
		updates = append(updates, db.Pet.Owner.Link(db.Owner.UserID.Equals(*data.OwnerID)))
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("pets -> UpdatePet: no fields to update")
	}

	updated, err := repo.Collection.Pet.FindUnique(
		db.Pet.Petid.Equals(petID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("pets -> UpdatePet: %v", err)
	}
	if updated == nil {
		return nil, fmt.Errorf("pets -> UpdatePet: pet not found")
	}

	breed, _ := updated.Breed()
	name, _ := updated.Name()

	return &entities.PetDataModel{
		PetID:     updated.Petid,
		OwnerID:   updated.Oid,
		Breed:     breed,
		Name:      name,
		BirthDate: updated.Birthdate,
		Weight:    updated.Weight,
		Kind:      updated.Kind,
		Sex:       updated.Sex,
	}, nil
}

func (repo *petRepository) DeletePet(petID string) (*entities.PetDataModel, error) {
	deleted, err := repo.Collection.Pet.FindUnique(
		db.Pet.Petid.Equals(petID),
	).Delete().Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("pets -> DeletePet: %v", err)
	}
	if deleted == nil {
		return nil, fmt.Errorf("pets -> DeletePet: pet not found")
	}

	breed, _ := deleted.Breed()
	name, _ := deleted.Name()

	return &entities.PetDataModel{
		PetID:     deleted.Petid,
		OwnerID:   deleted.Oid,
		Breed:     breed,
		Name:      name,
		BirthDate: deleted.Birthdate,
		Weight:    deleted.Weight,
		Kind:      deleted.Kind,
		Sex:       deleted.Sex,
	}, nil
}
