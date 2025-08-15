package datasources

import (
	"context"
	"lama-backend/domain/prisma/db"
	"log"
)

type PrismaDB struct {
	Context  context.Context
	PrismaDB *db.PrismaClient
}

func ConnectPrisma() *PrismaDB {
	// Connect to the Prisma database
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		log.Fatal("error connection : ", err)
	}
	log.Println("Connected to Prisma database successfully")
	return &PrismaDB{
		Context:  context.Background(),
		PrismaDB: client,
	}
}
