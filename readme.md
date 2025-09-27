# lama-backend
swagger url: https://lama-pet-care-backend-qbwz.onrender.com/swagger/index.html
# Set Environment `.env`
```
DATABASE_URL="<transaction_pooler>?pgbouncer=true&connect_timeout=10"
PRISMA_DISABLE_PREPARED_STATEMENTS=true
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
You can start the server using either of the following commands (run them from the project root directory):
```
go run .
```
or, if you have Air installed:
```
air
```
Note: air automatically reloads the server whenever you save changes to your code.
# Run this command whenever the Prisma schema changes to regenerate the database client:
```
go run github.com/steebchen/prisma-client-go generate --schema=./domain/prisma/schema.prisma dev
```
# Whenever packages are added or removed from the project
```
go mod tidy
```
# Swagger
swagger url: https://lama-pet-care-backend-qbwz.onrender.com/swagger/index.html
update swagger
```
swag init
```
In the response message, if the output is not just a single string, you have to use response entities instead of fiber.Map; otherwise, Swagger may not fully reflect the real API, and we might lose some points.