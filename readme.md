# lama-backend
# Set Environment `.env`
```
DATABASE_URL= <database in prismadb>
PORT=8080

JWT_SECRET_KEY=Test
JWT_REFESH_SECRET_KEY=Test
```
# Install dependencies
```
go get github.com/steebchen/prisma-client-go
go run github.com/steebchen/prisma-client-go generate --schema=./domain/prisma/schema.prisma dev
go mod tidy
```
# Run development server
```
go run .
```
or
```
air
```
# Run this command whenever the Prisma schema changes to regenerate the database client:
```
go run github.com/steebchen/prisma-client-go generate --schema=./domain/prisma/schema.prisma dev
```
# Whenever packages are added or removed from the project
```
go mod tidy
```