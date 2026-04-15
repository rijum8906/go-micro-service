package migrations

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed migrations/*.sql
var files embed.FS

type Migration struct {
	Name    string
	Content string
}

func All() ([]Migration, error) {
	matches, err := fs.Glob(files, "migrations/*.sql")
	if err != nil {
		return nil, fmt.Errorf("failed to list migration files: %w", err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no migration files found")
	}

	sort.Strings(matches)

	migrations := make([]Migration, 0, len(matches))
	for _, name := range matches {
		content, err := files.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", name, err)
		}

		migrations = append(migrations, Migration{
			Name:    name,
			Content: string(content),
		})
	}

	return migrations, nil
}
