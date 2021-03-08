# create new migration
migration-new:
	@read -p "enter migration name: " name; \
	dbmate new $$name;

# run pending migrations
migration-up:
	dbmate up

# migration down 1 step
migration-down:
	dbmate down

# migration status
migration-status:
	dbmate status
	
# drop database
db-drop:
	dbmate drop

# sqlboiler
boil:
	sqlboiler psql

# start redis
redis:
	redis-server &

# build
build:
	GOOS=linux GOARCH=amd64 go build -o build/go-rbac

# run development server
server:
	go run main.go