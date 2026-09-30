.PHONY: dev frontend backend

dev:
	(cd frontend/my_app && npm run dev) & \
	(cd backend && go run cmd/server/main.go) & \
	wait

frontend:
	cd frontend/my_app && npm run dev

backend:
	cd backend && go run cmd/server/main.go
