package ldb

type App struct {
	Migrations      map[string]*Migration
	DatabaseService *DatabaseService
	HttpService     *HttpService
}

type Migration struct {
	Up   func() error
	Down func() error
}

type DatabaseService interface {
	CreateCollection(schema CollectionSchema) error
	DropCollection(name string) error
}

type HttpService interface {
}

func (app *App) RegisterMigration(name string, migration Migration) {
	if app.Migrations == nil {
		app.Migrations = map[string]*Migration{}
	}

	app.Migrations[name] = &migration
}

func (app *App) Start() {

}
