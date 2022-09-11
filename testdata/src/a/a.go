package a

type hogeRepo struct{}

func (r hogeRepo) Rollback() {
}

func f() {
	// The pattern can be written in regular expression.
	repo := hogeRepo{}
	defer repo.Rollback()

	hogerepo := hogeRepo{} // want "this repository is not closed"
	print(hogerepo)
}
