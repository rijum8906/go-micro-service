# Root Makefile

.PHONY: infra user-setup user-test clean-all

infra:
	sudo systemctl start postgresql redis
infra-stop:
	sudo systemctl start postgresql redis

# 2. Fully prepare the User Service (Migrations + SQLC)
user-setup:
	$(MAKE) -C services/user-service setup

# 3. Run all tests in the entire monorepo (Common + Services)
test-all:
	go test -v ./...

# 4. Cleanup all generated binaries
clean-all:
	rm -rf bin/
	@echo "Cleanup complete."

stat-user-service:
	$(MAKE) -C services/user-service run
