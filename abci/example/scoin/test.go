package scoin

func (app *Application) TestSet(key, value []byte) {
	err := app.state.db.Set(prefixKey(key), value)
	if err != nil {
		panic(err)
	}
}
