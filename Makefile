# Nom de l'exécutable
OUTFILE = DictPolisher

all: build

# Règle pour compiler le programme
build:
	@go build -o $(OUTFILE)

# Règle pour exécuter le programme
run: build
	@./$(OUTFILE)

.PHONY: all build run
