package postgres

import "context"

func (p *Postgres) CleanAll(ctx context.Context) error {
	_, err := p.db.ExecContext(ctx, `
		TRUNCATE TABLE
			password_invites,
			user_sessions,
			users,
			brand_aliases,
			brands,
			exclusion_words,
			schedules,
			llms
	`)
	return err
}
