<div align="center">

# lama-backend
# Set Environment `.env`
```
DATABASE_URL= <database in prismadb>
PORT=8080

JWT_SECRET_KEY=Test
JWT_REFESH_SECRET_KEY=Test
```

# Generate or Update Prisma
go run github.com/steebchen/prisma-client-go generate --schema=./domain/prisma/schema.prisma dev